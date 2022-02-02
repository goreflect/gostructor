package pipeline

import (
	"errors"
	"os"
	"reflect"

	"github.com/goccy/go-yaml"
	"github.com/goreflect/gostructor/converters"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
	"github.com/goreflect/gostructor/tools"
	"github.com/sirupsen/logrus"
)

/*YamlConfig - source yaml for configuring*/
type YamlConfig struct {
	fileName   string
	parsedData map[string]interface{}
}

/*GetComplexType - getting from yaml slices, maps, arrays...*/
func (yaml YamlConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	logrus.Debug("Level: Debug. Yaml configurator source start.")
	parsed, notAValue := yaml.typeSafeLoadConfigFile(context)
	if !parsed {
		return *notAValue
	}
	nameField := context.StructField.Tag.Get(tags.TagYaml)
	if yaml.validation(nameField) {
		nameField = context.Prefix + context.StructField.Name
	}
	logrus.Debug("Level: Debug. Key for getting values from source: ", nameField)

	parsedValue := yaml.parsedData[nameField]
	if !yaml.validation(parsedValue.(string)) {
		return converters.ConvertBetweenComplexTypes(reflect.ValueOf(parsedValue), context.Value)
	}
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getcomplex type from yaml not implemented"))
}

/*GetBaseType - getting from yaml string, int, float32 ...*/
func (yaml YamlConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	logrus.Debug("Level: Debug. Yaml configurator source start.")
	parsed, notAValue := yaml.typeSafeLoadConfigFile(context)
	if !parsed {
		return *notAValue
	}
	nameField := context.StructField.Tag.Get(tags.TagYaml)
	if yaml.validation(nameField) {
		nameField = context.Prefix + context.StructField.Name
	}
	logrus.Debug("Level: Debug. Key for getting values from source: ", nameField)

	parsedValue := yaml.parsedData[nameField]
	logrus.Debug("Level: Debug. value: ", parsedValue)
	if !yaml.validation(parsedValue.(string)) {
		return converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(parsedValue), context.Value)
	}
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("get base type from yaml not implemented"))
}

// validation - true if everting ok
func (config YamlConfig) validation(value string) bool {
	return value == ""
}

func (config *YamlConfig) configuredFileFromEnv() {
	config.fileName = os.Getenv(tags.YamlFile)
}

// return true - if loaded config or successfully load config by filename
func (config *YamlConfig) typeSafeLoadConfigFile(context *structContext) (bool, *infra.GoStructorValue) {
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
		err1 := yaml.Unmarshal(fileBuffer.Bytes(), &parsedData)
		if err1 != nil {
			var notValue = infra.NewGoStructorNoValue(context.Value, err1)
			return false, &notValue
		}
		config.parsedData = tools.FlatMap(parsedData)
		return true, nil
	}
	return true, nil
}
