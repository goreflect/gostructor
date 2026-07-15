// Package consul resolves gostructor fields from a HashiCorp Consul KV prefix.
// Each field maps to a key under the prefix; the key is the field's per-source
// override (`cfg:"host,consul:database/host"`) or, lacking one, its base name.
//
// It implements gostructor.Watchable using Consul blocking queries: a single
// long-poll on the prefix returns as soon as any key under it changes, so a
// reload fires promptly without busy polling. An optional snapshot store
// (gostructor/snapshot) keeps the last-known-good KV set so the app still starts
// when Consul is unreachable.
package consul

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/snapshot"
)

// DefaultName is this source's identity and cfg per-source override key.
const DefaultName = "consul"

// DefaultWaitTime bounds a single blocking query; Consul returns earlier the
// moment a key changes. It also caps how long Watch takes to notice its context
// was cancelled between queries.
const DefaultWaitTime = 5 * time.Minute

// kvLister is the slice of *consulapi.KV this source uses, so tests can stand in
// a fake instead of a live agent.
type kvLister interface {
	List(prefix string, q *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error)
}

// Options configures a consul Source.
type Options struct {
	// Address is the Consul HTTP address (host:port). Empty uses the api
	// defaults (CONSUL_HTTP_ADDR, else 127.0.0.1:8500).
	Address string
	// Prefix is the KV path fields are read from, e.g. "config/orders". A field
	// key is joined onto it with "/".
	Prefix string
	// Token is the optional ACL token (empty uses CONSUL_HTTP_TOKEN if set).
	Token string
	// Datacenter optionally targets a non-default datacenter.
	Datacenter string
	// WaitTime bounds one blocking query; zero uses DefaultWaitTime.
	WaitTime time.Duration
	// Snapshot, if set, stores each good KV read and is served when Consul is
	// unreachable at startup.
	Snapshot snapshot.Store
	// Name overrides the source identity (DefaultName otherwise).
	Name string
	// Logger receives watch/fallback diagnostics; nil discards them.
	Logger *slog.Logger
}

type source struct {
	opts     Options
	name     string
	prefix   string
	waitTime time.Duration
	log      *slog.Logger

	newKV func() (kvLister, error)

	mu      sync.Mutex
	kv      kvLister
	data    map[string]string
	index   uint64
	loaded  bool
	loadErr error
}

// New returns a Consul-backed gostructor.Source (also a gostructor.Watchable).
// The client is built lazily on first use, so constructing a source never
// blocks; connection problems surface at fill time (or trigger a snapshot
// fallback).
func New(opts Options) (*source, error) {
	if opts.Prefix == "" {
		return nil, fmt.Errorf("gostructor/consul: Prefix is required")
	}
	s := newSource(opts)
	s.newKV = func() (kvLister, error) {
		cfg := consulapi.DefaultConfig()
		if opts.Address != "" {
			cfg.Address = opts.Address
		}
		if opts.Token != "" {
			cfg.Token = opts.Token
		}
		if opts.Datacenter != "" {
			cfg.Datacenter = opts.Datacenter
		}
		client, err := consulapi.NewClient(cfg)
		if err != nil {
			return nil, err
		}
		return client.KV(), nil
	}
	return s, nil
}

// newSource builds the common source shell shared by New and the tests.
func newSource(opts Options) *source {
	s := &source{
		opts:     opts,
		name:     opts.Name,
		prefix:   strings.Trim(opts.Prefix, "/"),
		waitTime: opts.WaitTime,
		log:      opts.Logger,
	}
	if s.name == "" {
		s.name = DefaultName
	}
	if s.waitTime <= 0 {
		s.waitTime = DefaultWaitTime
	}
	if s.log == nil {
		s.log = slog.New(slog.DiscardHandler)
	}
	return s
}

func (s *source) Name() string { return s.name }

// Resolve fills a field from the KV set, loading it on first use.
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

