// Package structplan builds and caches, per reflect.Type, the flattened list
// of settable fields on a struct. Building this list requires walking every
// field's tags via reflection, which is done once per type and reused across
// repeated Configure calls, instead of re-walking on every call.
package structplan

import (
	"encoding"
	"reflect"
	"strings"
	"sync"
)

var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

// Field is one leaf (non-struct) field discovered while flattening a
// struct's fields, including nested/embedded structs.
type Field struct {
	Struct reflect.StructField
	Index  []int // path suitable for reflect.Value.FieldByIndex
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
		plan.Fields = append(plan.Fields, Field{Struct: field, Index: index})
	}
}

// isAtomic reports whether a struct-typed field should be treated as a single
// leaf value to be resolved and converted as a whole, rather than recursed
// into and flattened. This is true when the field carries a config tag (it is
// meant to be filled from one source key), when its type parses itself from
// text (time.Time, net.IP, ...), or when it has no exported fields to flatten
// - the last case previously yielded zero leaves and silently left the field
// unset.
func isAtomic(field reflect.StructField) bool {
	if strings.Contains(string(field.Tag), "cf_") {
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
