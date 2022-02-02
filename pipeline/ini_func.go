package pipeline

import (
	"os"
	"reflect"
	"strings"

	"github.com/go-ini/ini"
	"github.com/goreflect/gostructor/converters"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
	"github.com/goreflect/gostructor/tools"
	"github.com/sirupsen/logrus"
)

/*IniConfig - source ini for configuring*/
type IniConfig struct {
	fileName string
	iniFile  *ini.File
}

/*GetComplexType - getting from ini slices, maps, arrays...*/
func (config IniConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	logrus.Debug("Ini configurator source run")
	parsed, notAValue := config.typeSafeLoadConfigFile(context)
	if !parsed {
		return *notAValue
	}
	nameField := context.StructField.Tag.Get(tags.TagIni)
	var sectionName string = ""
	var field string = ""
	if config.validation(nameField) {
		nameField = context.Prefix + context.StructField.Name
	}

	if strings.Contains(nameField, "#") {
		splited := strings.Split(nameField, "#")
		logrus.Debug("Level: Debug. Section and key for getting values from source: ", splited)
		sectionName = splited[0]
		field = splited[1]
	} else {
		field = nameField
	}
	parsedSection, err := config.iniFile.GetSection(sectionName)
	if err != nil {
		logrus.Error("Can not parsed section", err)
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	parsedKey, err := parsedSection.GetKey(field)
	if err != nil {
		logrus.Error("Can not parsed key", err)
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	return converters.ConvertBetweenComplexTypes(reflect.ValueOf(parsedKey.Strings(",")), context.getSafeValue())
}

/*GetBaseType - getting from ini string, int, float32 ...*/
func (config IniConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	logrus.Debug("Level: Debug. Ini configurator source start.")
	parsed, notAValue := config.typeSafeLoadConfigFile(context)
	if !parsed {
		return *notAValue
	}
	nameField := context.StructField.Tag.Get(tags.TagIni)
	var sectionName string = ""
	var field string = ""
	if config.validation(nameField) {
		nameField = context.Prefix + context.StructField.Name
	}

	if strings.Contains(nameField, "#") {
		splited := strings.Split(nameField, "#")
		logrus.Debug("Level: Debug. Section and key for getting values from source: ", splited)
		sectionName = splited[0]
		field = splited[1]
	} else {
		field = nameField
	}
	parsedSection, err := config.iniFile.GetSection(sectionName)
	if err != nil {
		logrus.Error("Can not parsed section", err)
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	parsedKey, err := parsedSection.GetKey(field)
	if err != nil {
		logrus.Error("Can not parsed key", err)
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	return converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(parsedKey.Value()), context.getSafeValue())
}

// validation - true if everting ok
func (config IniConfig) validation(value string) bool {
	return value == ""
}

func (config *IniConfig) configuredFileFromEnv() {
	config.fileName = os.Getenv(tags.IniFile)
}

// return true - if loaded config or successfully load config by filename
func (config *IniConfig) typeSafeLoadConfigFile(context *structContext) (bool, *infra.GoStructorValue) {
	if config.fileName == "" {
		config.configuredFileFromEnv()
	}

	if config.iniFile == nil {
		fileBuffer, err := tools.ReadFromFile(config.fileName)
		if err != nil {
			notValue := infra.NewGoStructorNoValue(context.Value, err)
			return false, &notValue
		}

		file, err := ini.Load(fileBuffer)
		if err != nil {
			logrus.Error("Error while reading data from file: " + err.Error())
		}
		config.iniFile = file
	}

	return true, nil
}
