// Package hocon is a small, hand-written parser for the practical subset of
// HOCON gostructor needs: JSON-shaped objects/arrays/strings/numbers/bools,
// plus the common HOCON extras: an optional root object (no outer `{ }`
// required), unquoted keys and bareword string values, `=` as an alias for
// `:`, `#`/`//` comments, and the on/off/yes/no boolean literals.
//
// Not supported: `${}` substitutions, `include` statements, duration/size
// unit literals (e.g. `10m`, `5 MB`), and string concatenation across values.
// A file needing those must avoid them or use a different tool.
package hocon

import (
	"fmt"
	"strconv"
)

// Parse parses a HOCON-subset document into a nested
// map[string]interface{}, in the same shape encoding/json.Unmarshal would
// produce for the equivalent JSON (numbers as float64, objects as
// map[string]interface{}, arrays as []interface{}).
func Parse(data []byte) (map[string]any, error) {
	p := &parser{lex: newLexer(string(data))}
	p.advance()
	p.skipSeparators()

	if p.tok.kind == tokLBrace {
		p.advance()
		obj, err := p.parseMembers(tokRBrace)
		if err != nil {
			return nil, err
		}
		if p.tok.kind != tokRBrace {
			return nil, p.errorf("expected '}'")
		}
		p.advance()
		p.skipSeparators()
		if p.tok.kind != tokEOF {
			return nil, p.errorf("unexpected trailing content after root object")
		}
		return obj, nil
	}

	result, err := p.parseMembers(tokEOF)
	if err != nil {
		return nil, err
	}
	if p.tok.kind != tokEOF {
		return nil, p.errorf("unexpected trailing content")
	}
	return result, nil
}

type parser struct {
	lex *lexer
	tok token
}

func (p *parser) advance() {
	p.tok = p.lex.next()
}

func (p *parser) errorf(format string, args ...any) error {
	return fmt.Errorf("hocon: line %d: %s", p.tok.line, fmt.Sprintf(format, args...))
}

func (p *parser) skipSeparators() {
	for p.tok.kind == tokComma || p.tok.kind == tokNewline {
		p.advance()
	}
}

// parseMembers parses `key (: | =) value` pairs, separated by commas and/or
// newlines, until a token of kind end is reached.
func (p *parser) parseMembers(end tokenKind) (map[string]any, error) {
	result := map[string]any{}
	p.skipSeparators()
	for p.tok.kind != end && p.tok.kind != tokEOF {
		key, err := p.parseKey()
		if err != nil {
			return nil, err
		}
		if p.tok.kind != tokAssign {
			return nil, p.errorf("expected ':' or '=' after key %q", key)
		}
		p.advance()
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		result[key] = value
		p.skipSeparators()
	}
	return result, nil
}

func (p *parser) parseKey() (string, error) {
	switch p.tok.kind {
	case tokString, tokBareword:
		key := p.tok.literal
		p.advance()
		return key, nil
	default:
		return "", p.errorf("expected a key")
	}
}

func (p *parser) parseValue() (any, error) {
	switch p.tok.kind {
	case tokLBrace:
		p.advance()
		obj, err := p.parseMembers(tokRBrace)
		if err != nil {
			return nil, err
		}
		if p.tok.kind != tokRBrace {
			return nil, p.errorf("expected '}'")
		}
		p.advance()
		return obj, nil
	case tokLBracket:
		return p.parseArray()
	case tokString:
		s := p.tok.literal
		p.advance()
		return s, nil
	case tokBareword:
		s := p.tok.literal
		p.advance()
		return interpretBareword(s), nil
	default:
		return nil, p.errorf("expected a value")
	}
}

func (p *parser) parseArray() (any, error) {
	p.advance() // consume '['
	result := []any{}
	p.skipSeparators()
	for p.tok.kind != tokRBracket && p.tok.kind != tokEOF {
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		result = append(result, value)
		p.skipSeparators()
	}
	if p.tok.kind != tokRBracket {
		return nil, p.errorf("expected ']'")
	}
	p.advance()
	return result, nil
}

func interpretBareword(s string) any {
	switch s {
	case "true", "yes", "on":
		return true
	case "false", "no", "off":
		return false
	case "null":
		return nil
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}
