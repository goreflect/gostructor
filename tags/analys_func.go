package tags

import (
	"reflect"

	"github.com/goreflect/gostructor/infra"
)

/*
GetFunctionTypes - return slice of functions which should configuring sourceStruct structure
*/
func GetFunctionTypes(sourceStruct interface{}) []infra.FuncType {
	summirize := []infra.FuncType{}
	value := reflect.Indirect(reflect.ValueOf(sourceStruct))
	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			summirize = GetFunctionTypes(value.Type().Field(i))
		}
	default:
	}
	return summirize
}

func CheckFuncsByTags(structField reflect.StructField) []infra.FuncType {

	summirize := []infra.FuncType{}
	for _, value := range []string{
		TagYaml,
		TagJson,
		TagHocon,
		TagHashiCorpVault,
		TagEnvironment,
		TagDefault,
		TagConfigServer,
	} {
		tagInField := structField.Tag.Get(value)
		if tagInField == "" {
			continue
		} else {
			summirize = append(summirize, GetFuncTypeByTag(tagInField))
		}
	}
	return summirize
}

func GetFuncTypeByTag(tagName string) infra.FuncType {
	switch tagName {
	case TagYaml:
		return infra.FunctionSetupYaml
	case TagConfigServer:
		return infra.FunctionSetupConfigServer
	case TagDefault:
		return infra.FunctionSetupDefault
	case TagEnvironment:
		return infra.FunctionSetupEnvironment
	case TagHashiCorpVault:
		return infra.FunctionSetupVault
	case TagHocon:
		return infra.FunctionSetupHocon
	case TagJson:
		return infra.FunctionSetupJson
	default:
		return infra.FuncType(-1)
	}
}
