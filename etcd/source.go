// Package etcd resolves gostructor fields from an etcd v3 key prefix. Each field
// maps to a key under the prefix; the key is the field's per-source override
// (`cfg:"host,etcd:database/host"`) or, lacking one, its base name.
//
// It implements gostructor.Watchable using the native etcd watch API: a single
// watch on the prefix streams change events, and each one refreshes the key set
// and fires a reload — no polling. An optional snapshot store
// (gostructor/snapshot) keeps the last-known-good keys so the app still starts
// when etcd is unreachable.
package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/snapshot"
)

// DefaultName is this source's identity and cfg per-source override key.
const DefaultName = "etcd"

// DefaultDialTimeout bounds the initial connection to etcd.
const DefaultDialTimeout = 5 * time.Second

// backend is the slice of etcd behaviour this source needs, so tests can
// substitute a fake for the real clientv3-backed implementation. Get returns
// full keys (including the prefix); the source strips the prefix itself.
type backend interface {
	Get(ctx context.Context, prefix string) (kvs map[string]string, err error)
	Watch(ctx context.Context, prefix string, onChange func()) error
	Close() error
}

// Options configures an etcd Source.
type Options struct {
	// Endpoints are the etcd client URLs, e.g. []string{"http://127.0.0.1:2379"}.
	// Required.
	Endpoints []string
	// Prefix is the key path fields are read from, e.g. "config/orders". A field
	// key is joined onto it with "/".
	Prefix string
	// Username and Password are optional auth credentials.
	Username string
	Password string
	// DialTimeout bounds the initial connection; zero uses DefaultDialTimeout.
	DialTimeout time.Duration
	// Snapshot, if set, stores each good key read and is served when etcd is
	// unreachable at startup.
	Snapshot snapshot.Store
	// Name overrides the source identity (DefaultName otherwise).
	Name string
	// Logger receives watch/fallback diagnostics; nil discards them.
	Logger *slog.Logger
}

type source struct {
	opts   Options
	name   string
	prefix string
	log    *slog.Logger

	newBackend func() (backend, error)

	mu      sync.Mutex
	be      backend
	data    map[string]string
	loaded  bool
	loadErr error
}

// New returns an etcd-backed gostructor.Source (also a gostructor.Watchable).
// The client is built lazily on first use, so constructing a source never
// blocks; connection problems surface at fill time (or trigger a snapshot
// fallback).
func New(opts Options) (*source, error) {
	if len(opts.Endpoints) == 0 {
		return nil, fmt.Errorf("gostructor/etcd: Endpoints is required")
	}
	if opts.Prefix == "" {
		return nil, fmt.Errorf("gostructor/etcd: Prefix is required")
	}
	s := newSource(opts)
	s.newBackend = func() (backend, error) {
		dial := opts.DialTimeout
		if dial <= 0 {
			dial = DefaultDialTimeout
		}
		client, err := clientv3.New(clientv3.Config{
			Endpoints:   opts.Endpoints,
			DialTimeout: dial,
			Username:    opts.Username,
			Password:    opts.Password,
		})
		if err != nil {
			return nil, err
		}
		return &clientBackend{client: client}, nil
	}
	return s, nil
}

func newSource(opts Options) *source {
	s := &source{
		opts:   opts,
		name:   opts.Name,
		prefix: strings.Trim(opts.Prefix, "/"),
		log:    opts.Logger,
	}
	if s.name == "" {
		s.name = DefaultName
	}
	if s.log == nil {
		s.log = slog.New(slog.DiscardHandler)
	}
	return s
}

func (s *source) Name() string { return s.name }

// Resolve fills a field from the key set, loading it on first use.
func (s *source) Resolve(field gostructor.FieldContext) (any, bool, error) {
	if err := s.ensureLoaded(); err != nil {
		return nil, false, err
	}
	key := field.SourceKey(s.name, gostructor.Identity)
	if key == "" {
		return nil, false, nil
	}
	s.mu.Lock()
	value, ok := s.data[strings.Trim(key, "/")]
	s.mu.Unlock()
	if !ok {
		return nil, false, nil
	}
	return splitIfSlice(field, value), true, nil
}

