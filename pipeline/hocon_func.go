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

// Configure - configuration struct value by hocon file configuration
func (config HoconConfig) Configure(context *structContext) error {
	fmt.Println("Level: Debug. hocon values sources start")
	if config.configureFileParsed == nil {
		config.configureFileParsed = gohocon.LoadConfig(config.fileName)
	}
	valueIndirect := reflect.Indirect(context.Value)
	switch valueIndirect.Kind() {
	case reflect.Slice:
		return config.getSliceFromHocon(context)
	case reflect.Array:
	case reflect.Map:
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
	case reflect.String:
	case reflect.Float32:
	case reflect.Float64:
	case reflect.Bool:
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
	default:
		return errors.New("not supported type for hocon parsing")
	}
	return nil
}

func (config *HoconConfig) getSliceFromHocon(context *structContext) (err error) {
	defer func() {
		if errPanic := recover(); errPanic != nil {
			err = errPanic.(error)
		}
	}()
	path := context.Prefix + "." + context.StructField.Name
	fmt.Println("level: debug. get path from hocon: ", path)
	list := config.configureFileParsed.GetStringList(path)
	valueList := reflect.ValueOf(list)
	valueIndirect := reflect.Indirect(context.Value)
	setupSlice := reflect.MakeSlice(valueIndirect.Type(), valueList.Len(), valueList.Cap())
	for i := 0; i < valueList.Len(); i++ {
		// fmt.Println("type convertable: ", valueIndirect.Index(0).Type())
		fmt.Println("type source: ", valueList.Index(i).Type())
		// elementSetupSlice := reflect.Indirect(setupSlice.Index(0))
		fmt.Println("type of 1 element makeble slice: ", setupSlice.Index(0).Type())
		insertedValue := valueList.Index(i).Type().ConvertibleTo(setupSlice.Index(0).Type())
		fmt.Println("value can be convertable to: ", insertedValue)

		if insertedValue {
			convertableValue := valueList.Index(i).Convert(setupSlice.Index(0).Type())
			reflect.Indirect(setupSlice.Index(i)).Set(convertableValue)
		} else {
			fmt.Println("can not convert your types. Converte from: ", valueList.Index(0).Type(), " to: ", setupSlice.Index(0).Type())
		}
		// result, err := context.conversion(valueList.Index(i), valueNotPointer.Elem().Type())
		// if err != nil {
		// fmt.Println("can not insert in your slice: ", path, " value. Error: ", err.Error())
		// return errors.New(err.Error())
		// }

	}
	fmt.Println("Setuped slice: ", setupSlice.Interface())
	return nil
}

func (config *HoconConfig) getArrayFromHocon(context *structContext) error {

	return nil
}

func (config *HoconConfig) getMapFromHocon(context *structContext) error {

	return nil
}
