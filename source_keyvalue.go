package gostructor

import (
	"fmt"
	"os"
	"sync"

	"github.com/goreflect/gostructor/internal/tools"
)

// SourceKeyValue is the key/value source's identity and its cfg per-source
// override key, e.g. `cfg:"host,keyvalue:DATABASE_HOST"`.
const SourceKeyValue = "keyvalue"

// KeyValueFileEnvVar names the environment variable KeyValue reads its file path
// from, unless a path was given explicitly to KeyValueFile.
const KeyValueFileEnvVar = "GOSTRUCTOR_KEYVALUE"

type keyValueSource struct {
	fileName string
	once     sync.Once
	data     map[string]string
	loadErr  error
}

// KeyValue resolves fields from a flat key/value file (the `.env`/`.properties`
// shape: `KEY=VALUE` lines, `#`/`;` comments, optional `export`, quoted values)
// named by the GOSTRUCTOR_KEYVALUE environment variable. A field's key is its
// base name as written (`cfg:"host"` -> the `host` key); pin a different key
// with a per-source override (`cfg:"host,keyvalue:DB_HOST"`). Values are strings;
// a slice/array field's value is split on the field separator.
func KeyValue() Source { return &keyValueSource{} }

// KeyValueFile is like KeyValue but reads from path instead of the
// GOSTRUCTOR_KEYVALUE environment variable.
func KeyValueFile(path string) Source { return &keyValueSource{fileName: path} }

func (*keyValueSource) Name() string { return SourceKeyValue }

func (s *keyValueSource) Resolve(field FieldContext) (any, bool, error) {
	key := field.SourceKey(SourceKeyValue, Identity)
	if key == "" {
		return nil, false, nil
	}
	if err := s.load(); err != nil {
		return nil, false, err
	}
	value, found := s.data[key]
	if !found {
		return nil, false, nil
	}
	return splitIfSlice(field, value), true, nil
}

func (s *keyValueSource) load() error {
	s.once.Do(func() {
		fileName := s.fileName
		if fileName == "" {
			fileName = os.Getenv(KeyValueFileEnvVar)
		}
		if fileName == "" {
			s.loadErr = fmt.Errorf("gostructor: keyvalue source used but neither an explicit path nor %s is set", KeyValueFileEnvVar)
			return
		}
		buf, err := tools.ReadFromFile(fileName)
		if err != nil {
			s.loadErr = fmt.Errorf("gostructor: reading key/value file %q: %w", fileName, err)
			return
		}
		parsed, err := parseKeyValue(buf.Bytes())
		if err != nil {
			s.loadErr = fmt.Errorf("gostructor: parsing key/value file %q: %w", fileName, err)
			return
		}
		s.data = parsed
	})
	return s.loadErr
}
