package etcd

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/snapshot"
)

// fakeBackend is an in-memory etcd stand-in. Full keys are stored as
// prefix + "/" + rel to match what the real clientBackend.Get returns.
type fakeBackend struct {
	mu     sync.Mutex
	kv     map[string]string
	prefix string
	notify chan struct{}
	getErr error
	closed bool
}

func newFakeBackend(prefix string, kv map[string]string) *fakeBackend {
	return &fakeBackend{prefix: prefix, kv: full(prefix, kv), notify: make(chan struct{}, 1)}
}

func full(prefix string, kv map[string]string) map[string]string {
	out := make(map[string]string, len(kv))
	for k, v := range kv {
		out[prefix+"/"+k] = v
	}
	return out
}

func (f *fakeBackend) set(kv map[string]string) {
	f.mu.Lock()
	f.kv = full(f.prefix, kv)
	f.mu.Unlock()
	select {
	case f.notify <- struct{}{}:
	default:
	}
}

func (f *fakeBackend) Get(ctx context.Context, prefix string) (map[string]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return nil, f.getErr
	}
	out := make(map[string]string, len(f.kv))
	for k, v := range f.kv {
		out[k] = v
	}
	return out, nil
}

func (f *fakeBackend) Watch(ctx context.Context, prefix string, onChange func()) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-f.notify:
			onChange()
		}
	}
}

func (f *fakeBackend) Close() error {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
	return nil
}

func withFakeBackend(opts Options, be backend) *source {
	s := newSource(opts)
	s.newBackend = func() (backend, error) { return be, nil }
	return s
}

type svcConfig struct {
	Host  string `cfg:"host"`
	Port  int    `cfg:"port" gos:"optional"`
	Level string `cfg:"level" gos:"default:info"`
}

func TestEtcdSource_Resolves(t *testing.T) {
	fake := newFakeBackend("config/orders", map[string]string{"host": "10.0.0.5", "port": "5432"})
	src := withFakeBackend(Options{Endpoints: []string{"x"}, Prefix: "config/orders"}, fake)

	var cfg svcConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	if cfg.Host != "10.0.0.5" || cfg.Port != 5432 || cfg.Level != "info" {
		t.Fatalf("got %+v", cfg)
	}
}

func TestEtcdSource_WatchReloadsOnEvent(t *testing.T) {
	fake := newFakeBackend("config/orders", map[string]string{"host": "old"})
	src := withFakeBackend(Options{Endpoints: []string{"x"}, Prefix: "config/orders"}, fake)

	var cfg svcConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changed := make(chan struct{}, 4)
	go src.Watch(ctx, func() { changed <- struct{}{} })

	time.Sleep(30 * time.Millisecond)
	fake.set(map[string]string{"host": "new"})

	select {
	case <-changed:
	case <-time.After(3 * time.Second):
		t.Fatal("Watch did not fire on etcd event")
	}
	var reloaded svcConfig
	if _, err := gostructor.Configure(&reloaded, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatal(err)
	}
	if reloaded.Host != "new" {
		t.Fatalf("after event Host = %q, want new", reloaded.Host)
	}
}

func TestEtcdSource_SnapshotFallback(t *testing.T) {
	store, _ := snapshot.NewDirStore(t.TempDir())
	if err := store.Save("3", []byte(`{"host":"cached-host"}`)); err != nil {
		t.Fatal(err)
	}
	s := newSource(Options{Endpoints: []string{"x"}, Prefix: "config/orders", Snapshot: store})
	s.newBackend = func() (backend, error) { return nil, context.DeadlineExceeded }

	var cfg svcConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(s, gostructor.Default())); err != nil {
		t.Fatalf("expected snapshot fallback, got %v", err)
	}
	if cfg.Host != "cached-host" {
		t.Fatalf("Host = %q, want cached-host", cfg.Host)
	}
}
