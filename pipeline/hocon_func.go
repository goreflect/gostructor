package pipeline

import (
	"errors"
	"fmt"
	"reflect"

	gohocon "github.com/go-akka/configuration"
)

type HoconConfig struct {
	fileName            string
	configureFileParsed *gohocon.Config
}

func (config HoconConfig) GetComplexType(context *structContext) GoStructorValue {
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
	path := context.Prefix + "." + context.StructField.Name
	switch valueIndirect.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valueParsedFromConfigFile := config.configureFileParsed.GetInt64(path)
		return NewGoStructorTrueValue(reflect.ValueOf(valueParsedFromConfigFile))
	case reflect.Float32:
		valueParsedFromConfigFile := config.configureFileParsed.GetFloat32(path)
		return NewGoStructorTrueValue(reflect.ValueOf(valueParsedFromConfigFile))
	case reflect.Float64:
		valueParsedFromConfigFile := config.configureFileParsed.GetFloat64(path)
		return NewGoStructorTrueValue(reflect.ValueOf(valueParsedFromConfigFile))
	case reflect.Bool:
		valueParsedFromConfigFile := config.configureFileParsed.GetBoolean(path)
		return NewGoStructorTrueValue(reflect.ValueOf(valueParsedFromConfigFile))
	case reflect.String:
		valueParsedFromConfigFile := config.configureFileParsed.GetString(path)
		return NewGoStructorTrueValue(reflect.ValueOf(valueParsedFromConfigFile))
	default:
		return NewGoStructorNoValue(context.Value.Interface(), errors.New("can not parsed inserted type in GetBaseType of configuration by hocon"))
	}
}

func (config *HoconConfig) getSliceFromHocon(context *structContext) GoStructorValue {

	path := context.Prefix + "." + context.StructField.Name
	fmt.Println("level: debug. get path from hocon: ", path)
	valueIndirect := reflect.Indirect(context.Value)
	setupSlice := reflect.MakeSlice(valueIndirect.Type(), 1, 1)
	switch setupSlice.Index(0).Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		neededValues := config.configureFileParsed.GetInt32List(path)
		// if err != nil {
		// 	return reflect.Zero(nil), err
		// }
		return NewGoStructorTrueValue(reflect.ValueOf(neededValues))
	case reflect.Uint64, reflect.Int64:
		neededValues := config.configureFileParsed.GetInt64List(path)
		return NewGoStructorTrueValue(reflect.ValueOf(neededValues))
	case reflect.Float32:
		neededValues := config.configureFileParsed.GetFloat32List(path)
		return NewGoStructorTrueValue(reflect.ValueOf(neededValues))
	case reflect.Float64:
		neededValues := config.configureFileParsed.GetFloat64List(path)
		return NewGoStructorTrueValue(reflect.ValueOf(neededValues))
	case reflect.Bool:
		neededValues := config.configureFileParsed.GetBooleanList(path)
		return NewGoStructorTrueValue(reflect.ValueOf(neededValues))
	case reflect.Complex64, reflect.Complex128:
		return NewGoStructorNoValue(context.Value.Interface(), errors.New("not supported yet a complex values"))
	default:
		return NewGoStructorNoValue(context.Value.Interface(), errors.New("can not recognize inserted type"))
	}
	// list := config.configureFileParsed.GetStringList(path)
	// valueList := reflect.ValueOf(list)
	// return valueList, nil
	// valueIndirect := reflect.Indirect(context.Value)
	// setupSlice := reflect.MakeSlice(valueIndirect.Type(), valueList.Len(), valueList.Cap())
	// for i := 0; i < valueList.Len(); i++ {
	// 	// fmt.Println("type convertable: ", valueIndirect.Index(0).Type())
	// 	fmt.Println("type source: ", valueList.Index(i).Type())
	// 	// elementSetupSlice := reflect.Indirect(setupSlice.Index(0))
	// 	fmt.Println("type of 1 element makeble slice: ", setupSlice.Index(0).Type())
	// 	resultConvert, errResult := converters.Convert(valueList.Index(i), setupSlice.Index(0))
	// 	if errResult != nil {
	// 		fmt.Println("Error converted value: ", errResult)
	// 	}
	// 	fmt.Println("result of convert: ", resultConvert)
	// 	insertedValue := valueList.Index(i).Type().ConvertibleTo(setupSlice.Index(0).Type())
	// 	fmt.Println("value can be convertable to: ", insertedValue)

	// 	if insertedValue {
	// 		convertableValue := valueList.Index(i).Convert(setupSlice.Index(0).Type())
	// 		reflect.Indirect(setupSlice.Index(i)).Set(convertableValue)
	// 	} else {
	// 		fmt.Println("can not convert your types. Converte from: ", valueList.Index(0).Type(), " to: ", setupSlice.Index(0).Type())
	// 	}
	// 	// result, err := context.conversion(valueList.Index(i), valueNotPointer.Elem().Type())
	// 	// if err != nil {
	// 	// fmt.Println("can not insert in your slice: ", path, " value. Error: ", err.Error())
	// 	// return errors.New(err.Error())
	// 	// }

	// }
	// fmt.Println("Setuped slice: ", setupSlice.Interface())
	// return nil
}

func (config *HoconConfig) getMapFromHocon(context *structContext) GoStructorValue {
	// config.configureFileParsed
	// valueIndirect := reflect.Indirect(context.Value)
	makebleMap := reflect.MakeMap(context.Value.Type())
	for key, value := range makebleMap.MapRange() {

	}
	return NewGoStructorNoValue(context.Value.Interface(), errors.New("not implement getMap from hocon configuring"))
}
