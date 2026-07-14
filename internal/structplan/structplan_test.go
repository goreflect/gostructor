package structplan

import (
	"reflect"
	"testing"
	"time"
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

type withAtomicStructs struct {
	Created time.Time `cf_json:"created"` // TextUnmarshaler: one leaf, not descended
	Window  time.Duration
	Server  serverConfig `cf_json:"server"` // tagged struct: one whole-value leaf
	Nested  serverConfig // untagged struct: flattened into its leaves
}

type serverConfig struct {
	Host string `cf_env:"HOST"`
	Port int    `cf_env:"PORT"`
}

func TestForTreatsAtomicStructsAsLeaves(t *testing.T) {
	plan := For(reflect.TypeOf(withAtomicStructs{}))
	got := map[string]bool{}
	for _, f := range plan.Fields {
		got[f.Struct.Name] = true
	}
	// Created (time.Time via TextUnmarshaler), Window (time.Duration, a leaf
	// kind), and Server (tagged struct) each stay a single leaf; only the
	// untagged Nested is flattened into Host+Port.
	want := []string{"Created", "Window", "Server", "Host", "Port"}
	if len(plan.Fields) != len(want) {
		t.Fatalf("got %d fields %v, want %d %v", len(plan.Fields), got, len(want), want)
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("missing expected leaf %q; got %v", name, got)
		}
	}
}

type onlyUnexported struct {
	secret string //nolint:unused // exercises the zero-exported-fields leaf rule
}

type wrapsOnlyUnexported struct {
	Blob onlyUnexported `cf_env:"BLOB"`
}

func TestForTreatsZeroExportedFieldStructAsLeaf(t *testing.T) {
	plan := For(reflect.TypeOf(wrapsOnlyUnexported{}))
	if len(plan.Fields) != 1 || plan.Fields[0].Struct.Name != "Blob" {
		t.Fatalf("got %+v, want a single Blob leaf", plan.Fields)
	}
}

func TestForCachesByType(t *testing.T) {
	first := For(reflect.TypeOf(outer{}))
	second := For(reflect.TypeOf(outer{}))
	if first != second {
		t.Error("expected For to return the same cached *Plan for the same type")
	}
}
