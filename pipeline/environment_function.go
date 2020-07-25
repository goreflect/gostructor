package pipeline

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/goreflect/gostructor/converters"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/middlewares"
	"github.com/goreflect/gostructor/tags"
	"github.com/goreflect/gostructor/tools"
)

type EnvironmentConfig struct {
}

/*
GetComplexType - getting complex types like slices from environment variable
*/
func (config EnvironmentConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	valueIndirect := reflect.Indirect(context.Value)
	switch valueIndirect.Kind() {
	case reflect.Slice:
		valueTag := context.StructField.Tag.Get(tags.TagEnvironment)
		if err := middlewares.ExecutorMiddlewaresByTagValue(valueTag, tags.TagEnvironment); err != nil {
			return infra.NewGoStructorNoValue(context.Value, errors.New("can not checks by middlewares. err: "+err.Error()))
		}

		value := os.Getenv(valueTag)
		// add here additional logic for middlewares and other
		array, errConverting := tools.ConvertStringIntoArray(value, tools.ConfigureConverts{
			Separator: tools.COMMA,
		})
		if errConverting != nil {
			return infra.NewGoStructorNoValue(context.Value, errConverting)
		}
		return converters.ConvertBetweenComplexTypes(reflect.ValueOf(array), valueIndirect)
	default:
		return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("complex type "+valueIndirect.Kind().String()+" not implemented in environment parsing function"))
	}
}

/*
GetBaseType - getting base type values
*/
func (config EnvironmentConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Environment values sources start")
	valueIndirect := reflect.Indirect(context.Value)
	valueTag := context.StructField.Tag.Get(tags.TagEnvironment)
	if valueTag != "" {
		switch valueIndirect.Kind() {
		case reflect.String:
			value := os.Getenv(valueTag)
			if config.checksByMiddlewares(value) {
				infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getBaseType can not get empty value from environment by key: "+valueTag))
			}
			return infra.NewGoStructorTrueValue(reflect.ValueOf(value))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			value := os.Getenv(valueTag)
			if config.checksByMiddlewares(value) {
				infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getBaseType can not get empty value from environment by key: "+valueTag))
			}
			parsing, errParsing := strconv.ParseInt(value, 10, 64)
			if errParsing != nil {
				infra.NewGoStructorNoValue(context.Value.Interface(), errParsing)
			}
			return infra.NewGoStructorTrueValue(reflect.ValueOf(parsing).Convert(valueIndirect.Type()))
		case reflect.Float32, reflect.Float64:
			value := os.Getenv(valueTag)
			if config.checksByMiddlewares(value) {
				infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getBaseType can not get empty value from environment by key: "+valueTag))
			}
			parsing, errParsing := strconv.ParseFloat(value, 64)
			if errParsing != nil {
				infra.NewGoStructorNoValue(context.Value.Interface(), errParsing)
			}
			return infra.NewGoStructorTrueValue(reflect.ValueOf(parsing).Convert(valueIndirect.Type()))
		case reflect.Bool:
			value := os.Getenv(valueTag)
			if config.checksByMiddlewares(value) {
				infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getBaseType can not get empty value from environment by key: "+valueTag))
			}
			parsing, errParsing := strconv.ParseBool(value)
			if errParsing != nil {
				infra.NewGoStructorNoValue(context.Value.Interface(), errParsing)
			}
			return infra.NewGoStructorTrueValue(reflect.ValueOf(parsing))
		default:
			return infra.NewGoStructorNoValue(valueIndirect.Interface(), errors.New("can not recognized type of you variable"))
		}
	}
	return infra.NewGoStructorNoValue(context.Value, errors.New("getBaseType can not get field by empty tag value of tag: "+tags.TagEnvironment))
}

func (config EnvironmentConfig) checksByMiddlewares(value string) bool {
	return value == ""
}
