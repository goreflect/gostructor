package convert

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// toPrimitive converts source into a new value of destType for the scalar
// kinds. The result is constructed via reflect.New(destType) so it carries
// destType's exact (possibly named) type - the reason a `type Level int` or
// a time.Duration field can be Set directly without a further Convert. Every
// numeric conversion is checked: fractional floats, NaN/Inf, sign mismatches,
// and values that overflow the destination width are errors, never silent
// truncation or wraparound.
func toPrimitive(source reflect.Value, destType reflect.Type) (reflect.Value, error) {
	switch destType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := toInt64(source)
		if err != nil {
			return reflect.Value{}, conversionError(source, destType, err)
		}
		out := reflect.New(destType).Elem()
		if out.OverflowInt(v) {
			return reflect.Value{}, conversionError(source, destType, fmt.Errorf("value %d overflows %s", v, destType))
		}
		out.SetInt(v)
		return out, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := toUint64(source)
		if err != nil {
			return reflect.Value{}, conversionError(source, destType, err)
		}
		out := reflect.New(destType).Elem()
		if out.OverflowUint(v) {
			return reflect.Value{}, conversionError(source, destType, fmt.Errorf("value %d overflows %s", v, destType))
		}
		out.SetUint(v)
		return out, nil

	case reflect.Float32, reflect.Float64:
		bits := 64
		if destType.Kind() == reflect.Float32 {
			bits = 32
		}
		v, err := toFloat64(source, bits)
		if err != nil {
			return reflect.Value{}, conversionError(source, destType, err)
		}
		out := reflect.New(destType).Elem()
		if out.OverflowFloat(v) {
			return reflect.Value{}, conversionError(source, destType, fmt.Errorf("value %v overflows %s", v, destType))
		}
		out.SetFloat(v)
		return out, nil

	case reflect.String:
		s, err := toStringValue(source)
		if err != nil {
			return reflect.Value{}, conversionError(source, destType, err)
		}
		out := reflect.New(destType).Elem()
		out.SetString(s)
		return out, nil

	case reflect.Bool:
		b, err := toBoolValue(source)
		if err != nil {
			return reflect.Value{}, conversionError(source, destType, err)
		}
		out := reflect.New(destType).Elem()
		out.SetBool(b)
		return out, nil

	default:
		return reflect.Value{}, fmt.Errorf("convert: destination kind %s is not a primitive type", destType.Kind())
	}
}

// toInt64 extracts an exact int64 from source. Strings are parsed base-10;
// unsigned sources exceeding MaxInt64 and floats that are non-finite,
// fractional, or out of int64 range are rejected.
func toInt64(source reflect.Value) (int64, error) {
	switch source.Kind() {
	case reflect.String:
		return strconv.ParseInt(source.String(), 10, 64)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return source.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u := source.Uint()
		if u > math.MaxInt64 {
			return 0, fmt.Errorf("value %d overflows int64", u)
		}
		return int64(u), nil
	case reflect.Float32, reflect.Float64:
		f := source.Float()
		if err := checkIntegralFloat(f); err != nil {
			return 0, err
		}
		if f < math.MinInt64 || f >= math.MaxInt64 {
			return 0, fmt.Errorf("value %v overflows int64", f)
		}
		return int64(f), nil
	default:
		return 0, errUnsupportedSourceKind
	}
}

// toUint64 extracts an exact uint64 from source, rejecting negative integers,
// negative/non-finite/fractional floats, and floats out of uint64 range.
func toUint64(source reflect.Value) (uint64, error) {
	switch source.Kind() {
	case reflect.String:
		return strconv.ParseUint(source.String(), 10, 64)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i := source.Int()
		if i < 0 {
			return 0, fmt.Errorf("value %d is negative and cannot fit an unsigned type", i)
		}
		return uint64(i), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return source.Uint(), nil
	case reflect.Float32, reflect.Float64:
		f := source.Float()
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
	default:
		return 0, errUnsupportedSourceKind
	}
}

// toFloat64 extracts a float64 from source. Strings are parsed with the
// destination bit size so out-of-range float32 literals are rejected at parse
// time; numeric sources are range-checked by the caller via OverflowFloat.
func toFloat64(source reflect.Value, bits int) (float64, error) {
	switch source.Kind() {
	case reflect.String:
		return strconv.ParseFloat(source.String(), bits)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(source.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(source.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return source.Float(), nil
	default:
		return 0, errUnsupportedSourceKind
	}
}

func checkIntegralFloat(f float64) error {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return fmt.Errorf("value %v is not a finite number", f)
	}
	if math.Trunc(f) != f {
		return fmt.Errorf("value %v is not an integer", f)
	}
	return nil
}

func toBoolValue(source reflect.Value) (bool, error) {
	switch source.Kind() {
	case reflect.String:
		return strconv.ParseBool(source.String())
	case reflect.Bool:
		return source.Bool(), nil
	default:
		return false, errUnsupportedSourceKind
	}
}

func toStringValue(source reflect.Value) (string, error) {
	switch source.Kind() {
	case reflect.String:
		return source.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(source.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(source.Uint(), 10), nil
	case reflect.Float32:
		return strconv.FormatFloat(source.Float(), 'f', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(source.Float(), 'f', -1, 64), nil
	case reflect.Bool:
		return strconv.FormatBool(source.Bool()), nil
	default:
		return "", errUnsupportedSourceKind
	}
}
