package gostructor

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/goreflect/gostructor/tools"
)

// JSONTag is the struct tag JSON responds to: `cf_json:"server.host"`.
const JSONTag = "cf_json"

// JSONFileEnvVar names the environment variable JSON reads its file path
// from, unless a path was given explicitly to JSONFile.
const JSONFileEnvVar = "GOSTRUCTOR_JSON"

type jsonSource struct {
	fileName string
	once     sync.Once
	data     map[string]any
	loadErr  error
}

// JSON resolves fields from the JSON file named by the GOSTRUCTOR_JSON
// environment variable. `cf_json:"server.host"` addresses a nested "host"
// key inside a "server" object; `cf_json:"server"` addresses the whole
// nested object, for map[string]T destination fields.
func JSON() Source { return &jsonSource{} }

// JSONFile is like JSON but reads from path instead of the
// GOSTRUCTOR_JSON environment variable.
func JSONFile(path string) Source { return &jsonSource{fileName: path} }

func (*jsonSource) Tag() string { return JSONTag }

func (s *jsonSource) Resolve(field FieldContext) (any, bool, error) {
	if err := s.load(); err != nil {
		return nil, false, err
	}
	name := field.TagValue(JSONTag)
	if name == "" {
		return nil, false, nil
	}
	value, found := tools.LookupPath(s.data, name)
	if !found || value == nil {
		return nil, false, nil
	}
	return value, true, nil
}

func (s *jsonSource) load() error {
	s.once.Do(func() {
		fileName := s.fileName
		if fileName == "" {
			fileName = os.Getenv(JSONFileEnvVar)
		}
		if fileName == "" {
			s.loadErr = fmt.Errorf("gostructor: cf_json used but neither an explicit path nor %s is set", JSONFileEnvVar)
			return
		}
		buf, err := tools.ReadFromFile(fileName)
		if err != nil {
			s.loadErr = fmt.Errorf("gostructor: reading JSON file %q: %w", fileName, err)
			return
		}
		parsed := map[string]any{}
		if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
			s.loadErr = fmt.Errorf("gostructor: parsing JSON file %q: %w", fileName, err)
			return
		}
		s.data = parsed
	})
	return s.loadErr
}
