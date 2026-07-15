package gostructor

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goreflect/gostructor/internal/format/ini"
)

// A Decoder turns a configuration file's raw bytes into a nested map, the shape
// LookupKey addresses. It is the pluggable format layer for the remote,
// file-shaped sources (gostructor/git, gostructor/watch).
//
// The core ships zero-dependency decoders — DecodeJSON, DecodeINI,
// DecodeKeyValue. The yaml/toml/hocon modules expose their own, so a format's
// decoder pulls in only that format's dependency.
type Decoder func([]byte) (map[string]any, error)

// DecodeJSON parses JSON into a nested map, keeping numbers exact (json.Number)
// so large integers survive rather than being rounded through float64 — the
// same decoding the JSON source uses.
func DecodeJSON(raw []byte) (map[string]any, error) {
	parsed := map[string]any{}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

// DecodeINI parses an INI document into a nested map: global keys at the top
// level, each `[section]` a nested map. Values are strings.
func DecodeINI(raw []byte) (map[string]any, error) {
	file, err := ini.Parse(raw)
	if err != nil {
		return nil, err
	}
	return file.ToMap(), nil
}

// DecodeKeyValue parses a flat `.env`/`.properties` file (see parseKeyValue) into
// a flat map of string values. It backs the KeyValue source and is reusable as a
// git/watch Decoder.
func DecodeKeyValue(raw []byte) (map[string]any, error) {
	flat, err := parseKeyValue(raw)
	if err != nil {
		return nil, err
	}
	out := make(map[string]any, len(flat))
	for k, v := range flat {
		out[k] = v
	}
	return out, nil
}

// parseKeyValue is the shared flat-file parser behind DecodeKeyValue and the
// KeyValue source: one `KEY=VALUE` per line, `#`/`;` comments and blank lines
// ignored, an optional leading `export `, and optionally quoted values.
func parseKeyValue(raw []byte) (map[string]string, error) {
	out := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") || strings.HasPrefix(text, ";") {
			continue
		}
		text = strings.TrimPrefix(text, "export ")
		eq := strings.IndexByte(text, '=')
		if eq < 0 {
			return nil, fmt.Errorf("gostructor: key/value line %d: expected KEY=VALUE, got %q", line, text)
		}
		key := strings.TrimSpace(text[:eq])
		if key == "" {
			return nil, fmt.Errorf("gostructor: key/value line %d: empty key in %q", line, text)
		}
		out[key] = unquote(strings.TrimSpace(text[eq+1:]))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("gostructor: reading key/value data: %w", err)
	}
	return out, nil
}

// unquote strips a single pair of matching surrounding quotes, if present.
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
