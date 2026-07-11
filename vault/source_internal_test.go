package vault

import (
	"errors"
	"reflect"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/goreflect/gostructor"
)

func TestParseTagWellFormed(t *testing.T) {
	path, key, err := parseTag("secret/service/stage#my-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "secret/service/stage" || key != "my-key" {
		t.Errorf("got path=%q key=%q", path, key)
	}
}

func TestParseTagMalformed(t *testing.T) {
	cases := []string{"", "no-hash-here", "path#", "#key"}
	for _, tagValue := range cases {
		t.Run(tagValue, func(t *testing.T) {
			if _, _, err := parseTag(tagValue); err == nil {
				t.Errorf("expected an error for malformed tag %q", tagValue)
			}
		})
	}
}

func TestParseTagExtraHashGoesIntoKey(t *testing.T) {
	path, key, err := parseTag("a#b#c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "a" || key != "b#c" {
		t.Errorf("got path=%q key=%q", path, key)
	}
}

type fakeLogical struct {
	secrets map[string]*vaultapi.Secret
	err     error
}

func (f fakeLogical) Read(path string) (*vaultapi.Secret, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.secrets[path], nil
}

func newTestSource(l logicalReader) *source {
	s := &source{newClient: func() (logicalReader, error) { return l, nil }}
	return s
}

func fieldWithVaultTag(tagValue string, fieldType reflect.Type) gostructor.FieldContext {
	return gostructor.FieldContext{StructField: reflect.StructField{
		Name: "Value",
		Tag:  reflect.StructTag(`cf_vault:"` + tagValue + `"`),
		Type: fieldType,
	}}
}

func TestSourceResolveBaseType(t *testing.T) {
	s := newTestSource(fakeLogical{secrets: map[string]*vaultapi.Secret{
		"my-service/stage": {Data: map[string]any{"api-key": "s3cr3t"}},
	}})
	field := fieldWithVaultTag("my-service/stage#api-key", reflect.TypeOf(""))

	value, found, err := s.Resolve(field)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found || value != "s3cr3t" {
		t.Errorf("got value=%v found=%v", value, found)
	}
}

func TestSourceResolveSliceType(t *testing.T) {
	s := newTestSource(fakeLogical{secrets: map[string]*vaultapi.Secret{
		"my-service/stage": {Data: map[string]any{"allowlist": "10.0.0.1, 10.0.0.2"}},
	}})
	field := fieldWithVaultTag("my-service/stage#allowlist", reflect.TypeOf([]string{}))

	value, found, err := s.Resolve(field)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	got, ok := value.([]any)
	if !ok || len(got) != 2 || got[0] != "10.0.0.1" || got[1] != "10.0.0.2" {
		t.Errorf("got %v", value)
	}
}

func TestSourceResolveNoTag(t *testing.T) {
	s := newTestSource(fakeLogical{})
	field := gostructor.FieldContext{StructField: reflect.StructField{Name: "Value", Type: reflect.TypeOf("")}}
	_, found, err := s.Resolve(field)
	if err != nil || found {
		t.Errorf("expected found=false, no error for an untagged field; got found=%v err=%v", found, err)
	}
}

func TestSourceResolveMalformedTag(t *testing.T) {
	s := newTestSource(fakeLogical{})
	field := fieldWithVaultTag("no-hash", reflect.TypeOf(""))
	_, _, err := s.Resolve(field)
	if err == nil {
		t.Fatal("expected an error for a malformed cf_vault tag")
	}
}

func TestSourceResolveSecretNotFound(t *testing.T) {
	s := newTestSource(fakeLogical{secrets: map[string]*vaultapi.Secret{}})
	field := fieldWithVaultTag("missing/path#key", reflect.TypeOf(""))
	_, _, err := s.Resolve(field)
	if err == nil {
		t.Fatal("expected an error when the secret path doesn't exist")
	}
}

func TestSourceResolveKeyNotFound(t *testing.T) {
	s := newTestSource(fakeLogical{secrets: map[string]*vaultapi.Secret{
		"path": {Data: map[string]any{"other-key": "value"}},
	}})
	field := fieldWithVaultTag("path#missing-key", reflect.TypeOf(""))
	_, _, err := s.Resolve(field)
	if err == nil {
		t.Fatal("expected an error when the key isn't present in the secret")
	}
}

func TestSourceResolveReadError(t *testing.T) {
	s := newTestSource(fakeLogical{err: errors.New("connection refused")})
	field := fieldWithVaultTag("path#key", reflect.TypeOf(""))
	_, _, err := s.Resolve(field)
	if err == nil {
		t.Fatal("expected the underlying read error to propagate")
	}
}

type secretConfig struct {
	APIKey    string `cf_vault:"my-service/stage#api-key"`
	RateLimit int16  `cf_vault:"my-service/stage#rate"`
}

func TestConfigureEndToEndWithFakeVault(t *testing.T) {
	s := newTestSource(fakeLogical{secrets: map[string]*vaultapi.Secret{
		"my-service/stage": {Data: map[string]any{"api-key": "s3cr3t", "rate": "42"}},
	}})
	cfg, err := gostructor.Configure(&secretConfig{}, gostructor.WithSources(s))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "s3cr3t" || cfg.RateLimit != 42 {
		t.Errorf("got %+v", cfg)
	}
}

func TestSourceInitErrorIsCachedAndReturned(t *testing.T) {
	s := &source{newClient: func() (logicalReader, error) { return nil, errors.New("boom") }}
	field := fieldWithVaultTag("path#key", reflect.TypeOf(""))
	_, _, err1 := s.Resolve(field)
	_, _, err2 := s.Resolve(field)
	if err1 == nil || err2 == nil {
		t.Fatal("expected init error on both calls")
	}
}
