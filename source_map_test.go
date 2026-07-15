package gostructor

import "testing"

func TestMapSource_NestedAndOverride(t *testing.T) {
	type Config struct {
		Host string `cfg:"host,map:server.host"`
		Port int    `cfg:"port"`
		Env  string `cfg:"env" gos:"default:local"`
	}
	data := map[string]any{
		"server": map[string]any{"host": "db.internal"},
		"port":   5432,
	}
	cfg, err := Configure(&Config{}, WithSources(Map(SourceMap, data), Default()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "db.internal" {
		t.Errorf("Host = %q, want db.internal", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("Port = %d, want 5432", cfg.Port)
	}
	if cfg.Env != "local" {
		t.Errorf("Env = %q, want default local", cfg.Env)
	}
}

func TestMapSource_WinsWhenFirst(t *testing.T) {
	type Config struct {
		Value string `cfg:"value" gos:"default:from-default"`
	}
	data := map[string]any{"value": "from-map"}
	cfg, err := Configure(&Config{}, WithSources(Map("map", data), Default()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Value != "from-map" {
		t.Errorf("Value = %q, want from-map (map source has priority)", cfg.Value)
	}
}
