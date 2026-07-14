// Package vault resolves gostructor fields from HashiCorp Vault secrets via
// github.com/hashicorp/vault/api. The client reads VAULT_ADDR and
// VAULT_TOKEN itself, the same variables the `vault` CLI uses.
package vault

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/goreflect/gostructor"
)

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
}

// New resolves fields from Vault secrets, using a client configured from the
// environment (VAULT_ADDR, VAULT_TOKEN, and the rest of vaultapi.DefaultConfig).
func New() gostructor.Source {
	return &source{newClient: func() (logicalReader, error) {
		client, err := vaultapi.NewClient(vaultapi.DefaultConfig())
		if err != nil {
			return nil, err
		}
		return client.Logical(), nil
	}}
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
