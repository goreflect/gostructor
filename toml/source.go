package toml

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/internal/tools"
)

// Name is this source's identity, used as the per-source override key in a cfg
// tag: `cfg:"port,toml:server.net#port"`. Without an override the key is the
// field's base name, a top-level key outside any table.
const Name = "toml"

// FileEnvVar names the environment variable New reads its file path from,
// unless a path was given explicitly to File.
const FileEnvVar = "GOSTRUCTOR_TOML"

type source struct {
	fileName string
	once     sync.Once
	doc      Document
	loadErr  error
}

// New resolves fields from the TOML file named by the GOSTRUCTOR_TOML
// environment variable.
func New() gostructor.Source { return &source{} }

// File is like New but reads from path instead of the GOSTRUCTOR_TOML
// environment variable.
func File(path string) gostructor.Source { return &source{fileName: path} }

func (*source) Name() string { return Name }

func (s *source) Resolve(field gostructor.FieldContext) (any, bool, error) {
	tagValue := field.SourceKey(Name, gostructor.Identity)
	if tagValue == "" {
		return nil, false, nil
	}
	if err := s.load(); err != nil {
		return nil, false, err
	}
	tablePath, key := "", tagValue
	if idx := strings.LastIndexByte(tagValue, '#'); idx >= 0 {
		tablePath, key = tagValue[:idx], tagValue[idx+1:]
	}

	table := map[string]any(s.doc)
	if tablePath != "" {
		value, found := tools.LookupPath(table, tablePath)
		if !found {
			return nil, false, nil
		}
		nested, ok := value.(Document)
		if !ok {
			return nil, false, nil
		}
		table = map[string]any(nested)
	}
	value, found := table[key]
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
			s.loadErr = fmt.Errorf("gostructor/toml: toml source used but neither an explicit path nor %s is set", FileEnvVar)
			return
		}
		buf, err := tools.ReadFromFile(fileName)
		if err != nil {
			s.loadErr = fmt.Errorf("gostructor/toml: reading file %q: %w", fileName, err)
			return
		}
		parsed, err := Parse(buf.Bytes())
		if err != nil {
			s.loadErr = fmt.Errorf("gostructor/toml: parsing file %q: %w", fileName, err)
			return
		}
		s.doc = parsed
	})
	return s.loadErr
}
