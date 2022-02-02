package tools

import (
	"reflect"
)

func FlatMap(source map[string]interface{}) map[string]interface{} {
	return flatMap(source, "")
}

func flatMap(source map[string]interface{}, prefix string) map[string]interface{} {
	result := map[string]interface{}{}
	prefixEnh := ""
	if prefix != "" {
		prefixEnh += prefix + "."
	}
	for key, value := range source {
		switch reflect.ValueOf(value).Kind() {
		case reflect.Map:
			// TODO: Merge two maps result =  FlatMap(value.(map[string]interface{}))
			resultInline := flatMap(value.(map[string]interface{}), prefixEnh+key)
			result = mergeMap(result, resultInline)
		default:
			result[prefixEnh+key] = value
		}
	}
	return result
}

func mergeMap(source map[string]interface{}, destination map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	var firstKeyType reflect.Kind
	for key, value := range source {
		firstKeyType = reflect.ValueOf(key).Kind()
		result[key] = value
	}
	for key, value := range destination {
		if reflect.ValueOf(key).Kind() != firstKeyType {
			return result // TODO: make error
		}
		result[key] = value
	}
	return result
}