// Watch streams etcd change events for the prefix; each event refreshes the key
// set and fires onChange. It blocks until ctx is cancelled. A watch failure is
// logged and retried so a transient etcd outage doesn't end live reloads.
func (s *source) Watch(ctx context.Context, onChange func()) error {
	if err := s.ensureLoaded(); err != nil {
		s.log.Warn("gostructor/etcd: initial load failed, watch will keep retrying", "err", err)
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.mu.Lock()
		be := s.be
		s.mu.Unlock()
		if be == nil {
			if !sleep(ctx, time.Second) {
				return ctx.Err()
			}
			_ = s.ensureLoaded()
			continue
		}

		// The watch callback refreshes state before signalling, so the reload
		// reads fresh keys.
		err := be.Watch(ctx, s.prefix, func() {
			if err := s.refresh(ctx); err != nil {
				s.log.Warn("gostructor/etcd: refresh after change failed", "err", err)
				return
			}
			onChange()
		})
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			s.log.Warn("gostructor/etcd: watch stream ended, retrying", "err", err)
		}
		if !sleep(ctx, time.Second) {
			return ctx.Err()
		}
	}
}

func (s *source) ensureLoaded() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loaded {
		return s.loadErr
	}
	s.loaded = true

	if s.be == nil {
		be, err := s.newBackend()
		if err != nil {
			s.loadErr = fmt.Errorf("gostructor/etcd: creating client: %w", err)
			if s.loadFromSnapshotLocked(s.loadErr) {
				s.loadErr = nil
			}
			return s.loadErr
		}
		s.be = be
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultDialTimeout)
	defer cancel()
	kvs, err := s.be.Get(ctx, s.prefix)
	if err != nil {
		s.loadErr = fmt.Errorf("gostructor/etcd: reading %q: %w", s.prefix, err)
		if s.loadFromSnapshotLocked(s.loadErr) {
			s.loadErr = nil
		}
		return s.loadErr
	}
	s.applyLocked(kvs)
	return nil
}

// refresh re-reads the prefix and updates the key set under the lock.
func (s *source) refresh(ctx context.Context) error {
	s.mu.Lock()
	be := s.be
	s.mu.Unlock()
	if be == nil {
		return fmt.Errorf("gostructor/etcd: no client")
	}
	kvs, err := be.Get(ctx, s.prefix)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.applyLocked(kvs)
	s.mu.Unlock()
	return nil
}

// applyLocked rebuilds the relative-key map from a full-key listing and saves a
// snapshot. The caller holds mu.
func (s *source) applyLocked(kvs map[string]string) {
	data := make(map[string]string, len(kvs))
	for k, v := range kvs {
		rel := strings.Trim(strings.TrimPrefix(k, s.prefix), "/")
		if rel == "" {
			continue
		}
		data[rel] = v
	}
	s.data = data
	if s.opts.Snapshot != nil {
		if raw, err := json.Marshal(data); err == nil {
			if err := s.opts.Snapshot.Save(fmt.Sprint(len(data)), raw); err != nil {
				s.log.Warn("gostructor/etcd: saving snapshot failed", "err", err)
			}
		}
	}
}

func (s *source) loadFromSnapshotLocked(cause error) bool {
	if s.opts.Snapshot == nil {
		return false
	}
	version, raw, err := s.opts.Snapshot.Load()
	if err != nil {
		return false
	}
	data := map[string]string{}
	if err := json.Unmarshal(raw, &data); err != nil {
		s.log.Warn("gostructor/etcd: stored snapshot is undecodable", "err", err)
		return false
	}
	s.data = data
	s.log.Warn("gostructor/etcd: etcd unavailable, serving last-known-good snapshot",
		"cause", cause, "version", version)
	return true
}

// Close releases the underlying etcd client, if one was built.
func (s *source) Close() error {
	s.mu.Lock()
	be := s.be
	s.mu.Unlock()
	if be != nil {
		return be.Close()
	}
	return nil
}

// clientBackend adapts a *clientv3.Client to the backend interface.
type clientBackend struct {
	client *clientv3.Client
}

func (c *clientBackend) Get(ctx context.Context, prefix string) (map[string]string, error) {
	resp, err := c.client.Get(ctx, prefix+"/", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		out[string(kv.Key)] = string(kv.Value)
	}
	return out, nil
}

func (c *clientBackend) Watch(ctx context.Context, prefix string, onChange func()) error {
	ch := c.client.Watch(ctx, prefix+"/", clientv3.WithPrefix())
	for resp := range ch {
		if err := resp.Err(); err != nil {
			return err
		}
		if len(resp.Events) > 0 {
			onChange()
		}
	}
	return ctx.Err()
}

func (c *clientBackend) Close() error { return c.client.Close() }

func splitIfSlice(field gostructor.FieldContext, raw string) any {
	switch field.Type.Kind() {
	case reflect.Slice, reflect.Array:
		parts := strings.Split(raw, field.Separator())
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	default:
		return raw
	}
}

func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
