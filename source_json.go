package gostructor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/goreflect/gostructor/internal/tools"
)

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
// environment variable. The key is the field's base name (`cfg:"host"` -> the
// top-level "host" key); a nested value is addressed by overriding the key with
// a dotted path (`cfg:"host,json:server.host"` -> "host" inside "server").
// A bare object key like `cfg:"server"` addresses the whole nested object, for
// map[string]T destination fields.
func JSON() Source { return &jsonSource{} }

// JSONFile is like JSON but reads from path instead of the
// GOSTRUCTOR_JSON environment variable.
func JSONFile(path string) Source { return &jsonSource{fileName: path} }

func (*jsonSource) Name() string { return SourceJSON }

func (s *jsonSource) Resolve(field FieldContext) (any, bool, error) {
	name := field.SourceKey(SourceJSON, Identity)
	if name == "" {
		return nil, false, nil
	}
	if err := s.load(); err != nil {
		return nil, false, err
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
			s.loadErr = fmt.Errorf("gostructor: json source used but neither an explicit path nor %s is set", JSONFileEnvVar)
			return
		}
		buf, err := tools.ReadFromFile(fileName)
		if err != nil {
			s.loadErr = fmt.Errorf("gostructor: reading JSON file %q: %w", fileName, err)
			return
		}
		parsed := map[string]any{}
		// UseNumber decodes numbers as json.Number (a string) rather than
		// float64, so convert parses them base-10 and keeps integers exact
		// above 2^53 instead of silently truncating.
		dec := json.NewDecoder(bytes.NewReader(buf.Bytes()))
		dec.UseNumber()
		if err := dec.Decode(&parsed); err != nil {
			s.loadErr = fmt.Errorf("gostructor: parsing JSON file %q: %w", fileName, err)
			return
		}
		s.data = parsed
	})
	return s.loadErr
}
