package git

import (
	"testing"

	"github.com/goreflect/gostructor"
)

// TestGitSource_KeyValueDecoder verifies a git source can read a non-JSON
// format by plugging in a decoder — here the zero-dep key/value decoder.
func TestGitSource_KeyValueDecoder(t *testing.T) {
	dir, _ := testRepo(t, "app.env", "HOST=db.internal\nPORT=5432\nNAME=orders\n")

	src, err := New(Options{
		Repo:    dir,
		Ref:     "master",
		Path:    "app.env",
		Decoder: gostructor.DecodeKeyValue,
	})
	if err != nil {
		t.Fatal(err)
	}
	type Config struct {
		Host string `cfg:"host,git:HOST"`
		Port int    `cfg:"port,git:PORT"`
		Name string `cfg:"name,git:NAME"`
	}
	var cfg Config
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	if cfg.Host != "db.internal" || cfg.Port != 5432 || cfg.Name != "orders" {
		t.Fatalf("got %+v", cfg)
	}
}

// TestGitSource_INIDecoder verifies the INI decoder with a nested [section].
func TestGitSource_INIDecoder(t *testing.T) {
	dir, _ := testRepo(t, "config.ini", "[server]\nhost = api.internal\nport = 8443\n")

	src, err := New(Options{
		Repo:    dir,
		Ref:     "master",
		Path:    "config.ini",
		Decoder: gostructor.DecodeINI,
	})
	if err != nil {
		t.Fatal(err)
	}
	type Config struct {
		Host string `cfg:"host,git:server.host"`
		Port int    `cfg:"port,git:server.port"`
	}
	var cfg Config
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	if cfg.Host != "api.internal" || cfg.Port != 8443 {
		t.Fatalf("got %+v", cfg)
	}
}
