package tools

import "strings"

// LookupPath descends into a nested map[string]interface{} (the shape
// produced by encoding/json.Unmarshal, go-yaml, and this module's own HOCON
// parser) following a dot-separated path, e.g. "server.host".
//
// LookupPath does not flatten or destroy nested map/slice structure along the
// way: if the path resolves to a sub-object or a list, that value is returned
// as-is, so callers can address either a single leaf value or an entire
// nested object (for map[string]T or struct destination fields) with the same
// mechanism.
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
