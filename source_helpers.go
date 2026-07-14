package gostructor

import (
	"reflect"
	"strings"
)

// splitIfSlice returns raw split on commas as []string when field's target
// type is a slice or array, or raw unchanged otherwise. Used by sources
// (Env, Default) whose only representation of a value is a flat string.
func splitIfSlice(field FieldContext, raw string) any {
	switch field.Type.Kind() {
	case reflect.Slice, reflect.Array:
		parts := strings.Split(raw, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	default:
		return raw
	}
}
