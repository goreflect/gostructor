package hocon

import (
	"fmt"
	"os"
	"sync"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/internal/tools"
)

// Tag is the struct tag this source responds to:
// `cf_hocon:"MyStruct.Field1"`. Nested objects are flattened to dotted
// keys, so a value nested three objects deep is addressed the same way a
// nested YAML/JSON value would be.
const Tag = "cf_hocon"

// FileEnvVar names the environment variable New reads its file path from,
// unless a path was given explicitly to File.
const FileEnvVar = "GOSTRUCTOR_HOCON"

type source struct {
	fileName string
	once     sync.Once
	data     map[string]interface{}
	loadErr  error
}

// New resolves fields from the HOCON file named by the GOSTRUCTOR_HOCON
// environment variable.
func New() gostructor.Source { return &source{} }

// File is like New but reads from path instead of the GOSTRUCTOR_HOCON
// environment variable.
func File(path string) gostructor.Source { return &source{fileName: path} }

func (*source) Tag() string { return Tag }

func (s *source) Resolve(field gostructor.FieldContext) (any, bool, error) {
	if err := s.load(); err != nil {
		return nil, false, err
	}
	name := field.TagValue(Tag)
	if name == "" {
		return nil, false, nil
	}
	value, found := tools.LookupPath(s.data, name)
	if !found || value == nil {
		return nil, false, nil
	}
	return value, true, nil
}

func (s *source) load() error {
	s.once.Do(func() {
		fileName := s.fileName
		if fileName == "" {
			fileName = os.Getenv(FileEnvVar)
		}
		if fileName == "" {
			s.loadErr = fmt.Errorf("gostructor/hocon: cf_hocon used but neither an explicit path nor %s is set", FileEnvVar)
			return
		}
		buf, err := tools.ReadFromFile(fileName)
		if err != nil {
			s.loadErr = fmt.Errorf("gostructor/hocon: reading file %q: %w", fileName, err)
			return
		}
		parsed, err := Parse(buf.Bytes())
		if err != nil {
			s.loadErr = fmt.Errorf("gostructor/hocon: parsing file %q: %w", fileName, err)
			return
		}
		s.data = parsed
	})
	return s.loadErr
}
