// Package vault resolves gostructor fields from HashiCorp Vault secrets via
// github.com/hashicorp/vault/api. The client reads VAULT_ADDR and
// VAULT_TOKEN itself, the same variables the `vault` CLI uses.
package vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/goreflect/gostructor"
)

// DefaultPoll is the interval Watch re-checks referenced secrets for changes
// when WithPollInterval is not set. Vault (the OSS API) has no change-push, so
// live reload is poll-based.
const DefaultPoll = 30 * time.Second

// Name is this source's identity, used as the per-source override key in a cfg
// tag: `cfg:"password,vault:path/to/secret#key"`. Vault has no name-based
// default (a secret path cannot be derived from a field's base name), so Vault
// resolves a field only when it carries an explicit vault: override.
const Name = "vault"

// logicalReader is the slice of *vaultapi.Client's surface this source
// actually needs, so tests can substitute a fake instead of hitting a real
// Vault server.
type logicalReader interface {
	Read(path string) (*vaultapi.Secret, error)
}

type source struct {
	newClient func() (logicalReader, error)
	once      sync.Once
	logical   logicalReader
	initErr   error

	poll time.Duration
	log  *slog.Logger

	// mu guards watched, the set of secret paths seen during resolution. Watch
	// polls exactly these paths, so it re-checks only what the struct actually
	// reads.
	mu      sync.Mutex
	watched map[string]struct{}
}

// Option configures a Vault source.
type Option func(*source)

// WithPollInterval sets how often Watch re-checks the referenced secrets for
// changes. Zero or negative leaves the DefaultPoll.
func WithPollInterval(d time.Duration) Option {
	return func(s *source) {
		if d > 0 {
			s.poll = d
		}
	}
}

// WithLogger sets the logger for Watch diagnostics (poll errors). Nil is
// ignored; the default discards.
func WithLogger(l *slog.Logger) Option {
	return func(s *source) {
		if l != nil {
			s.log = l
		}
	}
}

// New resolves fields from Vault secrets, using a client configured from the
// environment (VAULT_ADDR, VAULT_TOKEN, and the rest of vaultapi.DefaultConfig).
// The returned source also implements gostructor.Watchable: it polls the secrets
// the struct references and signals a reload when any of them changes.
func New(opts ...Option) gostructor.Source {
	s := &source{
		newClient: func() (logicalReader, error) {
			client, err := vaultapi.NewClient(vaultapi.DefaultConfig())
			if err != nil {
				return nil, err
			}
			return client.Logical(), nil
		},
		poll: DefaultPoll,
		log:  slog.New(slog.DiscardHandler),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (*source) Name() string { return Name }

func (s *source) Resolve(field gostructor.FieldContext) (any, bool, error) {
	tagValue, ok := field.Override(Name)
	if !ok || tagValue == "" {
		return nil, false, nil
	}
	if err := s.init(); err != nil {
		return nil, false, err
	}
	path, key, err := parseTag(tagValue)
	if err != nil {
		return nil, false, err
	}
	s.recordPath(path)
	secret, err := s.logical.Read(path)
	if err != nil {
		return nil, false, fmt.Errorf("gostructor/vault: reading secret at %q: %w", path, err)
	}
	if secret == nil {
		return nil, false, fmt.Errorf("gostructor/vault: no secret found at path %q", path)
	}
	value, found := secret.Data[key]
	if !found {
		return nil, false, fmt.Errorf("gostructor/vault: key %q not found in secret at path %q", key, path)
	}
	return splitIfSlice(field, value), true, nil
}

// recordPath remembers a secret path resolution touched, so Watch polls only
// the secrets the struct actually references.
func (s *source) recordPath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.watched == nil {
		s.watched = make(map[string]struct{})
	}
	s.watched[path] = struct{}{}
}

// Watch polls the referenced secrets on the configured interval and signals a
// reload when any of them changes. Because Resolve reads Vault live on every
// call, a reload automatically picks up the new secret values — Watch only has
// to detect the change. It blocks until ctx is cancelled.
//
// Note that Watch can only poll secrets already seen during a fill, so call it
// after an initial Configure (which is exactly how gostructor.Watch drives it:
// initial fill first, then Watch).
func (s *source) Watch(ctx context.Context, onChange func()) error {
	log := s.log
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	poll := s.poll
	if poll <= 0 {
		poll = DefaultPoll
	}
	return gostructor.PollWatch(ctx, poll, s.fingerprint, onChange, log)
}

// fingerprint reads every watched secret and hashes their data, so any change
// to any referenced secret yields a different fingerprint. Reading the data
// directly (rather than KV-v2 metadata) works for both KV v1 and v2 mounts.
func (s *source) fingerprint(context.Context) (string, error) {
	if err := s.init(); err != nil {
		return "", err
	}
	s.mu.Lock()
	paths := make([]string, 0, len(s.watched))
	for p := range s.watched {
		paths = append(paths, p)
	}
	s.mu.Unlock()
	sort.Strings(paths)

	h := sha256.New()
	for _, path := range paths {
		secret, err := s.logical.Read(path)
		if err != nil {
			return "", fmt.Errorf("gostructor/vault: polling secret at %q: %w", path, err)
		}
		h.Write([]byte(path))
		h.Write([]byte{0})
		if secret != nil {
			// Marshal is deterministic for map[string]any (sorted keys), so the
			// hash is stable across reads of unchanged data.
			if payload, err := json.Marshal(secret.Data); err == nil {
				h.Write(payload)
			}
		}
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (s *source) init() error {
	s.once.Do(func() {
		logical, err := s.newClient()
		if err != nil {
			s.initErr = fmt.Errorf("gostructor/vault: creating client: %w", err)
			return
		}
		s.logical = logical
	})
	return s.initErr
}

// parseTag splits a vault override ("secret/path#key") into a Vault path and
// secret key, returning an error instead of panicking on a malformed value.
func parseTag(tagValue string) (path string, key string, err error) {
	parts := strings.SplitN(tagValue, "#", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("gostructor/vault: tag %q is malformed, expected 'path/to/secret#key'", tagValue)
	}
	return parts[0], parts[1], nil
}

// splitIfSlice mirrors the core module's env/default/ini sources: a slice
// destination field's secret value is a single string split on the field's
// separator (gos sep, default comma).
func splitIfSlice(field gostructor.FieldContext, raw any) any {
	kind := field.Type.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return raw
	}
	str, ok := raw.(string)
	if !ok {
		return raw
	}
	parts := strings.Split(str, field.Separator())
	result := make([]any, len(parts))
	for i, p := range parts {
		result[i] = strings.TrimSpace(p)
	}
	return result
}
