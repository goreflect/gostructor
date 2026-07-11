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
			nested, ok := value.(map[string]interface{})
			if !ok {
				result[prefixEnh+key] = value
				continue
			}
			for flatKey, flatValue := range flatMap(nested, prefixEnh+key) {
				result[flatKey] = flatValue
			}
		default:
			result[prefixEnh+key] = value
		}
	}
	return result
}
