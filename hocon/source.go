package hocon

import (
	"fmt"
	"os"
	"sync"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/internal/tools"
)

// Name is this source's identity, used as the per-source override key in a cfg
// tag: `cfg:"field1,hocon:MyStruct.Field1"`. Nested objects are addressed by
// dotted keys, the same way a nested YAML/JSON value would be.
const Name = "hocon"

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

func (*source) Name() string { return Name }

func (s *source) Resolve(field gostructor.FieldContext) (any, bool, error) {
	name := field.SourceKey(Name, gostructor.Identity)
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

func (s *source) load() error {
	s.once.Do(func() {
		fileName := s.fileName
		if fileName == "" {
			fileName = os.Getenv(FileEnvVar)
		}
		if fileName == "" {
			s.loadErr = fmt.Errorf("gostructor/hocon: hocon source used but neither an explicit path nor %s is set", FileEnvVar)
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
