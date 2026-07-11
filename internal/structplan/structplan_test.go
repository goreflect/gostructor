package structplan

import (
	"reflect"
	"testing"
)

type inner struct {
	Host string `cf_env:"HOST"`
}

type outer struct {
	Name   string `cf_env:"NAME"`
	Inner  inner
	hidden string //nolint:unused // exercises that unexported fields are skipped
}

func TestForFlattensNestedStructsAndSkipsUnexported(t *testing.T) {
	plan := For(reflect.TypeOf(outer{}))
	if len(plan.Fields) != 2 {
		t.Fatalf("got %d fields, want 2 (Name, Inner.Host), fields: %+v", len(plan.Fields), plan.Fields)
	}
	names := map[string]bool{}
	for _, f := range plan.Fields {
		names[f.Struct.Name] = true
	}
	if !names["Name"] || !names["Host"] {
		t.Errorf("got fields %v, want Name and Host", names)
	}
}

func TestForCachesByType(t *testing.T) {
	first := For(reflect.TypeOf(outer{}))
	second := For(reflect.TypeOf(outer{}))
	if first != second {
		t.Error("expected For to return the same cached *Plan for the same type")
	}
}
