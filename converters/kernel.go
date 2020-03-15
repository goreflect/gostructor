package converters

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/goreflect/gostructor/pipeline"
)

/*ConvertBetweenPrimitiveTypes - method for converting from any of base types into any of base types,
like string, bool, int, int8, int16. int32, int64
*/
func ConvertBetweenPrimitiveTypes(source reflect.Value, destination reflect.Value) pipeline.GoStructorValue {
	fmt.Println("Level: Debug. Message: start converting source: ", source.Kind().String(), " destination: ", destination.Kind().String())
	switch destination.Kind() {
	case reflect.Int:
		return convertToInt(source, destination)
	case reflect.Int8:
		return convertToInt8(source, destination)
	case reflect.Int16:
		return convertToInt16(source, destination)
	case reflect.Int32:
		return convertToInt32(source, destination)
	case reflect.Int64:
		return convertToInt64(source, destination)
	case reflect.String:
		return convertToString(source, destination)
	case reflect.Bool:
		return convertToBool(source, destination)
	default:
		return pipeline.NewGoStructorNoValue(destination, errors.New("can not converted to this type "+destination.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToInt(source reflect.Value, destination reflect.Value) pipeline.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int type")
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, 64)
		if errorConvert != nil {
			return pipeline.NewGoStructorNoValue(destination, errors.New("can not converted to this type: "+destination.Kind().String()))
		}
		return pipeline.NewGoStructorTrueValue(reflect.ValueOf(int(convertedValue)))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int type")
		if source.Type().ConvertibleTo(destination.Type()) {
			return pipeline.NewGoStructorTrueValue(source.Convert(destination.Type()))
		}
		return pipeline.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into int"))
	default:
		return pipeline.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToInt8(source reflect.Value, destination reflect.Value) pipeline.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int8 type")
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, 8)
		if errorConvert != nil {
			return pipeline.NewGoStructorNoValue(destination, errors.New("can not converted to this type: "+destination.Kind().String()))
		}
		return pipeline.NewGoStructorTrueValue(reflect.ValueOf(int8(convertedValue)))
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int8 type")
		if source.Type().ConvertibleTo(destination.Type()) {
			return pipeline.NewGoStructorTrueValue(source.Convert(destination.Type()))
		}
		return pipeline.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into int8"))
	default:
		return pipeline.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToInt16(source reflect.Value, destination reflect.Value) pipeline.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int16 type")
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, 16)
		if errorConvert != nil {
			return pipeline.NewGoStructorNoValue(destination, errors.New("can not converted to this type: "+destination.Kind().String()))
		}
		return pipeline.NewGoStructorTrueValue(reflect.ValueOf(int16(convertedValue)))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int16 type")
		if source.Type().ConvertibleTo(destination.Type()) {
			return pipeline.NewGoStructorTrueValue(source.Convert(destination.Type()))
		}
		return pipeline.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into int16"))
	default:
		return pipeline.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToInt32(source reflect.Value, destination reflect.Value) pipeline.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int32 type")
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, 32)
		if errorConvert != nil {
			return pipeline.NewGoStructorNoValue(destination, errors.New("can not converted to this type: "+destination.Kind().String()))
		}
		return pipeline.NewGoStructorTrueValue(reflect.ValueOf(int32(convertedValue)))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int32 type")
		if source.Type().ConvertibleTo(destination.Type()) {
			return pipeline.NewGoStructorTrueValue(source.Convert(destination.Type()))
		}
		return pipeline.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into int32"))
	default:
		return pipeline.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToInt64(source reflect.Value, destination reflect.Value) pipeline.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int64 type")
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, 64)
		if errorConvert != nil {
			return pipeline.NewGoStructorNoValue(destination, errors.New("can not converted to this type: "+destination.Kind().String()))
		}
		return pipeline.NewGoStructorTrueValue(reflect.ValueOf(int64(convertedValue)))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Println("Level: Debug. Message: start convert value ", source.String(), " into int64 type")
		if source.Type().ConvertibleTo(destination.Type()) {
			return pipeline.NewGoStructorTrueValue(source.Convert(destination.Type()))
		}
		return pipeline.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into int64"))
	default:
		return pipeline.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToBool(source reflect.Value, destination reflect.Value) pipeline.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		converted, errConvert := strconv.ParseBool(source.String())
		if errConvert != nil {
			return pipeline.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into bool"))
		}
		return pipeline.NewGoStructorTrueValue(reflect.ValueOf(converted))
	case reflect.Bool:
		return pipeline.NewGoStructorTrueValue(source)
	default:
		return pipeline.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}

func convertToString(source reflect.Value, destination reflect.Value) pipeline.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		return pipeline.NewGoStructorTrueValue(source)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return pipeline.NewGoStructorTrueValue(reflect.ValueOf(
			strconv.FormatInt(source.Int(), 10),
		))
	case reflect.Float32:
		return pipeline.NewGoStructorTrueValue(reflect.ValueOf(
			strconv.FormatFloat(source.Float(), 'E', -1, 32),
		))
	case reflect.Float64:
		return pipeline.NewGoStructorTrueValue(reflect.ValueOf(
			strconv.FormatFloat(source.Float(), 'E', -1, 64),
		))
	case reflect.Bool:
		return pipeline.NewGoStructorTrueValue(reflect.ValueOf(
			strconv.FormatBool(source.Bool()),
		))
	default:
		return pipeline.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+source.Kind().String()+" beacuse this type not supported"))
	}
}
