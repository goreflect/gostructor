package pipeline

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"

	"github.com/goreflect/gostructor/converters"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
	"github.com/goreflect/gostructor/tools"
	"github.com/sirupsen/logrus"
)

/*JSONConfig - source json configuring*/
type JSONConfig struct {
	fileName   string
	parsedData map[string]interface{}
}

/*GetComplexType - get complex types like arrays, slices, maps from json source*/
func (config JSONConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	logrus.Debug("Level: Debug. Json configurator source start.")
	parsed, notAValue := config.typeSafeLoadConfigFile(context)
	if !parsed {
		return *notAValue
	}
	nameField := context.StructField.Tag.Get(tags.TagJSON)
	if config.validation(nameField) {
		nameField = context.Prefix + context.StructField.Name
	}
	logrus.Debug("Level: Debug. Key for getting values from source: ", nameField)

	parsedValue, found := config.parsedData[nameField]
	if !found || parsedValue == nil {
		return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("value for key '"+nameField+"' was not found in json source"))
	}
	return converters.ConvertBetweenComplexTypes(reflect.ValueOf(parsedValue), context.getSafeValue())
}

/*GetBaseType - gettin base type like string, int, float32...*/
func (config JSONConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	logrus.Debug("Level: Debug. Json configurator source start.")
	parsed, notAValue := config.typeSafeLoadConfigFile(context)
	if !parsed {
		return *notAValue
	}
	nameField := context.StructField.Tag.Get(tags.TagJSON)
	if config.validation(nameField) {
		nameField = context.Prefix + context.StructField.Name
	}
	logrus.Debug("Level: Debug. Key for getting values from source: ", nameField)

	parsedValue, found := config.parsedData[nameField]
	logrus.Debug("Level: Debug. value: ", parsedValue)
	if !found || parsedValue == nil {
		return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("value for key '"+nameField+"' was not found in json source"))
	}
	return converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(parsedValue), context.getSafeValue())
}

// validation - true if everting ok
func (config JSONConfig) validation(value string) bool {
	return value == ""
}

func (config *JSONConfig) configuredFileFromEnv() {
	config.fileName = os.Getenv(tags.JSONFile)
}

// return true - if loaded config or successfully load config by filename
func (config *JSONConfig) typeSafeLoadConfigFile(context *structContext) (bool, *infra.GoStructorValue) {
	if config.fileName == "" {
		config.configuredFileFromEnv()
	}
	if config.parsedData == nil {
		fileBuffer, err := tools.ReadFromFile(config.fileName)
		if err != nil {
			notValue := infra.NewGoStructorNoValue(context.Value, err)
			return false, &notValue
		}
		parsedData := map[string]interface{}{}
		err1 := json.Unmarshal(fileBuffer.Bytes(), &parsedData)
		if err1 != nil {
			notValue := infra.NewGoStructorNoValue(context.Value, err1)
			return false, &notValue
		}
		config.parsedData = tools.FlatMap(parsedData)
		return true, nil
	}
	return true, nil
}
