package structplan

import (
	"reflect"
	"testing"
	"time"
)

type inner struct {
	Host string `cfg:"host"`
}

type outer struct {
	Name   string `cfg:"name"`
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
	Created time.Time `cfg:"created"` // TextUnmarshaler: one leaf, not descended
	Window  time.Duration
	Server  serverConfig `cfg:"server"` // tagged struct: one whole-value leaf
	Nested  serverConfig // untagged struct: flattened into its leaves
}

type serverConfig struct {
	Host string `cfg:"host"`
	Port int    `cfg:"port"`
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

// TestForTreatsGosOnlyStructAsLeaf checks a struct-typed field with only a gos
// tag (no cfg) is still treated as an atomic leaf, not descended into.
func TestForTreatsGosOnlyStructAsLeaf(t *testing.T) {
	type payload struct {
		A string
		B string
	}
	type cfgT struct {
		Blob payload `gos:"optional"`
	}
	plan := For(reflect.TypeOf(cfgT{}))
	if len(plan.Fields) != 1 || plan.Fields[0].Struct.Name != "Blob" {
		t.Fatalf("got %+v, want a single Blob leaf", plan.Fields)
	}
}

type onlyUnexported struct {
	secret string //nolint:unused // exercises the zero-exported-fields leaf rule
}

type wrapsOnlyUnexported struct {
	Blob onlyUnexported `cfg:"blob"`
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

// TestForParsesCfgAndGos checks the per-field cfg/gos metadata is parsed and
// cached on the Field.
func TestForParsesCfgAndGos(t *testing.T) {
	type cfgT struct {
		Port int `cfg:"port,env:SERVER_PORT,json:server.port" gos:"default:8080,sep:|,secret"`
	}
	plan := For(reflect.TypeOf(cfgT{}))
	if len(plan.Fields) != 1 {
		t.Fatalf("got %d fields, want 1", len(plan.Fields))
	}
	f := plan.Fields[0]
	if f.Base != "port" {
		t.Errorf("Base = %q, want port", f.Base)
	}
	if f.Overrides["env"] != "SERVER_PORT" || f.Overrides["json"] != "server.port" {
		t.Errorf("Overrides = %+v", f.Overrides)
	}
	if f.Meta["default"] != "8080" || f.Meta["sep"] != "|" {
		t.Errorf("Meta = %+v", f.Meta)
	}
	if !f.Flags["secret"] {
		t.Errorf("Flags = %+v, want secret set", f.Flags)
	}
}

func TestParseCfg(t *testing.T) {
	cases := []struct {
		tag      string
		wantBase string
		wantOver map[string]string
	}{
		{"", "", nil},
		{"port", "port", nil},
		{"port,env:PORT", "port", map[string]string{"env": "PORT"}},
		{"host,json:server.host,env:HOST", "host", map[string]string{"json": "server.host", "env": "HOST"}},
		{" host , env:HOST ", "host", map[string]string{"env": "HOST"}},
		{",env:HOST", "", map[string]string{"env": "HOST"}}, // empty base, override only
		{"port,bogus", "port", nil},                         // bare non-override token ignored
	}
	for _, tc := range cases {
		base, over := ParseCfg(tc.tag)
		if base != tc.wantBase {
			t.Errorf("ParseCfg(%q) base = %q, want %q", tc.tag, base, tc.wantBase)
		}
		if !reflect.DeepEqual(over, tc.wantOver) {
			t.Errorf("ParseCfg(%q) overrides = %+v, want %+v", tc.tag, over, tc.wantOver)
		}
	}
}

func TestParseGos(t *testing.T) {
	cases := []struct {
		tag       string
		wantMeta  map[string]string
		wantFlags map[string]bool
	}{
		{"", nil, nil},
		{"secret", nil, map[string]bool{"secret": true}},
		{"default:8080", map[string]string{"default": "8080"}, nil},
		{"secret,optional", nil, map[string]bool{"secret": true, "optional": true}},
		{"default:localhost:6379", map[string]string{"default": "localhost:6379"}, nil}, // colon in value kept
		{"sep:|,default:a|b|c", map[string]string{"sep": "|", "default": "a|b|c"}, nil},
		{"secret,default:", map[string]string{"default": ""}, map[string]bool{"secret": true}}, // empty default value
	}
	for _, tc := range cases {
		meta, flags := ParseGos(tc.tag)
		if !reflect.DeepEqual(meta, tc.wantMeta) {
			t.Errorf("ParseGos(%q) meta = %+v, want %+v", tc.tag, meta, tc.wantMeta)
		}
		if !reflect.DeepEqual(flags, tc.wantFlags) {
			t.Errorf("ParseGos(%q) flags = %+v, want %+v", tc.tag, flags, tc.wantFlags)
		}
	}
}
