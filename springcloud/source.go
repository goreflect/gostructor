// Package springcloud resolves gostructor fields from a Spring Cloud Config
// Server. It calls GET /{application}/{profile}[/{label}], merges the returned
// propertySources by their precedence (earlier sources win, as Spring defines),
// and resolves each field by its flat, dotted key (`cfg:"server.port"` or a
// `springcloud:` override).
//
// Spring Cloud Config has no change-push, so it implements gostructor.Watchable
// by polling: it re-fetches on an interval and signals a reload when the
// server's version or response body changes. An optional gostructor/snapshot
// store serves the last-known-good properties when the server is unreachable at
// startup.
package springcloud

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/snapshot"
)

// DefaultName is this source's identity and cfg per-source override key.
const DefaultName = "springcloud"

// DefaultPoll is the re-fetch interval used when Options.Poll is zero.
const DefaultPoll = 30 * time.Second

// Options configures a Spring Cloud Config Source.
type Options struct {
	// Address is the config server base URL, e.g. "http://localhost:8888".
	// Required.
	Address string
	// Application is the {application} path segment (the spring.application.name
	// of the client). Required.
	Application string
	// Profile is the {profile} segment; empty defaults to "default".
	Profile string
	// Label is the optional {label} segment (a git branch/tag on the server).
	Label string
	// Poll is the re-fetch interval for Watch; zero uses DefaultPoll, negative
	// disables polling.
	Poll time.Duration
	// Username and Password enable HTTP basic auth if set.
	Username string
	Password string
	// HTTPClient overrides the default client (timeouts, TLS, proxies).
	HTTPClient *http.Client
	// Snapshot, if set, stores each good fetch and is served when the server is
	// unreachable at startup.
	Snapshot snapshot.Store
	// Name overrides the source identity (DefaultName otherwise).
	Name string
	// Logger receives poll/fallback diagnostics; nil discards them.
	Logger *slog.Logger
}

// environment mirrors the JSON a Spring Cloud Config Server returns.
type environment struct {
	Name            string           `json:"name"`
	Profiles        []string         `json:"profiles"`
	Label           string           `json:"label"`
	Version         string           `json:"version"`
	PropertySources []propertySource `json:"propertySources"`
}

type propertySource struct {
	Name   string         `json:"name"`
	Source map[string]any `json:"source"`
}

type source struct {
	opts    Options
	name    string
	profile string
	client  *http.Client
	log     *slog.Logger

	mu          sync.Mutex
	data        map[string]any
	fingerprint string
	loaded      bool
	loadErr     error
}

// New returns a Spring-Cloud-Config-backed gostructor.Source (also a
// gostructor.Watchable). The first fetch happens lazily on first Resolve.
func New(opts Options) (*source, error) {
	if opts.Address == "" {
		return nil, fmt.Errorf("gostructor/springcloud: Address is required")
	}
	if opts.Application == "" {
		return nil, fmt.Errorf("gostructor/springcloud: Application is required")
	}
	s := &source{
		opts:    opts,
		name:    opts.Name,
		profile: opts.Profile,
		client:  opts.HTTPClient,
		log:     opts.Logger,
	}
	if s.name == "" {
		s.name = DefaultName
	}
	if s.profile == "" {
		s.profile = "default"
	}
	if s.client == nil {
		s.client = &http.Client{Timeout: 10 * time.Second}
	}
	if s.log == nil {
		s.log = slog.New(slog.DiscardHandler)
	}
	return s, nil
}

func (s *source) Name() string { return s.name }

// Resolve fills a field from the merged property set, fetching it on first use.
func (s *source) Resolve(field gostructor.FieldContext) (any, bool, error) {
	if err := s.ensureLoaded(); err != nil {
		return nil, false, err
	}
	key := field.SourceKey(s.name, gostructor.Identity)
	if key == "" {
		return nil, false, nil
	}
	s.mu.Lock()
	value, ok := s.data[key]
	s.mu.Unlock()
	if !ok || value == nil {
		return nil, false, nil
	}
	return normalize(field, value), true, nil
}

// Version returns the config server version (backing commit) of the currently
// served properties, or the empty string before the first fetch.
func (s *source) Version() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.fingerprint
}

