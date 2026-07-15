package consul

import (
	"context"
	"sync"
	"testing"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/snapshot"
)

// fakeKV is an in-memory kvLister that supports blocking-query semantics: a
// List with a WaitIndex equal to the current index blocks until the data is
// changed (or ctx is cancelled), mimicking a real Consul agent.
type fakeKV struct {
	mu     sync.Mutex
	pairs  consulapi.KVPairs
	index  uint64
	waiter chan struct{}
}

func newFakeKV(prefix string, kv map[string]string) *fakeKV {
	f := &fakeKV{index: 1, waiter: make(chan struct{})}
	f.pairs = toPairs(prefix, kv)
	return f
}

func toPairs(prefix string, kv map[string]string) consulapi.KVPairs {
	var pairs consulapi.KVPairs
	for k, v := range kv {
		pairs = append(pairs, &consulapi.KVPair{Key: prefix + "/" + k, Value: []byte(v)})
	}
	return pairs
}

func (f *fakeKV) set(prefix string, kv map[string]string) {
	f.mu.Lock()
	f.pairs = toPairs(prefix, kv)
	f.index++
	close(f.waiter)
	f.waiter = make(chan struct{})
	f.mu.Unlock()
}

func (f *fakeKV) List(prefix string, q *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error) {
	f.mu.Lock()
	if q != nil && q.WaitIndex == f.index {
		waiter := f.waiter
		f.mu.Unlock()
		// Block until a change or ctx cancellation, like a real blocking query.
		select {
		case <-waiter:
			f.mu.Lock()
		case <-q.Context().Done():
			f.mu.Lock()
			meta := &consulapi.QueryMeta{LastIndex: f.index}
			pairs := f.pairs
			f.mu.Unlock()
			return pairs, meta, nil
		}
	}
	pairs := f.pairs
	meta := &consulapi.QueryMeta{LastIndex: f.index}
	f.mu.Unlock()
	return pairs, meta, nil
}

// withFakeKV builds a source wired to a fake KV instead of a real client.
func withFakeKV(opts Options, kv kvLister) *source {
	s := newSource(opts)
	s.newKV = func() (kvLister, error) { return kv, nil }
	return s
}

type svcConfig struct {
	Host  string   `cfg:"host"`
	Port  int      `cfg:"database/port,consul:database/port" gos:"optional"`
	Tags  []string `cfg:"tags" gos:"optional"`
	Level string   `cfg:"level" gos:"default:info"`
}

func TestConsulSource_Resolves(t *testing.T) {
	fake := newFakeKV("config/orders", map[string]string{
		"host":          "10.0.0.5",
		"database/port": "5432",
		"tags":          "a, b, c",
	})
	src := withFakeKV(Options{Prefix: "config/orders"}, fake)

	var cfg svcConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	if cfg.Host != "10.0.0.5" || cfg.Port != 5432 || cfg.Level != "info" {
		t.Fatalf("got %+v", cfg)
	}
	if len(cfg.Tags) != 3 || cfg.Tags[2] != "c" {
		t.Fatalf("Tags = %v, want [a b c]", cfg.Tags)
	}
}

func TestConsulSource_WatchBlockingQuery(t *testing.T) {
	fake := newFakeKV("config/orders", map[string]string{"host": "old"})
	src := withFakeKV(Options{Prefix: "config/orders", WaitTime: 200 * time.Millisecond}, fake)

	var cfg svcConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changed := make(chan struct{}, 4)
	go src.Watch(ctx, func() { changed <- struct{}{} })

	// Let the blocking query establish, then change a key.
	time.Sleep(50 * time.Millisecond)
	fake.set("config/orders", map[string]string{"host": "new"})

	select {
	case <-changed:
	case <-time.After(3 * time.Second):
		t.Fatal("Watch did not fire on KV change")
	}

	var reloaded svcConfig
	if _, err := gostructor.Configure(&reloaded, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatal(err)
	}
	if reloaded.Host != "new" {
		t.Fatalf("after change Host = %q, want new", reloaded.Host)
	}
}

func TestConsulSource_SnapshotFallback(t *testing.T) {
	store, _ := snapshot.NewDirStore(t.TempDir())
	// Seed a good snapshot, then simulate a client that cannot be built.
	if err := store.Save("7", []byte(`{"host":"cached-host"}`)); err != nil {
		t.Fatal(err)
	}
	s := newSource(Options{Prefix: "config/orders", Snapshot: store})
	s.newKV = func() (kvLister, error) { return nil, context.DeadlineExceeded }

	var cfg svcConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(s, gostructor.Default())); err != nil {
		t.Fatalf("expected snapshot fallback, got %v", err)
	}
	if cfg.Host != "cached-host" {
		t.Fatalf("Host = %q, want cached-host from snapshot", cfg.Host)
	}
}
