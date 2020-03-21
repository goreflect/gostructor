package pipeline

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/goreflect/gostructor/converters"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
)

/*
DefaultConfig - one of most configuring source functions that should preparing fields in structures of data by default values setuped by special tag name
*/
type DefaultConfig struct {
}

/*
GetComplexType - get slices, maps, arrays or anything else hard types
*/
func (config DefaultConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Message: default values sources start")
	valueIndirect := reflect.Indirect(context.Value)
	value := context.StructField.Tag.Get(tags.TagDefault)
	if config.checkNotRightValue(value) {
		// TODO: increase message by information about what wrong in future issues
		return infra.NewGoStructorNoValue(context.Value, errors.New("wrong format inside the tag"))
	}
	return converters.ConvertBetweenComplexTypes(reflect.ValueOf(value), valueIndirect)
}

// this is main entrypoint for checking value in tag
func (config DefaultConfig) checkNotRightValue(value string) bool {
	return value == ""
}

/*
GetBaseType - get base type from default values.
*/
func (config DefaultConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Message: default values sources start")
	valueIndirect := reflect.Indirect(context.Value)
	value := context.StructField.Tag.Get(tags.TagDefault)
	if config.checkNotRightValue(value) {
		// TODO: increase message by information about what wrong in future issues
		return infra.NewGoStructorNoValue(context.Value, errors.New("wrong format inside the tag"))
	}
	return converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(value), valueIndirect)
}
