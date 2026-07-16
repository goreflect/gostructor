// Package convert turns the loosely-typed values a config source parses
// (strings, float64s from JSON, []interface{} from YAML) into the concrete Go
// types of the struct fields gostructor is filling.
//
// It lives under internal/ but is imported by the Source implementations in
// the gostructor/yaml, gostructor/toml, and sibling modules; it is not part
// of the public API for external callers.
//
// Value is the single entry point: it handles pointers, TextUnmarshaler types,
// time.Duration, primitives, and slice/array/map/struct composites, recursing
// for elements and fields. The exported ToPrimitive/ToComplex wrappers are for
// callers (and tests) that already hold a destination reflect.Value.
package convert

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	durationType        = reflect.TypeOf(time.Duration(0))
	timeType            = reflect.TypeOf(time.Time{})
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// Value converts source into a new value of destType, applying in order:
// pointer allocation, TextUnmarshaler parsing, the time.Duration special
// case, and finally kind-based primitive/composite conversion. The result
// already has destType's exact (possibly named) type, ready to Set onto a
// field of that type.
func Value(source reflect.Value, destType reflect.Type) (reflect.Value, error) {
	return value(source, destType, "")
}

// ValueWithLayout is Value with a custom time layout (from gos:"layout:...")
// used to parse time.Time fields with time.Parse instead of the RFC3339 that
// time.Time's own TextUnmarshaler enforces. An empty layout behaves exactly
// like Value. The layout propagates through pointer, slice, array and map-value
// recursion, so *time.Time, []time.Time and map[string]time.Time honor it too.
func ValueWithLayout(source reflect.Value, destType reflect.Type, layout string) (reflect.Value, error) {
	return value(source, destType, layout)
}

func value(source reflect.Value, destType reflect.Type, layout string) (reflect.Value, error) {
	if destType.Kind() == reflect.Pointer {
		if destType.Elem().Kind() == reflect.Pointer {
			return reflect.Value{}, conversionError(source, destType, errors.New("multi-level pointers are not supported"))
		}
		elem, err := value(source, destType.Elem(), layout)
		if err != nil {
			return reflect.Value{}, err
		}
		ptr := reflect.New(destType.Elem())
		ptr.Elem().Set(elem)
		return ptr, nil
	}

	// A custom layout parses time.Time with time.Parse, overriding the RFC3339
	// its TextUnmarshaler would otherwise enforce below.
	if layout != "" && destType == timeType && source.Kind() == reflect.String {
		t, err := time.Parse(layout, source.String())
		if err != nil {
			return reflect.Value{}, conversionError(source, destType, err)
		}
		return reflect.ValueOf(t), nil
	}

	// TextUnmarshaler wins for string sources so a type's own text parsing
	// (time.Time, net.IP, custom enums) is honored ahead of kind rules.
	if source.Kind() == reflect.String && reflect.PointerTo(destType).Implements(textUnmarshalerType) {
		ptr := reflect.New(destType)
		if err := ptr.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(source.String())); err != nil {
			return reflect.Value{}, conversionError(source, destType, err)
		}
		return ptr.Elem(), nil
	}

	// time.Duration from a human string like "1h30m"; numeric sources fall
	// through to the int path below (nanoseconds), matching encoding/json.
	if destType == durationType && source.Kind() == reflect.String {
		d, err := time.ParseDuration(source.String())
		if err != nil {
			return reflect.Value{}, conversionError(source, destType, err)
		}
		out := reflect.New(destType).Elem()
		out.SetInt(int64(d))
		return out, nil
	}

	switch destType.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return complexValue(source, destType, layout)
	case reflect.Struct:
		return toStruct(source, destType)
	default:
		return toPrimitive(source, destType)
	}
}

// ToPrimitive converts source into destination's type for scalar kinds:
// string, bool, the int/uint family, and float32/float64.
func ToPrimitive(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	return toPrimitive(source, destination.Type())
}

// ToComplex converts source into destination's type for slice/array/map
// kinds, converting each element/key/value through Value.
func ToComplex(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	return complexValue(source, destination.Type(), "")
}

func complexValue(source reflect.Value, destType reflect.Type, layout string) (reflect.Value, error) {
	switch destType.Kind() {
	case reflect.Slice:
		return toSlice(source, destType, layout)
	case reflect.Array:
		return toArray(source, destType, layout)
	case reflect.Map:
		return toMap(source, destType, layout)
	default:
		return reflect.Value{}, fmt.Errorf("convert: destination kind %s is not a supported complex type", destType.Kind())
	}
}

