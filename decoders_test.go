package gostructor

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDecodeJSON_KeepsNumbersExact(t *testing.T) {
	m, err := DecodeJSON([]byte(`{"big": 9007199254740993, "nested": {"host": "h"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if n, ok := m["big"].(json.Number); !ok || n.String() != "9007199254740993" {
		t.Fatalf("big = %v (%T), want exact json.Number", m["big"], m["big"])
	}
	nested, ok := m["nested"].(map[string]any)
	if !ok || nested["host"] != "h" {
		t.Fatalf("nested = %v", m["nested"])
	}
}

func TestDecodeINI_SectionsBecomeNestedMaps(t *testing.T) {
	m, err := DecodeINI([]byte("global = g\n[server]\nhost = h\nport = 8080\n"))
	if err != nil {
		t.Fatal(err)
	}
	if m["global"] != "g" {
		t.Errorf("global = %v, want g", m["global"])
	}
	server, ok := m["server"].(map[string]any)
	if !ok {
		t.Fatalf("server section = %v (%T), want nested map", m["server"], m["server"])
	}
	if server["host"] != "h" || server["port"] != "8080" {
		t.Errorf("server = %v", server)
	}
}

func TestDecodeKeyValue(t *testing.T) {
	raw := []byte("# comment\nexport HOST=db.internal\nPORT = 5432\nTAGS=\"a,b,c\"\n\n;also a comment\n")
	m, err := DecodeKeyValue(raw)
	if err != nil {
		t.Fatal(err)
	}
	if m["HOST"] != "db.internal" || m["PORT"] != "5432" || m["TAGS"] != "a,b,c" {
		t.Fatalf("got %v", m)
	}
}

func TestDecodeKeyValue_ErrorsOnMissingEquals(t *testing.T) {
	if _, err := DecodeKeyValue([]byte("HOST db.internal\n")); err == nil {
		t.Fatal("expected an error for a line without '='")
	}
}

func TestKeyValueSource_ResolvesAndSplits(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/app.env"
	if err := os.WriteFile(path, []byte("HOST=10.0.0.5\nPORT=5432\nALLOW=a, b, c\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	type Config struct {
		Host  string   `cfg:"host,keyvalue:HOST"`
		Port  int      `cfg:"port,keyvalue:PORT"`
		Allow []string `cfg:"allow,keyvalue:ALLOW"`
		Level string   `cfg:"level" gos:"default:info"`
	}
	cfg, err := Configure(&Config{}, WithSources(KeyValueFile(path), Default()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "10.0.0.5" || cfg.Port != 5432 || cfg.Level != "info" {
		t.Fatalf("got %+v", cfg)
	}
	if len(cfg.Allow) != 3 || cfg.Allow[2] != "c" {
		t.Fatalf("Allow = %v, want [a b c]", cfg.Allow)
	}
}

func TestLookupKey_ExactFallbackForFlatDottedKeys(t *testing.T) {
	// A flat map (as a key/value or INI decode produces) with a literal dotted
	// key must still resolve via LookupKey's exact fallback.
	type Config struct {
		Host string `cfg:"host,flat:server.host"`
	}
	data := map[string]any{"server.host": "flat-hit"}
	cfg, err := Configure(&Config{}, WithSources(Map("flat", data), Default()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "flat-hit" {
		t.Fatalf("Host = %q, want flat-hit", cfg.Host)
	}
}
