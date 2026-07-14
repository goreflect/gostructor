// Package toml is a small, hand-written parser for the practical subset of
// TOML gostructor needs: `[table]` and `[table.sub]` headers, `key = value`
// pairs with string/int/float/bool/array-of-scalars values, and `#`
// comments.
//
// Not supported: datetimes, inline tables (`{ a = 1 }`), arrays of tables
// (`[[table]]`), multiline/triple-quoted strings, and dotted keys inline
// (`a.b = 1`) - none of which gostructor's own usage of TOML relies on.
package toml

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Document is a parsed TOML document: nested map[string]interface{}, one
// level per `[table.path]` segment, with values as string/int64/float64/
// bool/[]interface{}.
//
// This is a type alias, not a distinct named type: nested tables must be
// indistinguishable from plain map[string]interface{} values so that
// tools.LookupPath's type assertions (which check against
// map[string]interface{} exactly) can descend into them.
type Document = map[string]any

// Parse reads a TOML document from data.
func Parse(data []byte) (Document, error) {
	root := Document{}
	current := root

	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := stripComment(scanner.Text())
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") {
			if strings.HasPrefix(line, "[[") {
				return nil, fmt.Errorf("toml: line %d: arrays of tables ([[...]]) are not supported", lineNumber)
			}
			if !strings.HasSuffix(line, "]") {
				return nil, fmt.Errorf("toml: line %d: unterminated table header %q", lineNumber, line)
			}
			tablePath := strings.TrimSpace(line[1 : len(line)-1])
			if tablePath == "" {
				return nil, fmt.Errorf("toml: line %d: empty table header", lineNumber)
			}
			table, err := descendCreate(root, strings.Split(tablePath, "."))
			if err != nil {
				return nil, fmt.Errorf("toml: line %d: %w", lineNumber, err)
			}
			current = table
			continue
		}

		key, value, err := parseKeyValue(line)
		if err != nil {
			return nil, fmt.Errorf("toml: line %d: %w", lineNumber, err)
		}
		current[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("toml: %w", err)
	}
	return root, nil
}

func stripComment(line string) string {
	inString := false
	var quote byte
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if inString {
			if ch == quote {
				inString = false
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inString = true
			quote = ch
			continue
		}
		if ch == '#' {
			return line[:i]
		}
	}
	return line
}

func descendCreate(root Document, segments []string) (Document, error) {
	current := root
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			return nil, fmt.Errorf("empty table path segment")
		}
		next, exists := current[segment]
		if !exists {
			table := Document{}
			current[segment] = table
			current = table
			continue
		}
		table, ok := next.(Document)
		if !ok {
			return nil, fmt.Errorf("key %q is already defined as a non-table value", segment)
		}
		current = table
	}
	return current, nil
}

func parseKeyValue(line string) (key string, value any, err error) {
	sep := strings.IndexByte(line, '=')
	if sep < 0 {
		return "", nil, fmt.Errorf("expected 'key = value', got %q", line)
	}
	key = strings.TrimSpace(line[:sep])
	if key == "" {
		return "", nil, fmt.Errorf("empty key in %q", line)
	}
	key = unquoteKey(key)
	rest := strings.TrimSpace(line[sep+1:])
	value, remainder, err := parseValue(rest)
	if err != nil {
		return "", nil, err
	}
	if strings.TrimSpace(remainder) != "" {
		return "", nil, fmt.Errorf("unexpected trailing content %q", remainder)
	}
	return key, value, nil
}

func unquoteKey(key string) string {
	if len(key) >= 2 && (key[0] == '"' || key[0] == '\'') && key[len(key)-1] == key[0] {
		return key[1 : len(key)-1]
	}
	return key
}

// parseValue parses one value off the front of s, returning the value and
// whatever text remains after it (used to parse array elements one at a
// time from a single-line array literal).
func parseValue(s string) (value any, remainder string, err error) {
	s = strings.TrimLeft(s, " \t")
	if s == "" {
		return nil, "", fmt.Errorf("expected a value")
	}
	switch s[0] {
	case '"', '\'':
		return parseQuotedString(s)
	case '[':
		return parseArray(s)
	default:
		return parseScalar(s)
	}
}

func parseQuotedString(s string) (value any, remainder string, err error) {
	quote := s[0]
	var sb strings.Builder
	i := 1
	for i < len(s) {
		ch := s[i]
		if ch == quote {
			return sb.String(), s[i+1:], nil
		}
		if quote == '"' && ch == '\\' && i+1 < len(s) {
			i++
			sb.WriteByte(unescape(s[i]))
			i++
			continue
		}
		sb.WriteByte(ch)
		i++
	}
	return nil, "", fmt.Errorf("unterminated string %q", s)
}

func unescape(ch byte) byte {
	switch ch {
	case 'n':
		return '\n'
	case 't':
		return '\t'
	case 'r':
		return '\r'
	default:
		return ch
	}
}

func parseArray(s string) (value any, remainder string, err error) {
	s = s[1:] // consume '['
	result := []any{}
	for {
		s = strings.TrimLeft(s, " \t")
		if s == "" {
			return nil, "", fmt.Errorf("unterminated array")
		}
		if s[0] == ']' {
			return result, s[1:], nil
		}
		if s[0] == ',' {
			s = s[1:]
			continue
		}
		var elem any
		elem, s, err = parseValue(s)
		if err != nil {
			return nil, "", err
		}
		result = append(result, elem)
	}
}

func parseScalar(s string) (value any, remainder string, err error) {
	end := 0
	for end < len(s) && !isScalarBoundary(s[end]) {
		end++
	}
	token := strings.TrimSpace(s[:end])
	remainder = s[end:]
	if token == "" {
		return nil, "", fmt.Errorf("expected a value, got %q", s)
	}
	switch token {
	case "true":
		return true, remainder, nil
	case "false":
		return false, remainder, nil
	}
	if i, convErr := strconv.ParseInt(token, 10, 64); convErr == nil {
		return i, remainder, nil
	}
	if f, convErr := strconv.ParseFloat(token, 64); convErr == nil {
		return f, remainder, nil
	}
	return nil, "", fmt.Errorf("unrecognized value %q (unquoted strings must be true/false or a number)", token)
}

func isScalarBoundary(ch byte) bool {
	return ch == ',' || ch == ']' || ch == '#'
}
