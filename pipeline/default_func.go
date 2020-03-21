package pipeline

import (
	"fmt"
	"reflect"

	"github.com/goreflect/gostructor/converters"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/middlewares"
	"github.com/goreflect/gostructor/tags"
	"github.com/goreflect/gostructor/tools"
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
	if err := middlewares.ExecutorMiddlewaresByTagValue(value, tags.TagDefault); err != nil {
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	array, err := tools.ConvertStringIntoArray(value, tools.ConfigureConverts{Separator: tools.COMMA})
	if err != nil {
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	return converters.ConvertBetweenComplexTypes(reflect.ValueOf(array), valueIndirect)
}

/*
GetBaseType - get base type from default values.
*/
func (config DefaultConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Message: default values sources start")
	valueIndirect := reflect.Indirect(context.Value)
	value := context.StructField.Tag.Get(tags.TagDefault)
	if err := middlewares.ExecutorMiddlewaresByTagValue(value, tags.TagDefault); err != nil {
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	return converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(value), valueIndirect)
}
