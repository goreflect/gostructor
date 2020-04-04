package pipeline

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/goreflect/gostructor/converters"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
)

/*EnvironmentConfig - configuring structures from environment*/
type EnvironmentConfig struct {
}

const (
	separator = ","
)

/*
GetComplexType - getting complex types like slices from environment variable
*/
func (config EnvironmentConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	valueIndirect := reflect.Indirect(context.Value)
	switch valueIndirect.Kind() {
	case reflect.Slice:
		valueTag := context.StructField.Tag.Get(tags.TagEnvironment)
		if valueTag != "" {
			value := os.Getenv(valueTag)
			// add here additional logic for middlewares and other
			array := config.convertStringIntoArray(value)
			return converters.ConvertBetweenComplexTypes(reflect.ValueOf(array), valueIndirect)
		}
		return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("complex type "+valueIndirect.Kind().String()+" not implemented in environment parsing function"))
	default:
		return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("complex type "+valueIndirect.Kind().String()+" not implemented in environment parsing function"))
	}
}

// TODO: Add variant for change separator. Currently it is comma
func (config EnvironmentConfig) convertStringIntoArray(value string) []string {
	return strings.Split(value, separator)
}

/*
GetBaseType - getting base type values
*/
func (config EnvironmentConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Environment values sources start")
	valueIndirect := reflect.Indirect(context.Value)
	valueTag := context.StructField.Tag.Get(tags.TagEnvironment)

	if valueTag != "" {
		value := os.Getenv(valueTag)
		return converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(value), valueIndirect)
	}
	return infra.NewGoStructorNoValue(context.Value, errors.New("getBaseType can not get field by empty tag value of tag: "+tags.TagEnvironment))
}

// TODO: using in future for run middlewares
// func (config EnvironmentConfig) checksByMiddlewares(tagvalue string) bool {
// 	// in the future in this case will be added a call middlewares functions
// 	return tagvalue == ""
// }
