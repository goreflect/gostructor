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
	for i := 0; i < value.NumField(); i++ {
		summirize = append(summirize, recurseStructField(value.Type().Field(i))...)
	}
	return summirize
}

func recurseStructField(structField reflect.StructField) []infra.FuncType {
	summirize := checkFuncsByTags(structField)
	switch structField.Type.Kind() {
	case reflect.Struct:
		for i := 0; i < structField.Type.NumField(); i++ {
			summirize = append(summirize, recurseStructField(structField.Type.Field(i))...)
		}
	}
	return summirize
}

func checkFuncsByTags(structField reflect.StructField) []infra.FuncType {
	summirize := make([]int, AmountTags) // amount repeats tags
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
			// TODO: add additional anaylys tag values for middlewares functions and others
			summirize[getFuncTypeByTag(value)]++
		}
	}
	result := []infra.FuncType{}
	for funcType, value := range summirize {
		if value > 0 {
			result = append(result, infra.FuncType(funcType))
		}
	}
	return result
}

func getFuncTypeByTag(tagName string) infra.FuncType {
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
		return infra.FunctionNotExist
	}
}