// Watch runs a single blocking query against the prefix: Consul returns as soon
// as any key changes (or WaitTime elapses), at which point the KV set is
// refreshed and onChange fires. It blocks until ctx is cancelled.
func (s *source) Watch(ctx context.Context, onChange func()) error {
	if err := s.ensureLoaded(); err != nil {
		s.log.Warn("gostructor/consul: initial load failed, watch will keep retrying", "err", err)
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.mu.Lock()
		kv, index := s.kv, s.index
		s.mu.Unlock()
		if kv == nil {
			// Client never built (initial load failed hard); retry shortly.
			if !sleep(ctx, time.Second) {
				return ctx.Err()
			}
			if err := s.ensureLoaded(); err != nil {
				continue
			}
			continue
		}

		pairs, meta, err := kv.List(s.prefix, (&consulapi.QueryOptions{
			WaitIndex: index,
			WaitTime:  s.waitTime,
		}).WithContext(ctx))
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			s.log.Warn("gostructor/consul: blocking query failed, retrying", "err", err)
			if !sleep(ctx, time.Second) {
				return ctx.Err()
			}
			continue
		}
		if meta == nil || meta.LastIndex == index {
			// WaitTime elapsed with no change (or a stale/again index): keep
			// polling with the same baseline.
			continue
		}
		s.applyPairs(pairs, meta.LastIndex)
		onChange()
	}
}

func (s *source) ensureLoaded() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loaded {
		return s.loadErr
	}
	s.loaded = true

	if s.kv == nil {
		kv, err := s.newKV()
		if err != nil {
			s.loadErr = fmt.Errorf("gostructor/consul: creating client: %w", err)
			if s.loadFromSnapshotLocked(s.loadErr) {
				s.loadErr = nil
			}
			return s.loadErr
		}
		s.kv = kv
	}

	pairs, meta, err := s.kv.List(s.prefix, nil)
	if err != nil {
		s.loadErr = fmt.Errorf("gostructor/consul: listing %q: %w", s.prefix, err)
		if s.loadFromSnapshotLocked(s.loadErr) {
			s.loadErr = nil
		}
		return s.loadErr
	}
	var index uint64
	if meta != nil {
		index = meta.LastIndex
	}
	s.applyPairsLocked(pairs, index)
	return nil
}

// applyPairs refreshes the KV snapshot under the lock; used from the watch loop.
func (s *source) applyPairs(pairs consulapi.KVPairs, index uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.applyPairsLocked(pairs, index)
}

// applyPairsLocked rebuilds the relative-key map from a KV listing and saves a
// snapshot. The caller holds mu.
func (s *source) applyPairsLocked(pairs consulapi.KVPairs, index uint64) {
	data := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		if pair == nil {
			continue
		}
		rel := strings.TrimPrefix(pair.Key, s.prefix)
		rel = strings.Trim(rel, "/")
		if rel == "" {
			continue // the prefix folder entry itself
		}
		data[rel] = string(pair.Value)
	}
	s.data = data
	s.index = index
	if s.opts.Snapshot != nil {
		if raw, err := json.Marshal(data); err == nil {
			if err := s.opts.Snapshot.Save(fmt.Sprint(index), raw); err != nil {
				s.log.Warn("gostructor/consul: saving snapshot failed", "err", err)
			}
		}
	}
}

// loadFromSnapshotLocked serves the last-known-good KV set after a failed
// connect/list. The caller holds mu.
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
		s.log.Warn("gostructor/consul: stored snapshot is undecodable", "err", err)
		return false
	}
	s.data = data
	s.log.Warn("gostructor/consul: Consul unavailable, serving last-known-good snapshot",
		"cause", cause, "version", version)
	return true
}

// splitIfSlice splits a flat KV string into elements for a slice/array field,
// mirroring the core env/default sources; scalar fields pass through unchanged.
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

// sleep waits for d or ctx cancellation, reporting false if ctx was cancelled.
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
