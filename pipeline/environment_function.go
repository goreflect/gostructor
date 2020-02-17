package pipeline

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/goreflect/gostructor/tags"
)

type EnvironmentConfig struct {
}

func (config EnvironmentConfig) GetComplexType(context *structContext) GoStructorValue {
	valueIndirect := reflect.Indirect(context.Value)
	switch valueIndirect.Kind() {
	case reflect.Slice:
	case reflect.Map:
	case reflect.Array:
	default:
		return config.GetBaseType(context)
	}
	return NewGoStructorNoValue(context.Value.Interface(), errors.New("getComplexType not implement for environment configuring"))
}

func (config EnvironmentConfig) GetBaseType(context *structContext) GoStructorValue {
	fmt.Println("Level: Debug. nvironment values sources start")
	valueIndirect := reflect.Indirect(context.Value)
	valueTag := context.StructField.Tag.Get(tags.TagEnvironment)
	if valueTag != "" {
		switch valueIndirect.Kind() {
		case reflect.String:
			value := os.Getenv(valueTag)
			if config.checksByMiddlewares(value) {
				NewGoStructorNoValue(context.Value.Interface(), errors.New("getBaseType can not get empty value from environment by key: "+valueTag))
			}
			return NewGoStructorTrueValue(reflect.ValueOf(value))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			value := os.Getenv(valueTag)
			if config.checksByMiddlewares(value) {
				NewGoStructorNoValue(context.Value.Interface(), errors.New("getBaseType can not get empty value from environment by key: "+valueTag))
			}
			parsing, errParsing := strconv.ParseInt(value, 10, 64)
			if errParsing != nil {
				NewGoStructorNoValue(context.Value.Interface(), errParsing)
			}
			return NewGoStructorTrueValue(reflect.ValueOf(parsing).Convert(valueIndirect.Type()))
		case reflect.Float32, reflect.Float64:
			value := os.Getenv(valueTag)
			if config.checksByMiddlewares(value) {
				NewGoStructorNoValue(context.Value.Interface(), errors.New("getBaseType can not get empty value from environment by key: "+valueTag))
			}
			parsing, errParsing := strconv.ParseFloat(value, 64)
			if errParsing != nil {
				NewGoStructorNoValue(context.Value.Interface(), errParsing)
			}
			return NewGoStructorTrueValue(reflect.ValueOf(parsing).Convert(valueIndirect.Type()))
		case reflect.Bool:
			value := os.Getenv(valueTag)
		}
	}
	return NewGoStructorNoValue(context.Value.Interface(), errors.New("getBaseType can not getting field by empty tag value of tag: "+tags.TagEnvironment))
}

func (config EnvironmentConfig) checksByMiddlewares(tagvalue string) bool {
	// in the future in this case will be added a call middlewares functions
	return tagvalue == ""
}
