// Package gen is runtime support for the code gostructor-gen emits. It's
// imported by the generated <Type>.gs.go files, not by hand-written application
// code (use the main gostructor package for that). The functions here are the
// reflection-free typed conversions the generated Fill uses to turn a source's
// raw value into a concrete field value, re-exported from internal/convert so
// generated code — which lives in your module and can't import gostructor's
// internal packages — can call them.
//
// Each conversion matches what the reflective engine does for the same value and
// destination type, including the *ConvertError it wraps on failure.
package gen

import (
	"reflect"
	"time"

	"github.com/goreflect/gostructor/internal/convert"
)

// Int converts a resolved value into an int.
func Int(v any) (int, error) { return convert.Int(v) }

// Int8 converts a resolved value into an int8.
func Int8(v any) (int8, error) { return convert.Int8(v) }

// Int16 converts a resolved value into an int16.
func Int16(v any) (int16, error) { return convert.Int16(v) }

// Int32 converts a resolved value into an int32.
func Int32(v any) (int32, error) { return convert.Int32(v) }

// Int64 converts a resolved value into an int64.
func Int64(v any) (int64, error) { return convert.Int64(v) }

// Uint converts a resolved value into a uint.
func Uint(v any) (uint, error) { return convert.Uint(v) }

// Uint8 converts a resolved value into a uint8.
func Uint8(v any) (uint8, error) { return convert.Uint8(v) }

// Uint16 converts a resolved value into a uint16.
func Uint16(v any) (uint16, error) { return convert.Uint16(v) }

// Uint32 converts a resolved value into a uint32.
func Uint32(v any) (uint32, error) { return convert.Uint32(v) }

// Uint64 converts a resolved value into a uint64.
func Uint64(v any) (uint64, error) { return convert.Uint64(v) }

// Float32 converts a resolved value into a float32.
func Float32(v any) (float32, error) { return convert.Float32(v) }

// Float64 converts a resolved value into a float64.
func Float64(v any) (float64, error) { return convert.Float64(v) }

// String converts a resolved value into a string.
func String(v any) (string, error) { return convert.String(v) }

// Bool converts a resolved value into a bool.
func Bool(v any) (bool, error) { return convert.Bool(v) }

// Duration converts a resolved value into a time.Duration.
func Duration(v any) (time.Duration, error) { return convert.Duration(v) }

// StringSlice converts a resolved value into a []string.
func StringSlice(v any) ([]string, error) { return convert.StringSlice(v) }

// Reflective converts a resolved value into T through the same reflective core
// the engine uses, for compound and named field shapes: non-string slices and
// arrays ([]int, [3]bool), maps (map[string]int), and named types without a
// dedicated helper. gostructor-gen emits a Reflective[T] call for these fields,
// so the generated Fill stays reflection-free for the common primitives and
// pays one reflect conversion only for the complex ones (the same conversion the
// reflective engine would have done). The returned error is the same
// *ConvertError the reflective path produces.
func Reflective[T any](v any) (T, error) {
	var zero T
	rv, err := convert.Value(reflect.ValueOf(v), reflect.TypeFor[T]())
	if err != nil {
		return zero, err
	}
	return rv.Interface().(T), nil
}
