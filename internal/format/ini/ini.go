// Package ini is a small, hand-written INI parser covering the practical
// subset gostructor needs: `[section]` headers, `key = value` or `key: value`
// pairs, `;`/`#` line comments, and quoted values. It does not support key
// interpolation (`%(name)s`), multi-line values, or duplicate-key merging
// beyond "last one wins".
package ini

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// GlobalSection is the section name for keys that appear before any
// `[section]` header.
const GlobalSection = ""

// File is a parsed INI document: section name -> key -> value.
type File struct {
	sections map[string]map[string]string
}

// Get returns the value for key in section, and whether it was present.
func (f *File) Get(section, key string) (string, bool) {
	values, ok := f.sections[section]
	if !ok {
		return "", false
	}
	value, ok := values[key]
	return value, ok
}

// Parse reads an INI document from data.
func Parse(data []byte) (*File, error) {
	file := &File{sections: map[string]map[string]string{GlobalSection: {}}}
	currentSection := GlobalSection

	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return nil, fmt.Errorf("ini: line %d: unterminated section header %q", lineNumber, line)
			}
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			if currentSection == "" {
				return nil, fmt.Errorf("ini: line %d: empty section name", lineNumber)
			}
			if _, exists := file.sections[currentSection]; !exists {
				file.sections[currentSection] = map[string]string{}
			}
			continue
		}

		key, value, err := parseKeyValue(line)
		if err != nil {
			return nil, fmt.Errorf("ini: line %d: %w", lineNumber, err)
		}
		file.sections[currentSection][key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ini: %w", err)
	}
	return file, nil
}

func parseKeyValue(line string) (key string, value string, err error) {
	sep := strings.IndexAny(line, "=:")
	if sep < 0 {
		return "", "", fmt.Errorf("expected 'key = value' or 'key: value', got %q", line)
	}
	key = strings.TrimSpace(line[:sep])
	if key == "" {
		return "", "", fmt.Errorf("empty key in %q", line)
	}
	value = strings.TrimSpace(line[sep+1:])
	value = unquote(value)
	return key, value, nil
}

func unquote(value string) string {
	if len(value) < 2 {
		return value
	}
	first, last := value[0], value[len(value)-1]
	if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
		return value[1 : len(value)-1]
	}
	return value
}
