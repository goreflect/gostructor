package pipeline

import (
	"errors"
	"reflect"
	"strings"

	gohocon "github.com/goreflect/go_hocon"
	"github.com/goreflect/gostructor/converters"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
	"github.com/sirupsen/logrus"
)

/*
HoconConfig - configuring source from hocon
*/
type HoconConfig struct {
	fileName            string
	configureFileParsed *gohocon.Config
}

func (config HoconConfig) getElementName(context *structContext) string {
	currentTagHoconValue := context.StructField.Tag.Get(tags.TagHocon)
	if strings.Contains(context.Prefix, currentTagHoconValue) {
		logrus.Debug("Current field name: ", context.Prefix)
		return context.Prefix
	}
	returnName := context.Prefix + "."
	if currentTagHoconValue == "" {
		returnName += context.StructField.Name
	}

	returnName += currentTagHoconValue
	logrus.Debug("Current field name: ", returnName)
	return returnName
}

/*
GetComplexType - get complex types like slices, maps, arrays,
*/
func (config HoconConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	if errLoading := config.typeSafeLoadConfigFile(); errLoading != nil {
		return infra.NewGoStructorNoValue(context.Value, errLoading)
	}
	valueIndirect := reflect.Indirect(context.Value)
	switch valueIndirect.Kind() {
	case reflect.Slice, reflect.Array:
		return config.getSliceFromHocon(context)
	case reflect.Map:
		return config.getMapFromHocon(context)
	default:
		return config.GetBaseType(context)
	}
}

// return true - if loaded config or successfully load config by filename
func (config *HoconConfig) typeSafeLoadConfigFile() error {
	if config.configureFileParsed == nil {
		configParsed, err := gohocon.LoadConfig(config.fileName)
		if err != nil {
			return err
		}
		config.configureFileParsed = configParsed
		return nil
	}
	return nil
}

/*
GetBaseType - get base types like string, int, float32
*/
func (config *HoconConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	if errLoading := config.typeSafeLoadConfigFile(); errLoading != nil {
		return infra.NewGoStructorNoValue(context.Value, errLoading)
	}
	path := config.getElementName(context)

	loadValue, errLoading := config.configureFileParsed.GetString(path)
	if errLoading != nil {
		return infra.NewGoStructorNoValue(context.Value.Interface(), errLoading)
	}
	return converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(loadValue), reflect.Indirect(context.Value))
}

func (config *HoconConfig) getSliceFromHocon(context *structContext) infra.GoStructorValue {
	if errLoading := config.typeSafeLoadConfigFile(); errLoading != nil {
		return infra.NewGoStructorNoValue(context.Value, errLoading)
	}
	path := config.getElementName(context)
	logrus.Debug("get path from hocon: ", path)
	valueIndirect := reflect.Indirect(context.Value)
	setupSlice := reflect.MakeSlice(valueIndirect.Type(), 1, 1)
	logrus.Debug("type of first element at slice: ", setupSlice.Index(0).Kind())
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
	if errLoading := config.typeSafeLoadConfigFile(); errLoading != nil {
		return infra.NewGoStructorNoValue(context.Value, errLoading)
	}
	valueIndirect := reflect.Indirect(context.Value)
	path := config.getElementName(context)
	logrus.Debug("current type: ", valueIndirect.Kind().String())
	logrus.Debug("key of map: ", valueIndirect.Type().Key().Kind())
	logrus.Debug("value of map: ", valueIndirect.Type().Elem().Kind())
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
