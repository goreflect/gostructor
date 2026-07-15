package gostructor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

type basicConfig struct {
	Host string `cfg:"host,env:GOSTRUCTOR_TEST_HOST" gos:"default:0.0.0.0"`
	Port int    `cfg:"port,env:GOSTRUCTOR_TEST_PORT" gos:"default:8080"`
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

// TestConfigureEnvNamingFromBase checks the env source derives its variable
// name from the base name in SCREAMING_SNAKE_CASE when there's no override.
func TestConfigureEnvNamingFromBase(t *testing.T) {
	type cfgT struct {
		MaxConns int `cfg:"maxConns" gos:"default:1"`
	}
	os.Setenv("MAX_CONNS", "42")
	defer os.Unsetenv("MAX_CONNS")

	cfg, err := Configure(&cfgT{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxConns != 42 {
		t.Errorf("MaxConns = %d, want 42 from MAX_CONNS", cfg.MaxConns)
	}
}

// TestConfigureAllNumericKinds guards against a regression where the
// unsized `uint` destination kind was routed through the same code path as
// uint32, producing a reflect.Value of Kind Uint32 instead of Uint - which
// panics at destination.Set() because the kinds don't match.
type numericConfig struct {
	I   int     `gos:"default:1"`
	I8  int8    `gos:"default:2"`
	I16 int16   `gos:"default:3"`
	I32 int32   `gos:"default:4"`
	I64 int64   `gos:"default:5"`
	U   uint    `gos:"default:6"`
	U8  uint8   `gos:"default:7"`
	U16 uint16  `gos:"default:8"`
	U32 uint32  `gos:"default:9"`
	U64 uint64  `gos:"default:10"`
	F32 float32 `gos:"default:1.5"`
	F64 float64 `gos:"default:2.5"`
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

func TestConfigureSlices(t *testing.T) {
	os.Setenv("GOSTRUCTOR_TEST_NAMES", "a,b,c")
	defer os.Unsetenv("GOSTRUCTOR_TEST_NAMES")

	type withEnvSlice struct {
		// Comma is the gos separator, so a comma-valued default needs a
		// non-comma sep, used for both the default and any source value.
		Flags []bool   `gos:"sep:|,default:true|false|true"`
		Names []string `cfg:"names,env:GOSTRUCTOR_TEST_NAMES"`
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
		Host string `gos:"default:localhost"`
	}
	Name string `gos:"default:svc"`
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
	APIKey string `cfg:"apiKey,env:GOSTRUCTOR_TEST_MISSING_KEY"`
}

func TestConfigureErrorsWhenConfiguredFieldUnresolved(t *testing.T) {
	os.Unsetenv("GOSTRUCTOR_TEST_MISSING_KEY")
	_, err := Configure(&requiredConfig{})
	if err == nil {
		t.Fatal("expected an error for a configured field with no source able to resolve it")
	}
}

// TestConfigureOptionalFieldStaysZero verifies gos:"optional" turns an
// otherwise-required unresolved field into a silent zero value.
func TestConfigureOptionalFieldStaysZero(t *testing.T) {
	type cfgT struct {
		APIKey string `cfg:"apiKey,env:GOSTRUCTOR_TEST_MISSING_OPTIONAL" gos:"optional"`
	}
	os.Unsetenv("GOSTRUCTOR_TEST_MISSING_OPTIONAL")
	cfg, err := Configure(&cfgT{})
	if err != nil {
		t.Fatalf("optional unresolved field should not error, got: %v", err)
	}
	if cfg.APIKey != "" {
		t.Errorf("APIKey = %q, want zero value", cfg.APIKey)
	}
}

type untaggedConfig struct {
	Internal string // no tags at all: gostructor has nothing to say about it
}

func TestConfigureSkipsUnconfiguredFields(t *testing.T) {
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

// TestConfigureCompositionPriority replaces the old cf_priority behavior:
// the SAME struct resolves differently based purely on the WithSources order.
func TestConfigureCompositionPriority(t *testing.T) {
	type cfgT struct {
		Value string `cfg:"value,env:GOSTRUCTOR_TEST_PRIORITY_VALUE" gos:"default:fallback"`
	}
	os.Setenv("GOSTRUCTOR_TEST_PRIORITY_VALUE", "from-env")
	defer os.Unsetenv("GOSTRUCTOR_TEST_PRIORITY_VALUE")

	// Default first: the env override never gets a turn.
	devCfg, err := Configure(&cfgT{}, WithSources(Default(), Env()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if devCfg.Value != "fallback" {
		t.Errorf("Value = %q, want %q (Default listed first wins)", devCfg.Value, "fallback")
	}

	// Env first: the operator override wins.
	prodCfg, err := Configure(&cfgT{}, WithSources(Env(), Default()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prodCfg.Value != "from-env" {
		t.Errorf("Value = %q, want %q (Env listed first wins)", prodCfg.Value, "from-env")
	}
}

func TestConfigureWithHookTransform(t *testing.T) {
	type cfgT struct {
		Name string `gos:"default:world"`
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
// string "80" from a gos default) instead of the field-typed value (int(80)) -
// which made a type assertion like value.(int) panic for any field not
// natively sourced as that Go type.
func TestConfigureWithHookReceivesTypedValue(t *testing.T) {
	type cfgT struct {
		Port int `gos:"default:80"`
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
		Name string `gos:"default:bad"`
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
		Name string `cfg:"name,env:GOSTRUCTOR_TEST_EXPLICIT" gos:"default:unused"`
	}
	os.Setenv("GOSTRUCTOR_TEST_EXPLICIT", "from-env")
	defer os.Unsetenv("GOSTRUCTOR_TEST_EXPLICIT")

	// Only Default in the source list: the env source should be absent entirely.
	cfg, err := Configure(&cfgT{}, WithSources(Default()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "unused" {
		t.Errorf("Name = %q, want the default value since env wasn't in WithSources", cfg.Name)
	}
}

type jsonConfig struct {
	Host string `cfg:"host,json:server.host"`
	Tags []int  `cfg:"tags,json:server.tags"`
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
	Test  string   `cfg:"test,ini:TEST#test"`
	Test2 int32    `cfg:"test2,ini:TEST#test2"`
	Test3 float32  `cfg:"test3,ini:TEST#test3"`
	Test4 []string `cfg:"test4,ini:TEST#test4"`
	Test5 int16    `cfg:"test5,ini:TEST#test5"`
}

func TestConfigureINISourceEndToEnd(t *testing.T) {
	cfg, err := Configure(&iniConfig{}, WithSources(INIFile("testdata/config.ini")))
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
		Name string `cfg:"name"`
	}
	_, err := Configure(&cfgT{}, WithSources(panicSource{}))
	if err == nil {
		t.Fatal("expected Configure to convert a source panic into an error")
	}
}

type panicSource struct{}

func (panicSource) Name() string { return "boom" }
func (panicSource) Resolve(FieldContext) (any, bool, error) {
	panic("boom")
}

func TestConfigureUnresolvedFieldMatchesSentinel(t *testing.T) {
	_, err := Configure(&requiredConfig{})
	if !errors.Is(err, ErrFieldNotResolved) {
		t.Fatalf("err = %v, want it to wrap ErrFieldNotResolved", err)
	}
	var nre *NotResolvedError
	if !errors.As(err, &nre) {
		t.Fatalf("err = %v, want a *NotResolvedError", err)
	}
	if len(nre.Sources) == 0 {
		t.Errorf("NotResolvedError.Sources is empty; want the sources that were tried")
	}
}

func TestConfigureInvalidTargetSentinel(t *testing.T) {
	if _, err := Configure[basicConfig](nil); !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("nil target: err = %v, want ErrInvalidTarget", err)
	}
}

func TestConfigureSourceError(t *testing.T) {
	type cfgT struct {
		Host string `cfg:"host,json:server.host"`
	}
	_, err := Configure(&cfgT{}, WithSources(JSONFile("testdata/does-not-exist.json")))
	var se *SourceError
	if !errors.As(err, &se) {
		t.Fatalf("err = %v, want a *SourceError", err)
	}
	if se.Source != SourceJSON || se.Field != "Host" {
		t.Errorf("SourceError = %+v, want Field=Host Source=json", se)
	}
}

func TestConfigureHookError(t *testing.T) {
	type cfgT struct {
		Name string `gos:"default:x"`
	}
	sentinel := errors.New("rejected")
	_, err := Configure(&cfgT{}, WithHook(func(_ FieldContext, _ any) (any, error) {
		return nil, sentinel
	}))
	var he *HookError
	if !errors.As(err, &he) {
		t.Fatalf("err = %v, want a *HookError", err)
	}
	if he.Field != "Name" {
		t.Errorf("HookError.Field = %q, want Name", he.Field)
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected the hook's own error to be reachable via errors.Is")
	}
}

func TestConfigureErrorsSatisfyFieldError(t *testing.T) {
	type cfgT struct {
		Port int `gos:"default:nope"`
	}
	_, err := Configure(&cfgT{})
	var fe FieldError
	if !errors.As(err, &fe) {
		t.Fatalf("err = %v, want it to satisfy FieldError", err)
	}
	if fe.FieldName() != "Port" {
		t.Errorf("FieldName() = %q, want Port", fe.FieldName())
	}
}

func TestConfigureConvertErrorCarriesContext(t *testing.T) {
	type cfgT struct {
		Port int `gos:"default:not-a-number"`
	}
	_, err := Configure(&cfgT{})
	var ce *ConvertError
	if !errors.As(err, &ce) {
		t.Fatalf("err = %v, want a *ConvertError", err)
	}
	if ce.Field != "Port" {
		t.Errorf("ConvertError.Field = %q, want Port", ce.Field)
	}
	var numErr *strconv.NumError
	if !errors.As(err, &numErr) {
		t.Errorf("expected the underlying strconv error to be reachable, got %v", err)
	}
}

func TestConfigureDurationAndNamedType(t *testing.T) {
	type level int
	type cfgT struct {
		Timeout time.Duration `gos:"default:1h30m"`
		Level   level         `gos:"default:5"`
	}
	cfg, err := Configure(&cfgT{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Timeout != 90*time.Minute {
		t.Errorf("Timeout = %v, want 1h30m", cfg.Timeout)
	}
	if cfg.Level != 5 {
		t.Errorf("Level = %v, want 5", cfg.Level)
	}
}

func TestConfigureTextUnmarshalerField(t *testing.T) {
	type cfgT struct {
		Created time.Time `gos:"default:2020-01-02T03:04:05Z"`
	}
	cfg, err := Configure(&cfgT{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	if !cfg.Created.Equal(want) {
		t.Errorf("Created = %v, want %v", cfg.Created, want)
	}
}

func TestConfigurePointerField(t *testing.T) {
	type cfgT struct {
		Port *int `gos:"default:8080"`
	}
	cfg, err := Configure(&cfgT{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port == nil || *cfg.Port != 8080 {
		t.Errorf("Port = %v, want *8080", cfg.Port)
	}
}

func TestConfigureRejectsOverflow(t *testing.T) {
	type cfgT struct {
		Small int8 `gos:"default:300"`
	}
	_, err := Configure(&cfgT{})
	if err == nil {
		t.Fatal("expected an overflow error converting 300 into int8")
	}
}

func TestConfigureJSONExactIntAndFractionalError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.json")
	if err := os.WriteFile(path, []byte(`{"id":9007199254740993}`), 0o644); err != nil {
		t.Fatal(err)
	}
	type cfgT struct {
		ID int64 `cfg:"id,json:id"`
	}
	cfg, err := Configure(&cfgT{}, WithSources(JSONFile(path)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ID != 9007199254740993 {
		t.Errorf("ID = %d, want 9007199254740993 (exact, no float rounding)", cfg.ID)
	}

	fracPath := filepath.Join(dir, "frac.json")
	if err := os.WriteFile(fracPath, []byte(`{"id":3.9}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Configure(&cfgT{}, WithSources(JSONFile(fracPath))); err == nil {
		t.Fatal("expected an error converting 3.9 into int64")
	}
}

func TestConfigureJSONSliceOfStructs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")
	if err := os.WriteFile(path, []byte(`{"servers":[{"host":"a","port":1},{"host":"b","port":2}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	type server struct {
		Host string
		Port int
	}
	type cfgT struct {
		Servers []server `cfg:"servers,json:servers"`
	}
	cfg, err := Configure(&cfgT{}, WithSources(JSONFile(path)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 2 || cfg.Servers[0].Host != "a" || cfg.Servers[1].Port != 2 {
		t.Errorf("Servers = %+v, want [{a 1} {b 2}]", cfg.Servers)
	}
}
