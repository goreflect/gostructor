package converters

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/goreflect/gostructor/infra"
)

func convertToInt(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int type")
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, 64)
		if errorConvert != nil {
			return infra.NewGoStructorNoValue(destination, errors.New("can not converted to this type: "+destination.Kind().String()))
		}
		return infra.NewGoStructorTrueValue(reflect.ValueOf(int(convertedValue)))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int type")
		if source.Type().ConvertibleTo(destination.Type()) {
			return infra.NewGoStructorTrueValue(source.Convert(destination.Type()))
		}
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into int"))
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToInt8(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int8 type")
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, 8)
		if errorConvert != nil {
			return infra.NewGoStructorNoValue(destination, errors.New("can not converted to this type: "+destination.Kind().String()))
		}
		return infra.NewGoStructorTrueValue(reflect.ValueOf(int8(convertedValue)))
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int8 type")
		if source.Type().ConvertibleTo(destination.Type()) {
			return infra.NewGoStructorTrueValue(source.Convert(destination.Type()))
		}
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into int8"))
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToInt16(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int16 type")
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, 16)
		if errorConvert != nil {
			return infra.NewGoStructorNoValue(destination, errors.New("can not converted to this type: "+destination.Kind().String()))
		}
		return infra.NewGoStructorTrueValue(reflect.ValueOf(int16(convertedValue)))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int16 type")
		if source.Type().ConvertibleTo(destination.Type()) {
			return infra.NewGoStructorTrueValue(source.Convert(destination.Type()))
		}
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into int16"))
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToInt32(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int32 type")
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, 32)
		if errorConvert != nil {
			return infra.NewGoStructorNoValue(destination, errors.New("can not converted to this type: "+destination.Kind().String()))
		}
		return infra.NewGoStructorTrueValue(reflect.ValueOf(int32(convertedValue)))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int32 type")
		if source.Type().ConvertibleTo(destination.Type()) {
			return infra.NewGoStructorTrueValue(source.Convert(destination.Type()))
		}
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into int32"))
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToInt64(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int64 type")
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, 64)
		if errorConvert != nil {
			return infra.NewGoStructorNoValue(destination, errors.New("can not converted to this type: "+destination.Kind().String()))
		}
		return infra.NewGoStructorTrueValue(reflect.ValueOf(int64(convertedValue)))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int64 type")
		if source.Type().ConvertibleTo(destination.Type()) {
			return infra.NewGoStructorTrueValue(source.Convert(destination.Type()))
		}
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into int64"))
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToBool(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		converted, errConvert := strconv.ParseBool(source.String())
		if errConvert != nil {
			return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into bool"))
		}
		return infra.NewGoStructorTrueValue(reflect.ValueOf(converted))
	case reflect.Bool:
		return infra.NewGoStructorTrueValue(source)
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToString(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		return infra.NewGoStructorTrueValue(source)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return infra.NewGoStructorTrueValue(reflect.ValueOf(
			strconv.FormatInt(source.Int(), 10),
		))
	case reflect.Float32:
		return infra.NewGoStructorTrueValue(reflect.ValueOf(
			strconv.FormatFloat(source.Float(), 'E', -1, 32),
		))
	case reflect.Float64:
		return infra.NewGoStructorTrueValue(reflect.ValueOf(
			strconv.FormatFloat(source.Float(), 'E', -1, 64),
		))
	case reflect.Bool:
		return infra.NewGoStructorTrueValue(reflect.ValueOf(
			strconv.FormatBool(source.Bool()),
		))
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToFloat32(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return infra.NewGoStructorTrueValue(reflect.ValueOf(float32(source.Int())))
	case reflect.String:
		parsed, errParsed := strconv.ParseFloat(source.String(), 32)
		if errParsed != nil {
			return infra.NewGoStructorNoValue(destination, errParsed)
		}
		return infra.NewGoStructorTrueValue(reflect.ValueOf(float32(parsed)))
	case reflect.Float32:
		return infra.NewGoStructorTrueValue(reflect.ValueOf(float32(source.Float())))
	case reflect.Float64:
		return infra.NewGoStructorNoValue(destination, errors.New("can not convert from float64 into float32"))
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToFloat64(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return infra.NewGoStructorTrueValue(reflect.ValueOf(float64(source.Int())))
	case reflect.String:
		parsed, errParsed := strconv.ParseFloat(source.String(), 64)
		if errParsed != nil {
			return infra.NewGoStructorNoValue(destination, errParsed)
		}
		return infra.NewGoStructorTrueValue(reflect.ValueOf(parsed))
	case reflect.Float32, reflect.Float64:
		return infra.NewGoStructorTrueValue(reflect.ValueOf(source.Float()))
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}
