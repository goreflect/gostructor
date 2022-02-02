package pipeline

import (
	"os"
	"reflect"
	"strings"

	"github.com/goreflect/gostructor/converters"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
	"github.com/goreflect/gostructor/tools"
	toml "github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
)

/*TomlConfig - source toml for configuring*/
type TomlConfig struct {
	fileName   string
	parsedData *toml.Tree
}

/*GetComplexType - getting from config slices, maps, arrays...*/
func (config TomlConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	logrus.Debug("toml configurator source run")
	parsed, notAValue := config.typeSafeLoadConfigFile(context)
	if !parsed {
		return *notAValue
	}
	nameField := context.StructField.Tag.Get(tags.TagToml)
	var sectionName string = ""
	var field string = ""
	if config.validation(nameField) {
		nameField = context.Prefix + context.StructField.Name
	}

	if strings.Contains(nameField, "#") {
		splited := strings.Split(nameField, "#")
		logrus.Debug("Level: Debug. Section and key for getting values from source: ", splited)
		sectionName = splited[0] + "."
		field = splited[1]
	} else {
		field = nameField
	}
	parsedSection := config.parsedData.Get(sectionName + field)
	return converters.ConvertBetweenComplexTypes(reflect.ValueOf(parsedSection), context.getSafeValue())
}

/*GetBaseType - getting from config string, int, float32 ...*/
func (config TomlConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	logrus.Debug("Level: Debug. Toml configurator source start.")
	parsed, notAValue := config.typeSafeLoadConfigFile(context)
	if !parsed {
		return *notAValue
	}
	nameField := context.StructField.Tag.Get(tags.TagToml)
	var sectionName string = ""
	var field string = ""
	if config.validation(nameField) {
		nameField = context.Prefix + context.StructField.Name
	}

	if strings.Contains(nameField, "#") {
		splited := strings.Split(nameField, "#")
		logrus.Debug("Level: Debug. Section and key for getting values from source: ", splited)
		sectionName = splited[0] + "."
		field = splited[1]
	} else {
		field = nameField
	}
	parsedSection := config.parsedData.Get(sectionName + field)
	return converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(parsedSection), context.getSafeValue())
}

// validation - true if everting ok
func (config TomlConfig) validation(value string) bool {
	return value == ""
}

func (config *TomlConfig) configuredFileFromEnv() {
	config.fileName = os.Getenv(tags.TomlFile)
}

// return true - if loaded config or successfully load config by filename
func (config *TomlConfig) typeSafeLoadConfigFile(context *structContext) (bool, *infra.GoStructorValue) {
	if config.fileName == "" {
		config.configuredFileFromEnv()
	}
	if config.parsedData == nil {
		fileBuffer, err := tools.ReadFromFile(config.fileName)
		if err != nil {
			notValue := infra.NewGoStructorNoValue(context.Value, err)
			return false, &notValue
		}
		tomlData, err1 := toml.Load(fileBuffer.String())
		if err1 != nil {
			var notValue = infra.NewGoStructorNoValue(context.getSafeValue(), err1)
			return false, &notValue
		}
		config.parsedData = tomlData
		return true, nil
	}
	return true, nil
}
