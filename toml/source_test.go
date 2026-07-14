package toml_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/toml"
)

type postgresConfig struct {
	User     string   `cf_toml:"postgres#user"`
	Password string   `cf_toml:"postgres#password"`
	Test1    int32    `cf_toml:"postgres#test1"`
	Test2    float32  `cf_toml:"postgres#test2"`
	Test3    []string `cf_toml:"postgres#test3"`
	Test4    int      `cf_toml:"postgres#test4"`
}

func TestTOMLSourceEndToEnd(t *testing.T) {
	cfg, err := gostructor.Configure(&postgresConfig{}, gostructor.WithSources(toml.File("../testdata/config.toml")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := postgresConfig{
		User:     "pelletier",
		Password: "mypassword",
		Test1:    1231,
		Test2:    43.52,
		Test3:    []string{"myTest1", "myTest2"},
		Test4:    123,
	}
	if cfg.User != want.User || cfg.Password != want.Password || cfg.Test1 != want.Test1 ||
		cfg.Test2 != want.Test2 || cfg.Test4 != want.Test4 {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
	if !reflect.DeepEqual(cfg.Test3, want.Test3) {
		t.Errorf("Test3 = %v, want %v", cfg.Test3, want.Test3)
	}
}

type nestedTableConfig struct {
	Value bool `cf_toml:"a.b.c#key"`
}

func TestTOMLSourceNestedTable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested.toml")
	if err := os.WriteFile(path, []byte("[a.b.c]\nkey = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := gostructor.Configure(&nestedTableConfig{}, gostructor.WithSources(toml.File(path)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Value {
		t.Error("Value = false, want true")
	}
}

func TestTOMLSourceMissingFileEnvVar(t *testing.T) {
	type cfgT struct {
		Value string `cf_toml:"x"`
	}
	os.Unsetenv(toml.FileEnvVar)
	_, err := gostructor.Configure(&cfgT{}, gostructor.WithSources(toml.New()))
	if err == nil {
		t.Fatal("expected an error when GOSTRUCTOR_TOML is unset and no explicit path was given")
	}
}