func toSlice(source reflect.Value, destType reflect.Type, layout string) (reflect.Value, error) {
	if source.Kind() != reflect.Slice && source.Kind() != reflect.Array {
		return reflect.Value{}, fmt.Errorf("convert: cannot convert %s into a slice", source.Kind())
	}
	result := reflect.MakeSlice(destType, source.Len(), source.Len())
	for i := 0; i < source.Len(); i++ {
		element := reflect.ValueOf(source.Index(i).Interface())
		converted, err := value(element, destType.Elem(), layout)
		if err != nil {
			return reflect.Value{}, err
		}
		result.Index(i).Set(converted)
	}
	return result, nil
}

func toArray(source reflect.Value, destType reflect.Type, layout string) (reflect.Value, error) {
	if source.Kind() != reflect.Slice && source.Kind() != reflect.Array {
		return reflect.Value{}, fmt.Errorf("convert: cannot convert %s into an array", source.Kind())
	}
	if source.Len() != destType.Len() {
		return reflect.Value{}, fmt.Errorf("convert: array length mismatch: got %d elements, want %d", source.Len(), destType.Len())
	}
	result := reflect.New(destType).Elem()
	for i := 0; i < source.Len(); i++ {
		element := reflect.ValueOf(source.Index(i).Interface())
		converted, err := value(element, destType.Elem(), layout)
		if err != nil {
			return reflect.Value{}, err
		}
		result.Index(i).Set(converted)
	}
	return result, nil
}

func toMap(source reflect.Value, destType reflect.Type, layout string) (reflect.Value, error) {
	if source.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf("convert: cannot convert %s into a map", source.Kind())
	}
	result := reflect.MakeMapWithSize(destType, source.Len())
	for _, key := range source.MapKeys() {
		convertedKey, err := value(reflect.ValueOf(key.Interface()), destType.Key(), "")
		if err != nil {
			return reflect.Value{}, err
		}
		convertedValue, err := value(reflect.ValueOf(source.MapIndex(key).Interface()), destType.Elem(), layout)
		if err != nil {
			return reflect.Value{}, err
		}
		result.SetMapIndex(convertedKey, convertedValue)
	}
	return result, nil
}

// toStruct fills a struct from an object source (map[string]any, the shape
// JSON/YAML/TOML/HOCON produce for nested objects), matching each exported
// field to a key by name, case-insensitively. Unknown keys are ignored and
// absent fields left zero: a struct here is a collection element or a
// whole-object field, not the top-level target whose missing values are fatal.
func toStruct(source reflect.Value, destType reflect.Type) (reflect.Value, error) {
	if source.Kind() != reflect.Map {
		return reflect.Value{}, conversionError(source, destType, errors.New("expected an object"))
	}
	byName := make(map[string]reflect.Value, source.Len())
	for _, key := range source.MapKeys() {
		name, ok := key.Interface().(string)
		if !ok {
			return reflect.Value{}, conversionError(source, destType, errors.New("object keys must be strings"))
		}
		byName[strings.ToLower(name)] = source.MapIndex(key)
	}
	out := reflect.New(destType).Elem()
	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)
		if field.PkgPath != "" {
			continue // unexported
		}
		raw, ok := byName[strings.ToLower(field.Name)]
		if !ok {
			continue
		}
		converted, err := Value(reflect.ValueOf(raw.Interface()), field.Type)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("convert: field %s: %w", field.Name, err)
		}
		out.Field(i).Set(converted)
	}
	return out, nil
}

var errUnsupportedSourceKind = errors.New("unsupported source kind")

// conversionError formats a convert-package error that names the offending
// value and target type and, when present, wraps the underlying cause
// (a strconv/parse error) so callers can errors.Is/As it.
func conversionError(source reflect.Value, target reflect.Type, cause error) error {
	var value any
	if source.IsValid() {
		value = source.Interface()
	}
	if cause != nil {
		return fmt.Errorf("convert: cannot convert %#v (%s) into %s: %w", value, source.Kind(), target, cause)
	}
	return fmt.Errorf("convert: cannot convert %#v (%s) into %s", value, source.Kind(), target)
}
