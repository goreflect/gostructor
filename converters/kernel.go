package converters

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

/*ConvertBetweenPrimitiveTypes - method for converting from any of base types into any of base types,
// like string, bool, int, int8, int16. int32, int64
*/
func ConvertBetweenPrimitiveTypes(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
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
		return reflect.Zero(nil), errors.New("can not converted to this type: " + destination.Kind().String())
	}
}

func convertToInt(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	switch source.Kind() {
	case reflect.String:
		convertByReflection := source.String()
		fmt.Println("Level: Debug. Message: start convert value ", convertByReflection, " into int type")

		// fmt.Println("Level: Debug. Message: Convert by reflection: ", convertByReflection.Kind(), " can set?:", convertByReflection.CanSet())
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, 64)
		if errorConvert != nil {
			return reflect.ValueOf(-1), errorConvert
		}
		return reflect.ValueOf(int(convertedValue)), nil
	case reflect.Int64:
		convertByreflection := source.Convert(destination.Type())
		fmt.Println("LogLevel: debug. Message: Convert by reflection: ", convertByreflection.Kind(), " can set?:", convertByreflection.CanSet())
		return reflect.Zero(nil), errors.New("not implemented")
	default:
		return reflect.Zero(nil), errors.New("cannot convert " + source.Kind().String() + " to " + reflect.Int.String())
	}
}

func convertToInt8(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	return reflect.Zero(nil), errors.New("not implemented")
}

func convertToInt16(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	return reflect.Zero(nil), errors.New("not implemented")
}

func convertToInt32(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	return reflect.Zero(nil), errors.New("not implemented")
}

func convertToInt64(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	return reflect.Zero(nil), errors.New("not implemented")
}

func convertToBool(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	return reflect.Zero(nil), errors.New("not implemented")
}

func convertToString(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	return reflect.Zero(nil), errors.New("not implemented")
}
