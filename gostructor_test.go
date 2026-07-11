package gostructor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type basicConfig struct {
	Host string `cf_env:"GOSTRUCTOR_TEST_HOST" cf_default:"0.0.0.0"`
	Port int    `cf_env:"GOSTRUCTOR_TEST_PORT" cf_default:"8080"`
}

func TestConfigureEnvOverridesDefault(t *testing.T) {
	os.Setenv("GOSTRUCTOR_TEST_HOST", "10.0.0.1")
	defer os.Unsetenv("GOSTRUCTOR_TEST_HOST")

	cfg, err := Configure(&basicConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "10.0.0.1" {
		t.Errorf("Host = %q, want env value", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want default 8080", cfg.Port)
	}
}

func TestConfigureFallsBackToDefaultWhenEnvUnset(t *testing.T) {
	os.Unsetenv("GOSTRUCTOR_TEST_HOST")
	os.Unsetenv("GOSTRUCTOR_TEST_PORT")

	cfg, err := Configure(&basicConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "0.0.0.0" || cfg.Port != 8080 {
		t.Errorf("got %+v, want defaults", cfg)
	}
}

// TestConfigureAllNumericKinds guards against a regression where the
// unsized `uint` destination kind was routed through the same code path as
// uint32, producing a reflect.Value of Kind Uint32 instead of Uint - which
// panics at destination.Set() because the kinds don't match.
type numericConfig struct {
	I   int     `cf_default:"1"`
	I8  int8    `cf_default:"2"`
	I16 int16   `cf_default:"3"`
	I32 int32   `cf_default:"4"`
	I64 int64   `cf_default:"5"`
	U   uint    `cf_default:"6"`
	U8  uint8   `cf_default:"7"`
	U16 uint16  `cf_default:"8"`
	U32 uint32  `cf_default:"9"`
	U64 uint64  `cf_default:"10"`
	F32 float32 `cf_default:"1.5"`
	F64 float64 `cf_default:"2.5"`
}

func TestConfigureAllNumericKinds(t *testing.T) {
	cfg, err := Configure(&numericConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.I != 1 || cfg.I8 != 2 || cfg.I16 != 3 || cfg.I32 != 4 || cfg.I64 != 5 {
		t.Errorf("signed ints: %+v", cfg)
	}
	if cfg.U != 6 || cfg.U8 != 7 || cfg.U16 != 8 || cfg.U32 != 9 || cfg.U64 != 10 {
		t.Errorf("unsigned ints: %+v", cfg)
	}
	if cfg.F32 != 1.5 || cfg.F64 != 2.5 {
		t.Errorf("floats: %+v", cfg)
	}
}

type sliceConfig struct {
	Flags []bool `cf_default:"true,false,true"`
	Names []string
}

func TestConfigureSlices(t *testing.T) {
	os.Setenv("GOSTRUCTOR_TEST_NAMES", "a,b,c")
	defer os.Unsetenv("GOSTRUCTOR_TEST_NAMES")

	type withEnvSlice struct {
		Flags []bool   `cf_default:"true,false,true"`
		Names []string `cf_env:"GOSTRUCTOR_TEST_NAMES"`
	}
	cfg, err := Configure(&withEnvSlice{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantFlags := []bool{true, false, true}
	if len(cfg.Flags) != len(wantFlags) {
		t.Fatalf("Flags = %v, want %v", cfg.Flags, wantFlags)
	}
	for i := range wantFlags {
		if cfg.Flags[i] != wantFlags[i] {
			t.Errorf("Flags[%d] = %v, want %v", i, cfg.Flags[i], wantFlags[i])
		}
	}
	wantNames := []string{"a", "b", "c"}
	if len(cfg.Names) != len(wantNames) {
		t.Fatalf("Names = %v, want %v", cfg.Names, wantNames)
	}
}

type nestedConfig struct {
	Server struct {
		Host string `cf_default:"localhost"`
	}
	Name string `cf_default:"svc"`
}

func TestConfigureNestedStruct(t *testing.T) {
	cfg, err := Configure(&nestedConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("Server.Host = %q, want localhost", cfg.Server.Host)
	}
	if cfg.Name != "svc" {
		t.Errorf("Name = %q, want svc", cfg.Name)
	}
}

type requiredConfig struct {
	APIKey string `cf_env:"GOSTRUCTOR_TEST_MISSING_KEY"`
}

func TestConfigureErrorsWhenTaggedFieldUnresolved(t *testing.T) {
	os.Unsetenv("GOSTRUCTOR_TEST_MISSING_KEY")
	_, err := Configure(&requiredConfig{})
	if err == nil {
		t.Fatal("expected an error for a tagged field with no source able to resolve it")
	}
}

type untaggedConfig struct {
	Internal string // no tags at all: gostructor has nothing to say about it
}

func TestConfigureSkipsUntaggedFields(t *testing.T) {
	cfg, err := Configure(&untaggedConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Internal != "" {
		t.Errorf("Internal = %q, want zero value", cfg.Internal)
	}
}

func TestConfigureNilTarget(t *testing.T) {
	_, err := Configure[basicConfig](nil)
	if err == nil {
		t.Fatal("expected an error for a nil target")
	}
}

type priorityConfig struct {
	Value string `cf_env:"GOSTRUCTOR_TEST_PRIORITY_VALUE" cf_default:"fallback" cf_priority:"prod:cf_env,cf_default;dev:cf_default,cf_env"`
}

func TestConfigurePriorityTag(t *testing.T) {
	os.Setenv("GOSTRUCTOR_TEST_PRIORITY_VALUE", "from-env")
	defer os.Unsetenv("GOSTRUCTOR_TEST_PRIORITY_VALUE")

	os.Setenv(PriorityEnvVar, "dev")
	defer os.Unsetenv(PriorityEnvVar)

	cfg, err := Configure(&priorityConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Value != "fallback" {
		t.Errorf("Value = %q, want %q (dev stage prefers cf_default)", cfg.Value, "fallback")
	}

	os.Setenv(PriorityEnvVar, "prod")
	cfg2, err := Configure(&priorityConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg2.Value != "from-env" {
		t.Errorf("Value = %q, want %q (prod stage prefers cf_env)", cfg2.Value, "from-env")
	}
}

func TestConfigureWithHookTransform(t *testing.T) {
	type cfgT struct {
		Name string `cf_default:"world"`
	}
	cfg, err := Configure(&cfgT{}, WithHook(func(_ FieldContext, value any) (any, error) {
		if s, ok := value.(string); ok {
			return "hello-" + s, nil
		}
		return value, nil
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "hello-world" {
		t.Errorf("Name = %q, want hello-world", cfg.Name)
	}
}

// TestConfigureWithHookReceivesTypedValue guards against a regression where
// hooks ran on the raw, pre-conversion value a source produced (e.g. the
// string "80" from a cf_default tag) instead of the field-typed value
// (int(80)) - which made a type assertion like value.(int) panic for any
// field not natively sourced as that Go type.
func TestConfigureWithHookReceivesTypedValue(t *testing.T) {
	type cfgT struct {
		Port int `cf_default:"80"`
	}
	var sawType string
	_, err := Configure(&cfgT{}, WithHook(func(_ FieldContext, value any) (any, error) {
		sawType = fmt.Sprintf("%T", value)
		if _, ok := value.(int); !ok {
			return nil, fmt.Errorf("hook received %T, want int", value)
		}
		return value, nil
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sawType != "int" {
		t.Errorf("hook saw type %s, want int", sawType)
	}
}

func TestConfigureWithHookValidationError(t *testing.T) {
	type cfgT struct {
		Name string `cf_default:"bad"`
	}
	_, err := Configure(&cfgT{}, WithHook(func(_ FieldContext, value any) (any, error) {
		return nil, errors.New("rejected")
	}))
	if err == nil {
		t.Fatal("expected the hook's error to abort Configure")
	}
}

func TestConfigureWithExplicitSources(t *testing.T) {
	type cfgT struct {
		Name string `cf_default:"unused" cf_env:"GOSTRUCTOR_TEST_EXPLICIT"`
	}
	os.Setenv("GOSTRUCTOR_TEST_EXPLICIT", "from-env")
	defer os.Unsetenv("GOSTRUCTOR_TEST_EXPLICIT")

	// Only Default in the source list: cf_env should be ignored entirely.
	cfg, err := Configure(&cfgT{}, WithSources(Default()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "unused" {
		t.Errorf("Name = %q, want the default value since env wasn't in WithSources", cfg.Name)
	}
}

type jsonConfig struct {
	Host string `cf_json:"server.host"`
	Tags []int  `cf_json:"server.tags"`
}

func TestConfigureJSONSourceEndToEnd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"server":{"host":"0.0.0.0","tags":[1,2,3]}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Configure(&jsonConfig{}, WithSources(JSONFile(path)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("Host = %q, want 0.0.0.0", cfg.Host)
	}
	wantTags := []int{1, 2, 3}
	if len(cfg.Tags) != len(wantTags) {
		t.Fatalf("Tags = %v, want %v", cfg.Tags, wantTags)
	}
}

type iniConfig struct {
	Test  string   `cf_ini:"TEST#test"`
	Test2 int32    `cf_ini:"TEST#test2"`
	Test3 float32  `cf_ini:"TEST#test3"`
	Test4 []string `cf_ini:"TEST#test4"`
	Test5 int16    `cf_ini:"TEST#test5"`
}

func TestConfigureINISourceEndToEnd(t *testing.T) {
	cfg, err := Configure(&iniConfig{}, WithSources(INIFile("test_configs/config.ini")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := iniConfig{
		Test:  "tururu",
		Test2: 614,
		Test3: 86.27,
		Test4: []string{"str1", "str2", "str3"},
		Test5: 15,
	}
	if cfg.Test != want.Test || cfg.Test2 != want.Test2 || cfg.Test3 != want.Test3 || cfg.Test5 != want.Test5 {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
	if len(cfg.Test4) != len(want.Test4) {
		t.Fatalf("Test4 = %v, want %v", cfg.Test4, want.Test4)
	}
	for i := range want.Test4 {
		if cfg.Test4[i] != want.Test4[i] {
			t.Errorf("Test4[%d] = %q, want %q", i, cfg.Test4[i], want.Test4[i])
		}
	}
}

func TestConfigureRecoversFromPanicInSource(t *testing.T) {
	type cfgT struct {
		Name string `cf_boom:"x"`
	}
	_, err := Configure(&cfgT{}, WithSources(panicSource{}))
	if err == nil {
		t.Fatal("expected Configure to convert a source panic into an error")
	}
}

type panicSource struct{}

func (panicSource) Tag() string { return "cf_boom" }
func (panicSource) Resolve(FieldContext) (any, bool, error) {
	panic("boom")
}
