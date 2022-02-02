package converters

import (
	"reflect"

	"github.com/goreflect/gostructor/infra"
)

func convertSlice(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	destResult := reflect.MakeSlice(destination.Type(), source.Len(), source.Cap())
	for i := 0; i < source.Len(); i++ {
		convertedValue := ConvertBetweenPrimitiveTypes(reflect.ValueOf(source.Index(i).Interface().(string)), destResult.Index(i))
		if convertedValue.GetNotAValue() != nil {
			return convertedValue
		}
		destResult.Index(i).Set(convertedValue.Value)
	}
	return infra.NewGoStructorTrueValue(destResult)
}

// func convertMap(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
// 	makebleMap := reflect.MakeMapWithSize(destination.Type(), 0)
// }
