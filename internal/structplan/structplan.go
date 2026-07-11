// Package structplan builds and caches, per reflect.Type, the flattened list
// of settable fields on a struct. Building this list requires walking every
// field's tags via reflection, which is done once per type and reused across
// repeated Configure calls, instead of re-walking on every call.
package structplan

import (
	"reflect"
	"sync"
)

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
		if field.Type.Kind() == reflect.Struct {
			walk(field.Type, index, plan)
			continue
		}
		plan.Fields = append(plan.Fields, Field{Struct: field, Index: index})
	}
}
