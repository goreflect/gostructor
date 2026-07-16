package convert

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
)

// This file holds the reflection-free, any-typed siblings of the reflect.Value
// -based conversions in convert.go/primitives.go. They exist for the generated
// fast path (gostructor-gen, Theme 7): the generator knows each field's exact
// Go type at build time and emits a direct call to the matching helper here, so
// a resolved value is converted into a concrete typed value with no
// reflect.Value round-trip on the hot path.
//
// Every helper produces byte-for-byte-identical results and errors to what the
// reflective Value would return for the same input and destination type: the
// common concrete source types (string, the sized numerics, bool, []string) are
// handled inline, and any uncommon type (json.Number, a named string/int type,
// an arbitrary slice/array) falls back to the reflective core so parity is
// guaranteed by construction rather than duplicated by hand. Errors are wrapped
// with the same conversionError used by the reflective path.

var (
	intType     = reflect.TypeOf(int(0))
	int8Type    = reflect.TypeOf(int8(0))
	int16Type   = reflect.TypeOf(int16(0))
	int32Type   = reflect.TypeOf(int32(0))
	int64Type   = reflect.TypeOf(int64(0))
	uintType    = reflect.TypeOf(uint(0))
	uint8Type   = reflect.TypeOf(uint8(0))
	uint16Type  = reflect.TypeOf(uint16(0))
	uint32Type  = reflect.TypeOf(uint32(0))
	uint64Type  = reflect.TypeOf(uint64(0))
	float32Type = reflect.TypeOf(float32(0))
	float64Type = reflect.TypeOf(float64(0))
	stringType  = reflect.TypeOf("")
	boolType    = reflect.TypeOf(false)
	stringSlice = reflect.TypeOf([]string(nil))
)

// asString reports the string form of a string-kinded value, including named
// string types such as json.Number, matching how the reflective path treats a
// reflect.Kind of String. Only the non-plain case pays a reflect call.
func asString(v any) (string, bool) {
	if s, ok := v.(string); ok {
		return s, true
	}
	rv := reflect.ValueOf(v)
	if rv.IsValid() && rv.Kind() == reflect.String {
		return rv.String(), true
	}
	return "", false
}

// int64Of mirrors toInt64 for an any source, reflection-free on the common
// concrete types and delegating to the reflective toInt64 otherwise.
func int64Of(v any) (int64, error) {
	switch x := v.(type) {
	case string:
		return strconv.ParseInt(x, 10, 64)
	case int:
		return int64(x), nil
	case int8:
		return int64(x), nil
	case int16:
		return int64(x), nil
	case int32:
		return int64(x), nil
	case int64:
		return x, nil
	case uint:
		return uintToInt64(uint64(x))
	case uint8:
		return int64(x), nil
	case uint16:
		return int64(x), nil
	case uint32:
		return int64(x), nil
	case uint64:
		return uintToInt64(x)
	case float32:
		return floatToInt64(float64(x))
	case float64:
		return floatToInt64(x)
	default:
		return toInt64(reflect.ValueOf(v))
	}
}

func uintToInt64(u uint64) (int64, error) {
	if u > math.MaxInt64 {
		return 0, fmt.Errorf("value %d overflows int64", u)
	}
	return int64(u), nil
}

func floatToInt64(f float64) (int64, error) {
	if err := checkIntegralFloat(f); err != nil {
		return 0, err
	}
	if f < math.MinInt64 || f >= math.MaxInt64 {
		return 0, fmt.Errorf("value %v overflows int64", f)
	}
	return int64(f), nil
}