// Watch re-fetches on the poll interval and fires onChange when the server's
// version or response changes. It blocks until ctx is cancelled.
func (s *source) Watch(ctx context.Context, onChange func()) error {
	if err := s.ensureLoaded(); err != nil {
		s.log.Warn("gostructor/springcloud: initial load failed, watch will keep retrying", "err", err)
	}
	poll := s.opts.Poll
	if poll < 0 {
		<-ctx.Done()
		return ctx.Err()
	}
	if poll == 0 {
		poll = DefaultPoll
	}
	ticker := time.NewTicker(poll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			changed, err := s.fetchAndApply(ctx)
			if err != nil {
				s.log.Warn("gostructor/springcloud: poll failed, keeping current config", "err", err)
				continue
			}
			if changed {
				onChange()
			}
		}
	}
}

func (s *source) ensureLoaded() error {
	s.mu.Lock()
	if s.loaded {
		defer s.mu.Unlock()
		return s.loadErr
	}
	s.loaded = true
	s.mu.Unlock()

	if _, err := s.fetchAndApply(context.Background()); err != nil {
		if s.loadFromSnapshot(err) {
			return nil
		}
		s.mu.Lock()
		s.loadErr = err
		s.mu.Unlock()
		return err
	}
	return nil
}

// fetchAndApply fetches the environment, merges its property sources, and — when
// the fingerprint changed — updates the served data and saves a snapshot. It
// reports whether the served config changed.
func (s *source) fetchAndApply(ctx context.Context) (changed bool, err error) {
	env, raw, err := s.fetch(ctx)
	if err != nil {
		return false, err
	}
	fingerprint := env.Version
	if fingerprint == "" {
		sum := sha256.Sum256(raw)
		fingerprint = hex.EncodeToString(sum[:])
	}
	s.mu.Lock()
	if fingerprint == s.fingerprint && s.data != nil {
		s.mu.Unlock()
		return false, nil
	}
	s.data = merge(env.PropertySources)
	s.fingerprint = fingerprint
	s.mu.Unlock()

	if s.opts.Snapshot != nil {
		if err := s.opts.Snapshot.Save(fingerprint, raw); err != nil {
			s.log.Warn("gostructor/springcloud: saving snapshot failed", "err", err)
		}
	}
	return true, nil
}

func (s *source) fetch(ctx context.Context) (*environment, []byte, error) {
	endpoint := s.opts.Address + "/" + url.PathEscape(s.opts.Application) + "/" + url.PathEscape(s.profile)
	if s.opts.Label != "" {
		endpoint += "/" + url.PathEscape(s.opts.Label)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/json")
	if s.opts.Username != "" || s.opts.Password != "" {
		req.SetBasicAuth(s.opts.Username, s.opts.Password)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("gostructor/springcloud: requesting %s: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, nil, fmt.Errorf("gostructor/springcloud: reading response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("gostructor/springcloud: %s returned %s", endpoint, resp.Status)
	}
	var env environment
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, nil, fmt.Errorf("gostructor/springcloud: decoding response: %w", err)
	}
	return &env, body, nil
}

func (s *source) loadFromSnapshot(cause error) bool {
	if s.opts.Snapshot == nil {
		return false
	}
	version, raw, err := s.opts.Snapshot.Load()
	if err != nil {
		return false
	}
	var env environment
	if err := json.Unmarshal(raw, &env); err != nil {
		s.log.Warn("gostructor/springcloud: stored snapshot is undecodable", "err", err)
		return false
	}
	s.mu.Lock()
	s.data = merge(env.PropertySources)
	s.fingerprint = version
	s.mu.Unlock()
	s.log.Warn("gostructor/springcloud: config server unavailable, serving last-known-good snapshot",
		"cause", cause, "version", version)
	return true
}

// merge flattens the property sources into one map, honouring Spring's
// precedence: sources earlier in the list win, so a later source only fills keys
// not already set.
func merge(sources []propertySource) map[string]any {
	merged := map[string]any{}
	for _, ps := range sources {
		for k, v := range ps.Source {
			if _, exists := merged[k]; !exists {
				merged[k] = v
			}
		}
	}
	return merged
}

// normalize adapts a JSON-typed property value to what the field expects: a
// scalar passes through, and a string bound to a slice field is split on the
// field separator.
func normalize(field gostructor.FieldContext, value any) any {
	kind := field.Type.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return value
	}
	str, ok := value.(string)
	if !ok {
		return value // already a JSON array
	}
	parts := strings.Split(str, field.Separator())
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
