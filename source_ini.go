package gostructor

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/goreflect/gostructor/internal/format/ini"
	"github.com/goreflect/gostructor/internal/tools"
)

// INIFileEnvVar names the environment variable INI reads its file path
// from, unless a path was given explicitly to INIFile.
const INIFileEnvVar = "GOSTRUCTOR_INI"

type iniSource struct {
	fileName string
	once     sync.Once
	file     *ini.File
	loadErr  error
}

// INI resolves fields from the INI file named by the GOSTRUCTOR_INI
// environment variable. The key is the field's base name in the global,
// section-less part of the file (`cfg:"port"`); a value inside a section is
// addressed by overriding with "section#key" (`cfg:"port,ini:server#port"`).
func INI() Source { return &iniSource{} }

// INIFile is like INI but reads from path instead of the GOSTRUCTOR_INI
// environment variable.
func INIFile(path string) Source { return &iniSource{fileName: path} }

func (*iniSource) Name() string { return SourceINI }

func (s *iniSource) Resolve(field FieldContext) (any, bool, error) {
	tagValue := field.SourceKey(SourceINI, Identity)
	if tagValue == "" {
		return nil, false, nil
	}
	if err := s.load(); err != nil {
		return nil, false, err
	}
	section, key := ini.GlobalSection, tagValue
	if idx := strings.IndexByte(tagValue, '#'); idx >= 0 {
		section, key = tagValue[:idx], tagValue[idx+1:]
	}
	value, found := s.file.Get(section, key)
	if !found {
		return nil, false, nil
	}
	return splitIfSlice(field, value), true, nil
}

func (s *iniSource) load() error {
	s.once.Do(func() {
		fileName := s.fileName
		if fileName == "" {
			fileName = os.Getenv(INIFileEnvVar)
		}
		if fileName == "" {
			s.loadErr = fmt.Errorf("gostructor: ini source used but neither an explicit path nor %s is set", INIFileEnvVar)
			return
		}
		buf, err := tools.ReadFromFile(fileName)
		if err != nil {
			s.loadErr = fmt.Errorf("gostructor: reading INI file %q: %w", fileName, err)
			return
		}
		parsed, err := ini.Parse(buf.Bytes())
		if err != nil {
			s.loadErr = fmt.Errorf("gostructor: parsing INI file %q: %w", fileName, err)
			return
		}
		s.file = parsed
	})
	return s.loadErr
}
