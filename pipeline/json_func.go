package pipeline

import (
	"errors"
	"fmt"

	"github.com/go-restit/lzjson"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
	"github.com/goreflect/gostructor/tools"
)

/*JSONConfig - source json configuring*/
type JSONConfig struct {
	FileName            string
	configureFileParsed lzjson.Node
}

/*GetComplexType - get complex types like arrays, slices, maps from json source*/
func (config JSONConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Json configurator source start.")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getcomplext type from json not implemented"))
}

/*GetBaseType - gettin base type like string, int, float32...*/
func (config JSONConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Json configurator source start.")
	parsed, notAValue := config.typeSafeLoadConfigFile(context)
	if !parsed {
		return *notAValue
	}
	nameField := context.StructField.Tag.Get(tags.TagJson)
	if config.validation(nameField) {
		nameField = context.Prefix + context.StructField.Name
	}
	fmt.Println("Level: Debug. Key for getting values from source: ", nameField)

	parsedValue := config.configureFileParsed.Get(nameField)
	fmt.Println("Level: Debug. Get from json source: ", parsedValue.String())
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getbase type from json not implemented"))
}

// validation - true if everting ok
func (config JSONConfig) validation(value string) bool {
	return value == ""
}

// return true - if loaded config or successfully load config by filename
func (config *JSONConfig) typeSafeLoadConfigFile(context *structContext) (bool, *infra.GoStructorValue) {
	if config.configureFileParsed == nil {
		fileBuffer, err := tools.ReadFromFile(config.FileName)
		if err != nil {
			notValue := infra.NewGoStructorNoValue(context.Value, err)
			return false, &notValue
		}
		configParsed := lzjson.Decode(fileBuffer)
		if err != nil {
			notValue := infra.NewGoStructorNoValue(context.Value, err)
			return false, &notValue
		}
		config.configureFileParsed = configParsed
		return true, nil
	}
	return true, nil
}