// uint64Of mirrors toUint64 for an any source.
func uint64Of(v any) (uint64, error) {
	switch x := v.(type) {
	case string:
		return strconv.ParseUint(x, 10, 64)
	case int:
		return intToUint64(int64(x))
	case int8:
		return intToUint64(int64(x))
	case int16:
		return intToUint64(int64(x))
	case int32:
		return intToUint64(int64(x))
	case int64:
		return intToUint64(x)
	case uint:
		return uint64(x), nil
	case uint8:
		return uint64(x), nil
	case uint16:
		return uint64(x), nil
	case uint32:
		return uint64(x), nil
	case uint64:
		return x, nil
	case float32:
		return floatToUint64(float64(x))
	case float64:
		return floatToUint64(x)
	default:
		return toUint64(reflect.ValueOf(v))
	}
}

func intToUint64(i int64) (uint64, error) {
	if i < 0 {
		return 0, fmt.Errorf("value %d is negative and cannot fit an unsigned type", i)
	}
	return uint64(i), nil
}

func floatToUint64(f float64) (uint64, error) {
	if err := checkIntegralFloat(f); err != nil {
		return 0, err
	}
	if f < 0 {
		return 0, fmt.Errorf("value %v is negative and cannot fit an unsigned type", f)
	}
	if f >= math.MaxUint64 {
		return 0, fmt.Errorf("value %v overflows uint64", f)
	}
	return uint64(f), nil
}

// float64Of mirrors toFloat64 for an any source, parsing strings at the given
// bit size so out-of-range float32 literals are rejected at parse time.
func float64Of(v any, bits int) (float64, error) {
	switch x := v.(type) {
	case string:
		return strconv.ParseFloat(x, bits)
	case int:
		return float64(x), nil
	case int8:
		return float64(x), nil
	case int16:
		return float64(x), nil
	case int32:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case uint:
		return float64(x), nil
	case uint8:
		return float64(x), nil
	case uint16:
		return float64(x), nil
	case uint32:
		return float64(x), nil
	case uint64:
		return float64(x), nil
	case float32:
		return float64(x), nil
	case float64:
		return x, nil
	default:
		return toFloat64(reflect.ValueOf(v), bits)
	}
}

// signed converts v into a signed integer of the width given by destType,
// applying the same overflow check as the reflective toPrimitive.
func signed(v any, destType reflect.Type) (int64, error) {
	n, err := int64Of(v)
	if err != nil {
		return 0, conversionError(reflect.ValueOf(v), destType, err)
	}
	out := reflect.New(destType).Elem()
	if out.OverflowInt(n) {
		return 0, conversionError(reflect.ValueOf(v), destType, fmt.Errorf("value %d overflows %s", n, destType))
	}
	return n, nil
}

func unsigned(v any, destType reflect.Type) (uint64, error) {
	n, err := uint64Of(v)
	if err != nil {
		return 0, conversionError(reflect.ValueOf(v), destType, err)
	}
	out := reflect.New(destType).Elem()
	if out.OverflowUint(n) {
		return 0, conversionError(reflect.ValueOf(v), destType, fmt.Errorf("value %d overflows %s", n, destType))
	}
	return n, nil
}

// Int converts v into an int with reflective-equivalent rules and errors.
func Int(v any) (int, error) { n, err := signed(v, intType); return int(n), err }

// Int8 converts v into an int8.
func Int8(v any) (int8, error) { n, err := signed(v, int8Type); return int8(n), err }

// Int16 converts v into an int16.
func Int16(v any) (int16, error) { n, err := signed(v, int16Type); return int16(n), err }

// Int32 converts v into an int32.
func Int32(v any) (int32, error) { n, err := signed(v, int32Type); return int32(n), err }

// Int64 converts v into an int64.
func Int64(v any) (int64, error) { n, err := signed(v, int64Type); return n, err }

// Uint converts v into a uint.
func Uint(v any) (uint, error) { n, err := unsigned(v, uintType); return uint(n), err }

// Uint8 converts v into a uint8.
func Uint8(v any) (uint8, error) { n, err := unsigned(v, uint8Type); return uint8(n), err }

// Uint16 converts v into a uint16.
func Uint16(v any) (uint16, error) { n, err := unsigned(v, uint16Type); return uint16(n), err }

