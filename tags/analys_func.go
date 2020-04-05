package tags

import (
	"reflect"

	"github.com/goreflect/gostructor/infra"
)

/*
GetFunctionTypes - return slice of functions which should configure sourceStruct structure
*/
func GetFunctionTypes(sourceStruct interface{}) []infra.FuncType {
	summarize := []int{}
	value := reflect.Indirect(reflect.ValueOf(sourceStruct))
	for i := 0; i < value.NumField(); i++ {
		summirizeLevel := recurseStructField(value.Type().Field(i))
		summarize = combineFields(summarize, summirizeLevel)
	}
	result := []infra.FuncType{}
	for funcType, value := range summarize {
		if value > 0 {
			result = append(result, infra.FuncType(funcType))
		}
	}
	return result
}

func recurseStructField(structField reflect.StructField) []int {
	summarize := checkFuncsByTags(structField)

	switch structField.Type.Kind() {
	case reflect.Struct:
		for i := 0; i < structField.Type.NumField(); i++ {
			summarizeLevel := recurseStructField(structField.Type.Field(i))
			summarize = combineFields(summarize, summarizeLevel)
		}
	}
	return summarize
}

//TODO: decomposition
func combineFields(summCurrent []int, newSumm []int) []int {
	maxCount := 0
	if len(summCurrent) > len(newSumm) {
		maxCount = len(summCurrent)
	} else {
		maxCount = len(newSumm)
	}
	newResult := make([]int, maxCount)
	for index, value := range newSumm {
		newResult[index] += value
	}
	for index, value := range summCurrent {
		newResult[index] += value
	}
	return newResult
}

func checkFuncsByTags(structField reflect.StructField) []int {
	summarize := make([]int, AmountTags) // amount repeats tags
	for _, value := range []string{
		TagYaml,
		TagJSON,
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

			summarize[getFuncTypeByTag(value)]++
		}
	}

	return summarize
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
	case TagJSON:
		return infra.FunctionSetupJson
	default:
		return infra.FunctionNotExist
	}
}
