package pipeline

import (
	"errors"
	"os"
	"reflect"

	"github.com/go-restit/lzjson"
	"github.com/goreflect/gostructor/converters"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
	"github.com/goreflect/gostructor/tools"
	"github.com/sirupsen/logrus"
)

/*JSONConfig - source json configuring*/
type JSONConfig struct {
	fileName            string
	configureFileParsed lzjson.Node
}

/*GetComplexType - get complex types like arrays, slices, maps from json source*/
func (config JSONConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	logrus.Debug("Level: Debug. Json configurator source start.")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getcomplext type from json not implemented"))
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

	parsedValue := config.configureFileParsed.Get(nameField)
	logrus.Error("Node: ", config.configureFileParsed.Type(), "values: ", string(config.configureFileParsed.Raw()))
	if parsedValue.ParseError() != nil {
		logrus.Error("Can not parsed value from json decoder: ", parsedValue.ParseError())
		return infra.NewGoStructorNoValue(context.Value, parsedValue.ParseError())
	}
	logrus.Debug("Level: Debug. Get from json source: ", parsedValue.String())
	if !config.validation(parsedValue.String()) {
		return converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(parsedValue.String()), context.Value)
	}
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getbase type from json not implemented"))
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

	}
	if config.configureFileParsed == nil {
		fileBuffer, err := tools.ReadFromFile(config.fileName)
		if err != nil {
			notValue := infra.NewGoStructorNoValue(context.Value, err)
			return false, &notValue
		}
		config.configureFileParsed = lzjson.Decode(fileBuffer)
		return true, nil
	}
	return true, nil
}
