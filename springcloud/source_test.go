package springcloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/snapshot"
)

// configServer is a minimal fake Spring Cloud Config Server whose response can
// be swapped at runtime to simulate a config change.
type configServer struct {
	mu   sync.Mutex
	env  environment
	hits int32
}

func (c *configServer) set(env environment) {
	c.mu.Lock()
	c.env = env
	c.mu.Unlock()
}

func (c *configServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt32(&c.hits, 1)
	c.mu.Lock()
	env := c.env
	c.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(env)
}

func env(version string, props map[string]any) environment {
	return environment{
		Name:            "orders",
		Profiles:        []string{"production"},
		Version:         version,
		PropertySources: []propertySource{{Name: "git:orders-production.yml", Source: props}},
	}
}

type appConfig struct {
	Host  string `cfg:"server.host,springcloud:server.host"`
	Port  int    `cfg:"server.port,springcloud:server.port"`
	Level string `cfg:"log.level" gos:"default:info"`
}

func TestSpringCloud_Resolves(t *testing.T) {
	cs := &configServer{}
	cs.set(env("v1", map[string]any{"server.host": "api.internal", "server.port": 8443}))
	ts := httptest.NewServer(cs)
	defer ts.Close()

	src, err := New(Options{Address: ts.URL, Application: "orders", Profile: "production"})
	if err != nil {
		t.Fatal(err)
	}
	var cfg appConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	if cfg.Host != "api.internal" || cfg.Port != 8443 || cfg.Level != "info" {
		t.Fatalf("got %+v", cfg)
	}
}

func TestSpringCloud_PrecedenceEarlierSourceWins(t *testing.T) {
	cs := &configServer{}
	cs.set(environment{
		Version: "v1",
		PropertySources: []propertySource{
			{Name: "high", Source: map[string]any{"server.host": "override"}},
			{Name: "low", Source: map[string]any{"server.host": "base", "server.port": 80}},
		},
	})
	ts := httptest.NewServer(cs)
	defer ts.Close()

	src, _ := New(Options{Address: ts.URL, Application: "orders"})
	var cfg appConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "override" {
		t.Fatalf("Host = %q, want override (earlier property source wins)", cfg.Host)
	}
	if cfg.Port != 80 {
		t.Fatalf("Port = %d, want 80 (from the lower source)", cfg.Port)
	}
}

func TestSpringCloud_WatchPollsForChange(t *testing.T) {
	cs := &configServer{}
	cs.set(env("v1", map[string]any{"server.host": "old", "server.port": 1}))
	ts := httptest.NewServer(cs)
	defer ts.Close()

	src, _ := New(Options{Address: ts.URL, Application: "orders", Poll: 30 * time.Millisecond})
	var cfg appConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changed := make(chan struct{}, 4)
	go src.Watch(ctx, func() { changed <- struct{}{} })

	cs.set(env("v2", map[string]any{"server.host": "new", "server.port": 2}))

	select {
	case <-changed:
	case <-time.After(3 * time.Second):
		t.Fatal("Watch did not fire on version change")
	}
	var reloaded appConfig
	if _, err := gostructor.Configure(&reloaded, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatal(err)
	}
	if reloaded.Host != "new" || reloaded.Port != 2 {
		t.Fatalf("after change got %+v, want host=new port=2", reloaded)
	}
}

func TestSpringCloud_SnapshotFallback(t *testing.T) {
	store, _ := snapshot.NewDirStore(t.TempDir())
	good := env("cached-version", map[string]any{"server.host": "cached", "server.port": 9})
	raw, _ := json.Marshal(good)
	if err := store.Save("cached-version", raw); err != nil {
		t.Fatal(err)
	}

	// Point at a dead address so the fetch fails.
	src, _ := New(Options{
		Address:     "http://127.0.0.1:1",
		Application: "orders",
		Snapshot:    store,
		HTTPClient:  &http.Client{Timeout: 200 * time.Millisecond},
	})
	var cfg appConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatalf("expected snapshot fallback, got %v", err)
	}
	if cfg.Host != "cached" || cfg.Port != 9 {
		t.Fatalf("snapshot fallback got %+v, want host=cached port=9", cfg)
	}
}
