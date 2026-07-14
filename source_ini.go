package gostructor

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/goreflect/gostructor/internal/format/ini"
	"github.com/goreflect/gostructor/internal/tools"
)

// INITag is the struct tag INI responds to: `cf_ini:"section#key"` (or
// just `cf_ini:"key"` for the global, section-less part of the file).
const INITag = "cf_ini"

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
// environment variable.
func INI() Source { return &iniSource{} }

// INIFile is like INI but reads from path instead of the GOSTRUCTOR_INI
// environment variable.
func INIFile(path string) Source { return &iniSource{fileName: path} }

func (*iniSource) Tag() string { return INITag }

func (s *iniSource) Resolve(field FieldContext) (any, bool, error) {
	if err := s.load(); err != nil {
		return nil, false, err
	}
	tagValue := field.TagValue(INITag)
	if tagValue == "" {
		return nil, false, nil
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
			s.loadErr = fmt.Errorf("gostructor: cf_ini used but neither an explicit path nor %s is set", INIFileEnvVar)
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
