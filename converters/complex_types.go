package converters

import (
	"errors"
	"reflect"

	"github.com/goreflect/gostructor/infra"
)

func convertSlice(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	if source.Kind() != reflect.Slice && source.Kind() != reflect.Array {
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into slice"))
	}
	destResult := reflect.MakeSlice(destination.Type(), source.Len(), source.Len())
	for i := 0; i < source.Len(); i++ {
		elementValue := reflect.ValueOf(source.Index(i).Interface())
		convertedValue := ConvertBetweenPrimitiveTypes(elementValue, destResult.Index(i))
		if convertedValue.GetNotAValue() != nil {
			return convertedValue
		}
		destResult.Index(i).Set(convertedValue.Value)
	}
	return infra.NewGoStructorTrueValue(destResult)
}

func convertMap(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	if source.Kind() != reflect.Map {
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into map"))
	}
	destType := destination.Type()
	result := reflect.MakeMapWithSize(destType, source.Len())
	keyType := destType.Key()
	valueType := destType.Elem()
	for _, key := range source.MapKeys() {
		convertedKey := ConvertBetweenPrimitiveTypes(reflect.ValueOf(key.Interface()), reflect.New(keyType).Elem())
		if convertedKey.GetNotAValue() != nil {
			return convertedKey
		}
		convertedValue := ConvertBetweenPrimitiveTypes(reflect.ValueOf(source.MapIndex(key).Interface()), reflect.New(valueType).Elem())
		if convertedValue.GetNotAValue() != nil {
			return convertedValue
		}
		result.SetMapIndex(convertedKey.Value, convertedValue.Value)
	}
	return infra.NewGoStructorTrueValue(result)
}
