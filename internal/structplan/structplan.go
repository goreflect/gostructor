// Package structplan builds and caches, per reflect.Type, the flattened list
// of settable fields on a struct, each with its config metadata already parsed
// from the two struct tags gostructor understands:
//
//   - cfg (routing & naming): `cfg:"base_name,source:override,source2:override"`.
//     The first element is the field's base name; the rest are per-source key
//     overrides.
//   - gos (behavior & metadata): `gos:"default:8080,sep:;,secret,optional"`, a
//     comma-list of bare flags (secret, optional) and key:value meta
//     (default, sep).
//
// The reflection walk runs once per type and is reused across Configure calls
// rather than repeated each call. Both tags are parsed with plain strings.Split.
package structplan

import (
	"encoding"
	"reflect"
	"strings"
	"sync"
)

var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

// CfgTag and GosTag are the struct tag names gostructor reads.
const (
	CfgTag = "cfg"
	GosTag = "gos"
)

// Field is one leaf (non-struct) field discovered while flattening a
// struct's fields, including nested/embedded structs, with its cfg/gos tags
// parsed once and cached.
type Field struct {
	Struct reflect.StructField
	Index  []int // path suitable for reflect.Value.FieldByIndex

	// Base is the base name (first cfg element), used by each source's naming
	// strategy to derive its key when no explicit override is given.
	Base string
	// Overrides maps a source name (e.g. "env") to the explicit key it should
	// use for this field, from `cfg:"base,env:OVERRIDE"`.
	Overrides map[string]string
	// Meta holds gos key:value entries (e.g. "default" -> "8080").
	Meta map[string]string
	// Flags holds gos bare flags that are present (e.g. "secret", "optional").
	Flags map[string]bool
}

// Plan is the flattened field list for one struct type.
type Plan struct {
	Fields []Field
}

var cache sync.Map // reflect.Type -> *Plan

// For returns the cached Plan for t, building and caching it on first use.
// t must be a struct type (not a pointer to one).
func For(t reflect.Type) *Plan {
	if cached, ok := cache.Load(t); ok {
		return cached.(*Plan)
	}
	plan := build(t)
	actual, _ := cache.LoadOrStore(t, plan)
	return actual.(*Plan)
}

func build(t reflect.Type) *Plan {
	plan := &Plan{}
	walk(t, nil, plan)
	return plan
}

func walk(t reflect.Type, prefix []int, plan *Plan) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			// unexported field: not settable without unsafe tricks: skip.
			continue
		}
		index := append(append([]int{}, prefix...), i)
		if field.Type.Kind() == reflect.Struct && !isAtomic(field) {
			walk(field.Type, index, plan)
			continue
		}
		base, overrides := ParseCfg(field.Tag.Get(CfgTag))
		meta, flags := ParseGos(field.Tag.Get(GosTag))
		plan.Fields = append(plan.Fields, Field{
			Struct:    field,
			Index:     index,
			Base:      base,
			Overrides: overrides,
			Meta:      meta,
			Flags:     flags,
		})
	}
}

// ParseCfg splits a cfg tag value into its base name and per-source overrides.
// "port,env:SERVER_PORT,json:server.host" -> "port", {env:SERVER_PORT,
// json:server.host}. An empty tag yields an empty base and no overrides.
func ParseCfg(tag string) (base string, overrides map[string]string) {
	if tag == "" {
		return "", nil
	}
	parts := strings.Split(tag, ",")
	base = strings.TrimSpace(parts[0])
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, key, ok := strings.Cut(part, ":")
		if !ok {
			// A bare token in cfg with no "source:" prefix is meaningless for
			// routing; ignore it rather than guess.
			continue
		}
		if overrides == nil {
			overrides = map[string]string{}
		}
		overrides[strings.TrimSpace(name)] = strings.TrimSpace(key)
	}
	return base, overrides
}

// ParseGos splits a gos tag value into meta (key:value) and flags (bare
// tokens). "default:8080,secret,optional" -> {default:8080}, {secret, optional}.
//
// Comma is the list separator, so a meta value may not contain a comma. For a
// multi-valued slice default, pick a non-comma separator with sep and use it in
// the default too, e.g. `gos:"sep:|,default:a|b|c"`.
func ParseGos(tag string) (meta map[string]string, flags map[string]bool) {
	if tag == "" {
		return nil, nil
	}
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if name, value, ok := strings.Cut(part, ":"); ok {
			if meta == nil {
				meta = map[string]string{}
			}
			meta[strings.TrimSpace(name)] = value
		} else {
			if flags == nil {
				flags = map[string]bool{}
			}
			flags[part] = true
		}
	}
	return meta, flags
}

// isAtomic reports whether a struct-typed field should be resolved and
// converted as a single leaf value rather than recursed into and flattened.
// That holds when the field carries a config tag (it is meant to be filled
// from one source key), when its type parses itself from text (time.Time,
// net.IP), or when it has no exported fields to flatten. Without the last
// case such a field would yield zero leaves and be silently left unset.
func isAtomic(field reflect.StructField) bool {
	if _, ok := field.Tag.Lookup(CfgTag); ok {
		return true
	}
	if _, ok := field.Tag.Lookup(GosTag); ok {
		return true
	}
	if reflect.PointerTo(field.Type).Implements(textUnmarshalerType) {
		return true
	}
	return !hasExportedField(field.Type)
}

func hasExportedField(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).PkgPath == "" {
			return true
		}
	}
	return false
}
