package pipeline

import (
	"errors"
	"os"
	"reflect"

	"github.com/goreflect/gostructor/converters"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/middlewares"
	"github.com/goreflect/gostructor/tags"
	"github.com/goreflect/gostructor/tools"
	"github.com/sirupsen/logrus"
)

/*EnvironmentConfig - configuring structures from environment*/
type EnvironmentConfig struct {
}

/*
GetComplexType - getting complex types like slices from environment variable
*/
func (config EnvironmentConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	logrus.Debug("get values from environment start")
	valueIndirect := reflect.Indirect(context.Value)
	valueTag := context.StructField.Tag.Get(tags.TagEnvironment)
	if config.checkTagValue(valueTag) {
		return infra.NewGoStructorNoValue(context.Value, errors.New("cf_env tag value for field '"+context.StructField.Name+"' is empty"))
	}
	value := os.Getenv(valueTag)
	if err := middlewares.ExecutorMiddlewaresByTagValue(value, tags.TagEnvironment); err != nil {
		return infra.NewGoStructorNoValue(context.Value, errors.New("middleware check failed: "+err.Error()))
	}
	array, err := tools.ConvertStringIntoArray(value, tools.ConfigureConverts{Separator: tools.COMMA})
	if err != nil {
		return infra.NewGoStructorNoValue(context.Value.Interface(), err)
	}
	return converters.ConvertBetweenComplexTypes(reflect.ValueOf(array), valueIndirect)
}

/*
GetBaseType - getting base type values
*/
func (config EnvironmentConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	logrus.Debug("get values from environment start")
	valueIndirect := reflect.Indirect(context.Value)
	valueTag := context.StructField.Tag.Get(tags.TagEnvironment)

	if !config.checkTagValue(valueTag) {
		value := os.Getenv(valueTag)
		if value == "" {
			return infra.NewGoStructorNoValue(context.Value, errors.New("environment variable '"+valueTag+"' was not set or was empty"))
		}
		return converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(value), valueIndirect)
	}
	return infra.NewGoStructorNoValue(context.Value, errors.New("cf_env tag value for field '"+context.StructField.Name+"' is empty"))
}

func (config EnvironmentConfig) checkTagValue(tagvalue string) bool {
	// in the future in this case will be added a call middlewares functions
	return tagvalue == ""
}
