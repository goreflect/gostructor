// Package yaml resolves gostructor fields from a YAML file, via
// goccy/go-yaml. Full YAML support is a large, subtle undertaking (block
// and flow styles, anchors and aliases, implicit typing, ...), so unlike
// gostructor's hand-rolled INI/HOCON/TOML parsers, YAML deliberately stays
// on a mature, actively-maintained third-party parser rather than
// reimplementing the spec.
package yaml

import (
	"fmt"
	"os"
	"sync"

	goyaml "github.com/goccy/go-yaml"
	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/tools"
)

// Tag is the struct tag this source responds to: `cf_yaml:"server.host"`.
// `cf_yaml:"server"` addresses the whole nested "server" object, for
// map[string]T destination fields.
const Tag = "cf_yaml"

// FileEnvVar names the environment variable New reads its file path from,
// unless a path was given explicitly to File.
const FileEnvVar = "GOSTRUCTOR_YAML"

type source struct {
	fileName string
	once     sync.Once
	data     map[string]any
	loadErr  error
}

// New resolves fields from the YAML file named by the GOSTRUCTOR_YAML
// environment variable.
func New() gostructor.Source { return &source{} }

// File is like New but reads from path instead of the GOSTRUCTOR_YAML
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
			s.loadErr = fmt.Errorf("gostructor/yaml: cf_yaml used but neither an explicit path nor %s is set", FileEnvVar)
			return
		}
		buf, err := tools.ReadFromFile(fileName)
		if err != nil {
			s.loadErr = fmt.Errorf("gostructor/yaml: reading file %q: %w", fileName, err)
			return
		}
		parsed := map[string]any{}
		if err := goyaml.Unmarshal(buf.Bytes(), &parsed); err != nil {
			s.loadErr = fmt.Errorf("gostructor/yaml: parsing file %q: %w", fileName, err)
			return
		}
		s.data = parsed
	})
	return s.loadErr
}
