package gostructor

import "github.com/goreflect/gostructor/internal/tools"

// SourceMap is the default identity of the in-memory Map source.
const SourceMap = "map"

// LookupKey resolves field against an already-decoded, possibly nested
// map[string]any — the shape encoding/json and go-yaml produce — using the same
// addressing as the JSON source: the field's per-source override for sourceName
// if it has one, otherwise its base name, traversed as a dot-separated path
// ("server.host"). It returns found=false when the field has no key for this
// source or the path is absent/nil.
//
// It is exported so remote, map-backed sources built as separate modules (git,
// Spring Cloud Config, and the like) resolve keys exactly the way the built-in
// file sources do, instead of each re-implementing traversal.
func LookupKey(field FieldContext, sourceName string, data map[string]any) (any, bool) {
	key := field.SourceKey(sourceName, Identity)
	if key == "" {
		return nil, false
	}
	// Dotted traversal for nested maps (JSON/YAML/TOML/HOCON), then an exact
	// top-level match so a flat map from a key/value or INI file resolves a
	// dotted key too (e.g. the literal key "server.host").
	if value, ok := tools.LookupPath(data, key); ok && value != nil {
		return value, true
	}
	if value, ok := data[key]; ok && value != nil {
		return value, true
	}
	return nil, false
}

// Map returns a Source backed by an in-memory map, resolving fields with the
// same nested-key addressing as the JSON source (see LookupKey). It is useful
// as a highest-priority override source in tests and programmatic setups — put
// it first in WithSources — and is the shared resolution core the remote
// map-backed adapters delegate to. name is the source's identity and its cfg
// per-source override key; an empty name defaults to SourceMap.
func Map(name string, data map[string]any) Source {
	if name == "" {
		name = SourceMap
	}
	return &mapSource{name: name, data: data}
}

type mapSource struct {
	name string
	data map[string]any
}

func (m *mapSource) Name() string { return m.name }

func (m *mapSource) Resolve(field FieldContext) (any, bool, error) {
	value, ok := LookupKey(field, m.name, m.data)
	return value, ok, nil
}
