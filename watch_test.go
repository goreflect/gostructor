package gostructor

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeWatchable is an in-memory Source whose value can change at runtime and
// which fires onChange when told, so the Watch driver can be tested without any
// real backend.
type fakeWatchable struct {
	mu       sync.RWMutex
	value    string
	onChange func()
	ready    chan struct{}
}

func newFakeWatchable(initial string) *fakeWatchable {
	return &fakeWatchable{value: initial, ready: make(chan struct{})}
}

func (f *fakeWatchable) Name() string { return "fake" }

func (f *fakeWatchable) Resolve(field FieldContext) (any, bool, error) {
	if field.Base() != "value" {
		return nil, false, nil
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.value, true, nil
}

func (f *fakeWatchable) Watch(ctx context.Context, onChange func()) error {
	f.mu.Lock()
	f.onChange = onChange
	f.mu.Unlock()
	close(f.ready) // signal the test that Watch is subscribed
	<-ctx.Done()
	return ctx.Err()
}

// set updates the value and fires the change signal, mimicking a source that
// refreshes its snapshot before notifying.
func (f *fakeWatchable) set(v string) {
	f.mu.Lock()
	f.value = v
	onChange := f.onChange
	f.mu.Unlock()
	if onChange != nil {
		onChange()
	}
}

type watchConfig struct {
	Value string `cfg:"value"`
}

func TestWatch_InitialFillAndReload(t *testing.T) {
	src := newFakeWatchable("one")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	var got []string
	reloaded := make(chan struct{}, 8)
	onReload := func(c *watchConfig, err error) {
		if err != nil {
			t.Errorf("unexpected reload error: %v", err)
			return
		}
		mu.Lock()
		got = append(got, c.Value)
		mu.Unlock()
		reloaded <- struct{}{}
	}

	done := make(chan error, 1)
	go func() {
		done <- Watch(ctx, &watchConfig{}, onReload,
			WithSources(src, Default()),
			WithDebounce(20*time.Millisecond))
	}()

	<-reloaded  // initial fill delivered
	<-src.ready // Watch subscribed

	src.set("two")
	<-reloaded

	src.set("three")
	<-reloaded

	cancel()
	if err := <-done; err != context.Canceled {
		t.Fatalf("Watch returned %v, want context.Canceled", err)
	}

	mu.Lock()
	defer mu.Unlock()
	want := []string{"one", "two", "three"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("reload sequence = %v, want %v", got, want)
	}
}

func TestWatch_DebounceCoalescesBurst(t *testing.T) {
	src := newFakeWatchable("start")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var reloads int32
	final := make(chan string, 1)
	onReload := func(c *watchConfig, err error) {
		n := atomic.AddInt32(&reloads, 1)
		if n == 1 {
			return // initial fill
		}
		select {
		case final <- c.Value:
		default:
		}
	}

	go Watch(ctx, &watchConfig{}, onReload,
		WithSources(src, Default()),
		WithDebounce(50*time.Millisecond))

	<-src.ready

	// A burst of rapid changes should collapse into a single reload landing on
	// the last value.
	for _, v := range []string{"a", "b", "c", "d"} {
		src.set(v)
		time.Sleep(5 * time.Millisecond)
	}

	select {
	case v := <-final:
		if v != "d" {
			t.Fatalf("debounced reload value = %q, want last value %q", v, "d")
		}
	case <-time.After(time.Second):
		t.Fatal("no reload delivered after burst")
	}

	// Give any stragglers a moment, then assert the burst produced exactly one
	// reload on top of the initial fill.
	time.Sleep(120 * time.Millisecond)
	if n := atomic.LoadInt32(&reloads); n != 2 {
		t.Fatalf("reloads = %d, want 2 (initial + one coalesced)", n)
	}
}

func TestWatch_ValidateRejectsKeepsLastGood(t *testing.T) {
	src := newFakeWatchable("good")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type result struct {
		val string
		err error
	}
	results := make(chan result, 8)
	onReload := func(c *watchConfig, err error) {
		if err != nil {
			results <- result{err: err}
			return
		}
		results <- result{val: c.Value}
	}

	validate := func(c *watchConfig) error {
		if c.Value == "bad" {
			return fmt.Errorf("value %q is rejected", c.Value)
		}
		return nil
	}

	go Watch(ctx, &watchConfig{}, onReload,
		WithSources(src, Default()),
		WithValidate(validate))

	if r := <-results; r.val != "good" {
		t.Fatalf("initial fill = %+v, want good", r)
	}
	<-src.ready // ensure the watcher is subscribed before mutating

	src.set("bad")
	if r := <-results; r.err == nil {
		t.Fatalf("expected validation error for bad value, got %+v", r)
	}

	src.set("better")
	if r := <-results; r.val != "better" {
		t.Fatalf("recovery reload = %+v, want better", r)
	}
}

func TestWatch_InitialFillErrorIsFatal(t *testing.T) {
	// A required field no source resolves must fail Watch at startup, before any
	// onReload.
	type strictConfig struct {
		Must string `cfg:"must"`
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := false
	err := Watch(ctx, &strictConfig{}, func(*strictConfig, error) { called = true },
		WithSources(Default()))
	if err == nil {
		t.Fatal("expected fatal error from initial fill")
	}
	if called {
		t.Fatal("onReload must not be called when the initial fill fails")
	}
}

func TestPollWatch_FiresOnFingerprintChange(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var fp atomic.Value
	fp.Store("v1")
	var changes int32
	fired := make(chan struct{}, 4)

	go PollWatch(ctx, 10*time.Millisecond, func(context.Context) (string, error) {
		return fp.Load().(string), nil
	}, func() {
		atomic.AddInt32(&changes, 1)
		fired <- struct{}{}
	}, nil)

	// Baseline established; no change yet.
	time.Sleep(40 * time.Millisecond)
	if atomic.LoadInt32(&changes) != 0 {
		t.Fatal("PollWatch fired before any fingerprint change")
	}

	fp.Store("v2")
	select {
	case <-fired:
	case <-time.After(time.Second):
		t.Fatal("PollWatch did not fire on fingerprint change")
	}
}
