package converters

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

func Convert(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	fmt.Println("start converting source: ", source.Kind().String(), " destination: ", destination.Kind().String())
	switch destination.Kind() {
	case reflect.Int16, reflect.Int32, reflect.Int64:
		return convertsToInt(source, destination)
	case reflect.String:
		return source, nil
	case reflect.Bool:
		return convertToBool(source, destination)
	}
	return reflect.Zero(nil), errors.New("not setuped")
}

func convertsToInt(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	switch source.Kind() {
	case reflect.String:
		convertByReflection := source.Convert(destination.Type())
		fmt.Println("Convert by reflection: ", convertByReflection.Kind(), " can set?:", convertByReflection.CanSet())
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, 64)
		if errorConvert != nil {
			return reflect.Zero(nil), errorConvert
		}
		return reflect.ValueOf(convertedValue), nil
	case reflect.Array, reflect.Bool, reflect.Slice, reflect.Map:
		return reflect.Zero(nil), errors.New("can not be convert from bool to " + source.Kind().String() + " type")
	default:
		return reflect.Zero(nil), errors.New("unhandled type for converts")
	}
}

func convertToBool(source reflect.Value, destination reflect.Value) (reflect.Value, error) {
	return reflect.Zero(nil), errors.New("not setup")
}
