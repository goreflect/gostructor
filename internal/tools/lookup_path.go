package tools

import "strings"

// LookupPath descends into a nested map[string]interface{} (the shape produced
// by encoding/json.Unmarshal, go-yaml, and the HOCON parser) following a
// dot-separated path, e.g. "server.host".
//
// It preserves nested structure: if the path resolves to a sub-object or a
// list, that value is returned as-is, so the same call can address a single
// leaf or a whole nested object (for map[string]T or struct fields).
func LookupPath(data map[string]interface{}, path string) (interface{}, bool) {
	var current interface{} = data
	for _, segment := range strings.Split(path, ".") {
		asMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = asMap[segment]
		if !ok {
			return nil, false
		}
	}
	return current, true
}
