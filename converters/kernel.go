package converters

import (
	"errors"
	"reflect"

	"github.com/goreflect/gostructor/infra"
	logrus "github.com/sirupsen/logrus"
)

/*ConvertBetweenPrimitiveTypes - method for converting from any of base types into any of base types,
like string, bool, int, int8, int16. int32, int64
*/
func ConvertBetweenPrimitiveTypes(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	logrus.Debug("start converting source: ", source.Kind().String(), " destination: ", destination.Kind().String())
	switch destination.Kind() {
	case reflect.Int:
		return convertToInt(source, destination)
	case reflect.Int8:
		return convertToIntOrder(source, destination, 8)
	case reflect.Int16:
		return convertToIntOrder(source, destination, 16)
	case reflect.Int32:
		return convertToIntOrder(source, destination, 32)
	case reflect.Int64:
		return convertToIntOrder(source, destination, 64)
	case reflect.Uint:
		return convertToUintOrder(source, destination, 32)
	case reflect.Uint8:
		return convertToUintOrder(source, destination, 8)
	case reflect.Uint16:
		return convertToUintOrder(source, destination, 16)
	case reflect.Uint32:
		return convertToUintOrder(source, destination, 32)
	case reflect.Uint64:
		return convertToUintOrder(source, destination, 64)
	case reflect.String:
		return convertToString(source, destination)
	case reflect.Float32:
		return convertToFloatOrder(source, destination, 32)
	case reflect.Float64:
		return convertToFloatOrder(source, destination, 64)
	case reflect.Bool:
		return convertToBool(source, destination)
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted to this type "+destination.Kind().String()+" because this type not supported"))
	}
}

/*
ConvertBetweenComplexTypes - converting between complex types like slice to slice, map to map
*/
func ConvertBetweenComplexTypes(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	switch destination.Kind() {
	case reflect.Slice, reflect.Array:
		return convertSlice(source, destination)
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("not implemented"))
	}
}
