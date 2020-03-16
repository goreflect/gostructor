package pipeline

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	gohocon "github.com/goreflect/go_hocon"
	"github.com/goreflect/gostructor/converters"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
)

type HoconConfig struct {
	fileName            string
	configureFileParsed *gohocon.Config
}

func (config HoconConfig) getElementName(context *structContext) string {
	currentTagHoconValue := context.StructField.Tag.Get(tags.TagHocon)
	if strings.Contains(context.Prefix, currentTagHoconValue) {
		fmt.Println("[HOCON]: Level: debug. Current field name: ", context.Prefix)
		return context.Prefix
	}
	returnName := context.Prefix + "."
	if currentTagHoconValue == "" {
		returnName += context.StructField.Name
	}

	returnName += currentTagHoconValue
	fmt.Println("[HOCON]: Level: debug. Current field name: ", returnName)
	return returnName
}

func (config HoconConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	loaded, structValue := config.typeSafeLoadConfigFile(context)
	if !loaded {
		return *structValue
	}
	valueIndirect := reflect.Indirect(context.Value)
	switch valueIndirect.Kind() {
	case reflect.Array:
		return config.getSliceFromHocon(context)
	case reflect.Slice:
		return config.getSliceFromHocon(context)
	case reflect.Map:
		return config.getMapFromHocon(context)
	default:
		return config.GetBaseType(context)
	}
}

// return true - if loaded config or successfully load config by filename
func (config *HoconConfig) typeSafeLoadConfigFile(context *structContext) (bool, *infra.GoStructorValue) {
	if config.configureFileParsed == nil {
		configParsed, err := gohocon.LoadConfig(config.fileName)
		if err != nil {
			notValue := infra.NewGoStructorNoValue(context.Value, err)
			return false, &notValue
		}
		config.configureFileParsed = configParsed
		return true, nil
	}
	return true, nil
}

func (config *HoconConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	configLoad, structValue := config.typeSafeLoadConfigFile(context)
	if !configLoad {
		return *structValue
	}
	valueIndirect := reflect.Indirect(context.Value)
	path := config.getElementName(context)
	switch valueIndirect.Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64, reflect.Bool, reflect.String:
		loadValue, errLoading := config.configureFileParsed.GetString(path)
		if errLoading != nil {
			return infra.NewGoStructorNoValue(context.Value.Interface(), errLoading)
		}
		return converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(loadValue), reflect.Indirect(context.Value))
	default:
		return infra.NewGoStructorNoValue(context.Value.Interface(),
			errors.New("can not parsed inserted type in GetBaseType of configuration by hocon"))
	}
}

func (config *HoconConfig) getSliceFromHocon(context *structContext) infra.GoStructorValue {
	configLoad, structValue := config.typeSafeLoadConfigFile(context)
	if !configLoad {
		return *structValue
	}
	path := config.getElementName(context)
	fmt.Println("[HOCON]: level: debug. get path from hocon: ", path)
	valueIndirect := reflect.Indirect(context.Value)
	setupSlice := reflect.MakeSlice(valueIndirect.Type(), 1, 1)
	fmt.Println("[HOCON]: level: debug. type of first element at slice: ", setupSlice.Index(0).Kind())
	// get string list
	valuesFromHocon, errGetting := config.configureFileParsed.GetStringList(path)
	if errGetting != nil {
		return infra.NewGoStructorNoValue(context.Value.Interface(), errGetting)
	}
	if setupSlice.Index(0).Kind() == reflect.Bool {
		neededValues, errLoading := config.configureFileParsed.GetBooleanList(path)
		if errLoading != nil {
			return infra.NewGoStructorNoValue(context.Value.Interface(), errLoading)
		}
		return infra.NewGoStructorTrueValue(reflect.ValueOf(neededValues))
	}
	convertedSlice := converters.ConvertBetweenComplexTypes(reflect.ValueOf(valuesFromHocon), valueIndirect)
	return convertedSlice
}

func (config *HoconConfig) getMapFromHocon(context *structContext) infra.GoStructorValue {
	configLoad, structValue := config.typeSafeLoadConfigFile(context)
	if !configLoad {
		return *structValue
	}
	valueIndirect := reflect.Indirect(context.Value)
	path := config.getElementName(context)
	fmt.Println("[HOCON]: level: debuf. current type: ", valueIndirect.Kind().String())
	fmt.Println("[HOCON]: level: debuf.key of map: ", valueIndirect.Type().Key().Kind())
	fmt.Println("[HOCON]: level: debuf.value of map: ", valueIndirect.Type().Elem().Kind())
	getValue, errLoading := config.configureFileParsed.GetValue(path)
	if errLoading != nil {
		return infra.NewGoStructorNoValue(context.Value.Interface(), errLoading)
	}
	object, errLoad := getValue.GetObject()
	if errLoad != nil {
		return infra.NewGoStructorNoValue(context.Value.Interface(), errLoad)
	}
	keys := object.GetKeys()
	makebleMap := reflect.MakeMapWithSize(valueIndirect.Type(), 0)
	for _, key := range keys {
		value := config.GetBaseType(&structContext{
			StructField: reflect.StructField{
				Name: key,
			},
			Prefix: context.Prefix + "." + key,
			Value:  reflect.New(valueIndirect.Type().Elem()),
		})

		parsedKey := converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(key), reflect.Indirect(reflect.New(valueIndirect.Type().Key())))
		if parsedKey.CheckIsValue() {
			if value.CheckIsValue() {
				makebleMap.SetMapIndex(reflect.Indirect(parsedKey.Value).Convert(valueIndirect.Type().Key()), value.Value)
			} else {
				return infra.NewGoStructorNoValue(parsedKey.Value, errors.New("can not parsed value for map"))
			}
		} else {
			return infra.NewGoStructorNoValue(parsedKey.Value, errors.New("can not parsed key for map"))
		}

	}
	return infra.NewGoStructorTrueValue(makebleMap)
}
