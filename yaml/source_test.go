package yaml_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/yaml"
)

type fixtureConfig struct {
	// Test..Test3 use a bare base name: the yaml source's Identity naming maps
	// "test" -> the top-level "test" key. Nested values use an explicit
	// yaml: override with a dotted path.
	Test  int      `cfg:"test"`
	Test2 []string `cfg:"test2"`
	Test3 []int    `cfg:"test3"`
	Test4 string   `cfg:"test4,yaml:test5.test4"`
	Test6 []int    `cfg:"test6,yaml:test5.test6"`
}

func TestYAMLSourceEndToEndFixture(t *testing.T) {
	cfg, err := gostructor.Configure(&fixtureConfig{}, gostructor.WithSources(yaml.File("../testdata/config.yml")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Test != 1 {
		t.Errorf("Test = %d, want 1", cfg.Test)
	}
	if !reflect.DeepEqual(cfg.Test2, []string{"string", "string2", "string3"}) {
		t.Errorf("Test2 = %v", cfg.Test2)
	}
	if !reflect.DeepEqual(cfg.Test3, []int{1, 2, 3}) {
		t.Errorf("Test3 = %v", cfg.Test3)
	}
	if cfg.Test4 != "str1" {
		t.Errorf("Test4 = %q, want str1", cfg.Test4)
	}
	if !reflect.DeepEqual(cfg.Test6, []int{1231, 15123}) {
		t.Errorf("Test6 = %v", cfg.Test6)
	}
}

type mapConfig struct {
	Nested map[string]string `cfg:"nested,yaml:test5"`
}

func TestYAMLSourceMapDestination(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "map.yml")
	if err := os.WriteFile(path, []byte("test5:\n  a: one\n  b: two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := gostructor.Configure(&mapConfig{}, gostructor.WithSources(yaml.File(path)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]string{"a": "one", "b": "two"}
	if !reflect.DeepEqual(cfg.Nested, want) {
		t.Errorf("Nested = %v, want %v", cfg.Nested, want)
	}
}

func TestYAMLSourceMissingFileEnvVar(t *testing.T) {
	type cfgT struct {
		Value string `cfg:"value,yaml:x"`
	}
	os.Unsetenv(yaml.FileEnvVar)
	_, err := gostructor.Configure(&cfgT{}, gostructor.WithSources(yaml.New()))
	if err == nil {
		t.Fatal("expected an error when GOSTRUCTOR_YAML is unset and no explicit path was given")
	}
}
