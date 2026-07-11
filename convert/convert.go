// Package convert provides reflection-based value conversion between the
// loosely-typed values a config source parses (strings, float64s from JSON,
// []interface{} from YAML, ...) and the concrete Go types of the struct
// fields gostructor is filling.
//
// It is a public package (not internal) because Source implementations
// living in separate modules (gostructor/yaml, gostructor/toml, ...) need
// it too.
package convert

import (
	"errors"
	"fmt"
	"reflect"
)

// ToPrimitive converts source into destination's type for scalar kinds:
// string, bool, the int/uint family, and float32/float64.
func ToPrimitive(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	switch destination.Kind() {
	case reflect.Int:
		return toInt(source, destination)
	case reflect.Int8:
		return toIntSized(source, destination, 8)
	case reflect.Int16:
		return toIntSized(source, destination, 16)
	case reflect.Int32:
		return toIntSized(source, destination, 32)
	case reflect.Int64:
		return toIntSized(source, destination, 64)
	case reflect.Uint:
		return toUint(source, destination)
	case reflect.Uint8:
		return toUintSized(source, destination, 8)
	case reflect.Uint16:
		return toUintSized(source, destination, 16)
	case reflect.Uint32:
		return toUintSized(source, destination, 32)
	case reflect.Uint64:
		return toUintSized(source, destination, 64)
	case reflect.String:
		return toString(source, destination)
	case reflect.Float32:
		return toFloatSized(source, destination, 32)
	case reflect.Float64:
		return toFloatSized(source, destination, 64)
	case reflect.Bool:
		return toBool(source, destination)
	default:
		return reflect.Value{}, fmt.Errorf("convert: destination kind %s is not a primitive type", destination.Kind())
	}
}

// ToComplex converts source into destination's type for slice/array/map
// kinds, converting each element/key/value through ToPrimitive.
func ToComplex(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	switch destination.Kind() {
	case reflect.Slice, reflect.Array:
		return toSlice(source, destination)
	case reflect.Map:
		return toMap(source, destination)
	default:
		return reflect.Value{}, fmt.Errorf("convert: destination kind %s is not a supported complex type", destination.Kind())
	}
}

func toSlice(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	if source.Kind() != reflect.Slice && source.Kind() != reflect.Array {
		return reflect.Value{}, fmt.Errorf("convert: cannot convert %s into a slice", source.Kind())
	}
	result := reflect.MakeSlice(destination.Type(), source.Len(), source.Len())
	for i := 0; i < source.Len(); i++ {
		element := reflect.ValueOf(source.Index(i).Interface())
		converted, err := ToPrimitive(element, result.Index(i))
		if err != nil {
			return reflect.Value{}, err
		}
		result.Index(i).Set(converted)
	}
	return result, nil
}

func toMap(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	if source.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf("convert: cannot convert %s into a map", source.Kind())
	}
	destType := destination.Type()
	result := reflect.MakeMapWithSize(destType, source.Len())
	keyType := destType.Key()
	valueType := destType.Elem()
	for _, key := range source.MapKeys() {
		convertedKey, err := ToPrimitive(reflect.ValueOf(key.Interface()), reflect.New(keyType).Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		convertedValue, err := ToPrimitive(reflect.ValueOf(source.MapIndex(key).Interface()), reflect.New(valueType).Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		result.SetMapIndex(convertedKey, convertedValue)
	}
	return result, nil
}

var errUnsupportedSourceKind = errors.New("convert: unsupported source kind")

func unsupportedSource(source reflect.Value, target string) error {
	return fmt.Errorf("%w: cannot convert %s into %s", errUnsupportedSourceKind, source.Kind(), target)
}
