package yaml

import goyaml "github.com/goccy/go-yaml"

// Decode parses YAML into a nested map, the shape gostructor.LookupKey
// addresses. It is a gostructor.Decoder, so a remote/file source (gostructor/git,
// gostructor/watch) can read YAML by passing yaml.Decode as its decoder:
//
//	git.New(git.Options{Repo: ..., Path: "config.yaml", Decoder: yaml.Decode})
func Decode(raw []byte) (map[string]any, error) {
	parsed := map[string]any{}
	if err := goyaml.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}