// Uint32 converts v into a uint32.
func Uint32(v any) (uint32, error) { n, err := unsigned(v, uint32Type); return uint32(n), err }

// Uint64 converts v into a uint64.
func Uint64(v any) (uint64, error) { n, err := unsigned(v, uint64Type); return n, err }

// Float32 converts v into a float32, rejecting out-of-range values.
func Float32(v any) (float32, error) {
	f, err := float64Of(v, 32)
	if err != nil {
		return 0, conversionError(reflect.ValueOf(v), float32Type, err)
	}
	out := reflect.New(float32Type).Elem()
	if out.OverflowFloat(f) {
		return 0, conversionError(reflect.ValueOf(v), float32Type, fmt.Errorf("value %v overflows %s", f, float32Type))
	}
	return float32(f), nil
}

// Float64 converts v into a float64.
func Float64(v any) (float64, error) {
	f, err := float64Of(v, 64)
	if err != nil {
		return 0, conversionError(reflect.ValueOf(v), float64Type, err)
	}
	return f, nil
}

// String converts v into a string, mirroring toStringValue (numbers and bools
// are formatted, strings pass through).
func String(v any) (string, error) {
	switch x := v.(type) {
	case string:
		return x, nil
	case bool:
		return strconv.FormatBool(x), nil
	case int:
		return strconv.FormatInt(int64(x), 10), nil
	case int8:
		return strconv.FormatInt(int64(x), 10), nil
	case int16:
		return strconv.FormatInt(int64(x), 10), nil
	case int32:
		return strconv.FormatInt(int64(x), 10), nil
	case int64:
		return strconv.FormatInt(x, 10), nil
	case uint:
		return strconv.FormatUint(uint64(x), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(x), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(x), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(x), 10), nil
	case uint64:
		return strconv.FormatUint(x, 10), nil
	case float32:
		return strconv.FormatFloat(float64(x), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), nil
	default:
		s, err := toStringValue(reflect.ValueOf(v))
		if err != nil {
			return "", conversionError(reflect.ValueOf(v), stringType, err)
		}
		return s, nil
	}
}

// Bool converts v into a bool, mirroring toBoolValue.
func Bool(v any) (bool, error) {
	switch x := v.(type) {
	case bool:
		return x, nil
	case string:
		b, err := strconv.ParseBool(x)
		if err != nil {
			return false, conversionError(reflect.ValueOf(v), boolType, err)
		}
		return b, nil
	default:
		b, err := toBoolValue(reflect.ValueOf(v))
		if err != nil {
			return false, conversionError(reflect.ValueOf(v), boolType, err)
		}
		return b, nil
	}
}

// Duration converts v into a time.Duration, matching Value's special case: a
// string is parsed as a human duration ("1h30m"), any other value is treated as
// an integer nanosecond count.
func Duration(v any) (time.Duration, error) {
	if s, ok := asString(v); ok {
		d, err := time.ParseDuration(s)
		if err != nil {
			return 0, conversionError(reflect.ValueOf(v), durationType, err)
		}
		return d, nil
	}
	n, err := int64Of(v)
	if err != nil {
		return 0, conversionError(reflect.ValueOf(v), durationType, err)
	}
	return time.Duration(n), nil
}

// StringSlice converts v into a []string. Flat string sources (env, default)
// already hand back a []string (split on the field separator); object sources
// (JSON, YAML) hand back a []any whose elements are converted with String.
// Anything else falls back to the reflective slice conversion.
func StringSlice(v any) ([]string, error) {
	switch xs := v.(type) {
	case []string:
		return xs, nil
	case []any:
		out := make([]string, len(xs))
		for i, e := range xs {
			s, err := String(e)
			if err != nil {
				return nil, err
			}
			out[i] = s
		}
		return out, nil
	default:
		rv, err := Value(reflect.ValueOf(v), stringSlice)
		if err != nil {
			return nil, err
		}
		return rv.Interface().([]string), nil
	}
}
