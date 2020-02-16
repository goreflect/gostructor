package pipeline

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	gohocon "github.com/go-akka/configuration"
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

func (config HoconConfig) GetComplexType(context *structContext) GoStructorValue {
	if config.configureFileParsed == nil {
		config.configureFileParsed = gohocon.LoadConfig(config.fileName)
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

func (config *HoconConfig) GetBaseType(context *structContext) GoStructorValue {
	valueIndirect := reflect.Indirect(context.Value)
	if config.configureFileParsed == nil {
		config.configureFileParsed = gohocon.LoadConfig(config.fileName)
	}
	path := config.getElementName(context)
	switch valueIndirect.Kind() {
	case reflect.Int:
		valueParsedFromConfigFile, errParsed := strconv.ParseInt(config.configureFileParsed.GetString(path), 10, 16)
		if errParsed != nil {
			return NewGoStructorNoValue(valueIndirect, errParsed)
		}
		return NewGoStructorTrueValue(reflect.ValueOf(int(valueParsedFromConfigFile)))
	case reflect.Int8:
		valueParsedFromConfigFile, errParsed := strconv.ParseInt(config.configureFileParsed.GetString(path), 10, 8)
		if errParsed != nil {
			return NewGoStructorNoValue(valueIndirect, errParsed)
		}
		return NewGoStructorTrueValue(reflect.ValueOf(int8(valueParsedFromConfigFile)))
	case reflect.Int16:
		valueParsedFromConfigFile, errParsed := strconv.ParseInt(config.configureFileParsed.GetString(path), 10, 16)
		if errParsed != nil {
			return NewGoStructorNoValue(valueIndirect, errParsed)
		}
		return NewGoStructorTrueValue(reflect.ValueOf(int16(valueParsedFromConfigFile)))
	case reflect.Int32:
		// TODO: upgrade while completing lib for safety parsing from file
		valueParsedFromConfigFile := config.configureFileParsed.GetInt32(path)
		return NewGoStructorTrueValue(reflect.ValueOf(valueParsedFromConfigFile))
	case reflect.Int64:
		// TODO: upgrade while completing lib for safety parsing from file
		valueParsedFromConfigFile := config.configureFileParsed.GetInt64(path)
		return NewGoStructorTrueValue(reflect.ValueOf(valueParsedFromConfigFile))
	case reflect.Float32:
		// TODO: upgrade while completing lib for safety parsing from file
		valueParsedFromConfigFile := config.configureFileParsed.GetFloat32(path)
		return NewGoStructorTrueValue(reflect.ValueOf(valueParsedFromConfigFile))
	case reflect.Float64:
		// TODO: upgrade while completing lib for safety parsing from file
		valueParsedFromConfigFile := config.configureFileParsed.GetFloat64(path)
		return NewGoStructorTrueValue(reflect.ValueOf(valueParsedFromConfigFile))
	case reflect.Bool:
		// TODO: upgrade while completing lib for safety parsing from file
		valueParsedFromConfigFile := config.configureFileParsed.GetBoolean(path)
		return NewGoStructorTrueValue(reflect.ValueOf(valueParsedFromConfigFile))
	case reflect.String:
		// TODO: upgrade while completing lib for safety parsing from file
		valueParsedFromConfigFile := config.configureFileParsed.GetString(path)
		fmt.Println("[HOCON]: Level: debug. Get value from config file: ", valueParsedFromConfigFile)
		return NewGoStructorTrueValue(reflect.ValueOf(valueParsedFromConfigFile))
	default:
		return NewGoStructorNoValue(context.Value.Interface(), errors.New("can not parsed inserted type in GetBaseType of configuration by hocon"))
	}
}

func (config *HoconConfig) getSliceFromHocon(context *structContext) GoStructorValue {

	path := config.getElementName(context)
	fmt.Println("[HOCON]: level: debug. get path from hocon: ", path)
	valueIndirect := reflect.Indirect(context.Value)
	setupSlice := reflect.MakeSlice(valueIndirect.Type(), 1, 1)
	switch setupSlice.Index(0).Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		// TODO: upgrade while completing lib for safety parsing from file
		neededValues := config.configureFileParsed.GetInt32List(path)
		return NewGoStructorTrueValue(reflect.ValueOf(neededValues))
	case reflect.Uint64, reflect.Int64:
		// TODO: upgrade while completing lib for safety parsing from file
		neededValues := config.configureFileParsed.GetInt64List(path)
		return NewGoStructorTrueValue(reflect.ValueOf(neededValues))
	case reflect.Float32:
		// TODO: upgrade while completing lib for safety parsing from file
		neededValues := config.configureFileParsed.GetFloat32List(path)
		return NewGoStructorTrueValue(reflect.ValueOf(neededValues))
	case reflect.Float64:
		// TODO: upgrade while completing lib for safety parsing from file
		neededValues := config.configureFileParsed.GetFloat64List(path)
		return NewGoStructorTrueValue(reflect.ValueOf(neededValues))
	case reflect.Bool:
		// TODO: upgrade while completing lib for safety parsing from file
		neededValues := config.configureFileParsed.GetBooleanList(path)
		return NewGoStructorTrueValue(reflect.ValueOf(neededValues))
	case reflect.Complex64, reflect.Complex128:
		// TODO: upgrade while completing lib for safety parsing from file
		return NewGoStructorNoValue(context.Value.Interface(), errors.New("not supported yet a complex values"))
	default:
		return NewGoStructorNoValue(context.Value.Interface(), errors.New("can not recognize inserted type"))
	}
}

func (config *HoconConfig) getMapFromHocon(context *structContext) GoStructorValue {
	// config.configureFileParsed
	if config.configureFileParsed == nil {
		config.configureFileParsed = gohocon.LoadConfig(config.fileName)
	}
	valueIndirect := reflect.Indirect(context.Value)
	path := config.getElementName(context)
	fmt.Println("[HOCON]: level: debuf. current type: ", valueIndirect.Kind().String())
	fmt.Println("[HOCON]: level: debuf.key of map: ", valueIndirect.Type().Key().Kind())
	fmt.Println("[HOCON]: level: debuf.value of map: ", valueIndirect.Type().Elem().Kind())
	getValue := config.configureFileParsed.GetValue(path)
	keys := getValue.GetObject().GetKeys()
	makebleMap := reflect.MakeMapWithSize(valueIndirect.Type(), 0)
	for _, key := range keys {
		value := config.GetBaseType(&structContext{
			StructField: reflect.StructField{
				Name: key,
			},
			Prefix: context.Prefix + "." + context.StructField.Name,
			Value:  reflect.New(valueIndirect.Type().Elem()),
		})

		parsedKey := parseMapType(valueIndirect.Type().Key(), reflect.ValueOf(key))
		if parsedKey.CheckIsValue() {
			if value.CheckIsValue() {
				makebleMap.SetMapIndex(reflect.Indirect(parsedKey.Value).Convert(valueIndirect.Type().Key()), value.Value)
			} else {
				return NewGoStructorNoValue(parsedKey.Value, errors.New("can not parsed value for map"))
			}
		} else {
			return NewGoStructorNoValue(parsedKey.Value, errors.New("can not parsed key for map"))
		}

	}
	return NewGoStructorTrueValue(makebleMap)
}

func parseMapType(needType reflect.Type, value reflect.Value) GoStructorValue {
	switch needType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsedValue, errParsed := strconv.ParseInt(reflect.Indirect(value).Interface().(string), 10, 64)
		if errParsed != nil {
			return NewGoStructorNoValue(nil, errParsed)
		}

		return NewGoStructorTrueValue(reflect.ValueOf(parsedValue))
	case reflect.String:
		return NewGoStructorTrueValue(value)
	case reflect.Float32:
		parsedValue, errParsed := strconv.ParseFloat(reflect.Indirect(value).Interface().(string), 32)
		if errParsed != nil {
			return NewGoStructorNoValue(nil, errParsed)
		}

		return NewGoStructorTrueValue(reflect.ValueOf(parsedValue))
	case reflect.Float64:
		parsedValue, errParsed := strconv.ParseFloat(reflect.Indirect(value).Interface().(string), 64)
		if errParsed != nil {
			return NewGoStructorNoValue(nil, errParsed)
		}

		return NewGoStructorTrueValue(reflect.ValueOf(parsedValue))
	default:
		return NewGoStructorNoValue(nil, errors.New("can not set for map key by insert type: "+needType.Kind().String()))
	}
}
