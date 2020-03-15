package converters

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/goreflect/gostructor/infra"
)

/*ConvertBetweenPrimitiveTypes - method for converting from any of base types into any of base types,
like string, bool, int, int8, int16. int32, int64
*/
func ConvertBetweenPrimitiveTypes(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	fmt.Println("Level: Debug. Message: start converting source: ", source.Kind().String(), " destination: ", destination.Kind().String())
	switch destination.Kind() {
	case reflect.Int:
		return convertToInt(source, destination)
	case reflect.Int8:
		return convertToInt8(source, destination)
	case reflect.Int16:
		return convertToInt16(source, destination)
	case reflect.Int32:
		return convertToInt32(source, destination)
	case reflect.Int64:
		return convertToInt64(source, destination)
	case reflect.String:
		return convertToString(source, destination)
	case reflect.Bool:
		return convertToBool(source, destination)
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not converted to this type "+destination.Kind().String()+" beacuse this type not supported"))
	}
}

func ConvertBetweenComplexTypes(source reflect.Value, destination reflect.Value) {

}
