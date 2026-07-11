package convert

import (
	"reflect"
	"strconv"
)

func toInt(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	switch source.Kind() {
	case reflect.String:
		parsed, err := strconv.ParseInt(source.String(), 10, 64)
		if err != nil {
			return reflect.Value{}, unsupportedSource(source, "int")
		}
		return reflect.ValueOf(int(parsed)), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(int(source.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(int(source.Uint())), nil
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(int(source.Float())), nil
	default:
		return reflect.Value{}, unsupportedSource(source, "int")
	}
}

func toIntSized(source reflect.Value, destination reflect.Value, bits int) (reflect.Value, error) {
	var value int64
	switch source.Kind() {
	case reflect.String:
		parsed, err := strconv.ParseInt(source.String(), 10, bits)
		if err != nil {
			return reflect.Value{}, unsupportedSource(source, "int"+strconv.Itoa(bits))
		}
		value = parsed
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = source.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = int64(source.Uint())
	case reflect.Float32, reflect.Float64:
		value = int64(source.Float())
	default:
		return reflect.Value{}, unsupportedSource(source, "int"+strconv.Itoa(bits))
	}
	return sizedInt(value, bits), nil
}

func sizedInt(value int64, bits int) reflect.Value {
	switch bits {
	case 8:
		return reflect.ValueOf(int8(value))
	case 16:
		return reflect.ValueOf(int16(value))
	case 32:
		return reflect.ValueOf(int32(value))
	default:
		return reflect.ValueOf(value)
	}
}

func toUint(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	switch source.Kind() {
	case reflect.String:
		parsed, err := strconv.ParseUint(source.String(), 10, 64)
		if err != nil {
			return reflect.Value{}, unsupportedSource(source, "uint")
		}
		return reflect.ValueOf(uint(parsed)), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(uint(source.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(uint(source.Uint())), nil
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(uint(source.Float())), nil
	default:
		return reflect.Value{}, unsupportedSource(source, "uint")
	}
}

func toUintSized(source reflect.Value, destination reflect.Value, bits int) (reflect.Value, error) {
	var value uint64
	switch source.Kind() {
	case reflect.String:
		parsed, err := strconv.ParseUint(source.String(), 10, bits)
		if err != nil {
			return reflect.Value{}, unsupportedSource(source, "uint"+strconv.Itoa(bits))
		}
		value = parsed
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = uint64(source.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = source.Uint()
	case reflect.Float32, reflect.Float64:
		value = uint64(source.Float())
	default:
		return reflect.Value{}, unsupportedSource(source, "uint"+strconv.Itoa(bits))
	}
	return sizedUint(value, bits), nil
}

func sizedUint(value uint64, bits int) reflect.Value {
	switch bits {
	case 8:
		return reflect.ValueOf(uint8(value))
	case 16:
		return reflect.ValueOf(uint16(value))
	case 32:
		return reflect.ValueOf(uint32(value))
	default:
		return reflect.ValueOf(value)
	}
}

func toFloatSized(source reflect.Value, destination reflect.Value, bits int) (reflect.Value, error) {
	var value float64
	switch source.Kind() {
	case reflect.String:
		parsed, err := strconv.ParseFloat(source.String(), bits)
		if err != nil {
			return reflect.Value{}, unsupportedSource(source, "float"+strconv.Itoa(bits))
		}
		value = parsed
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = float64(source.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = float64(source.Uint())
	case reflect.Float32, reflect.Float64:
		value = source.Float()
	default:
		return reflect.Value{}, unsupportedSource(source, "float"+strconv.Itoa(bits))
	}
	if bits == 32 {
		return reflect.ValueOf(float32(value)), nil
	}
	return reflect.ValueOf(value), nil
}

func toBool(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	switch source.Kind() {
	case reflect.String:
		parsed, err := strconv.ParseBool(source.String())
		if err != nil {
			return reflect.Value{}, unsupportedSource(source, "bool")
		}
		return reflect.ValueOf(parsed), nil
	case reflect.Bool:
		return source, nil
	default:
		return reflect.Value{}, unsupportedSource(source, "bool")
	}
}

func toString(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	switch source.Kind() {
	case reflect.String:
		return source, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(strconv.FormatInt(source.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(strconv.FormatUint(source.Uint(), 10)), nil
	case reflect.Float32:
		return reflect.ValueOf(strconv.FormatFloat(source.Float(), 'f', -1, 32)), nil
	case reflect.Float64:
		return reflect.ValueOf(strconv.FormatFloat(source.Float(), 'f', -1, 64)), nil
	case reflect.Bool:
		return reflect.ValueOf(strconv.FormatBool(source.Bool())), nil
	default:
		return reflect.Value{}, unsupportedSource(source, "string")
	}
}
