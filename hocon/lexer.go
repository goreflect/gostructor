package hocon

import (
	"strings"
	"unicode"
)

type tokenKind int

const (
	tokEOF tokenKind = iota
	tokLBrace
	tokRBrace
	tokLBracket
	tokRBracket
	tokAssign // ':' or '='
	tokComma
	tokNewline
	tokString   // quoted
	tokBareword // unquoted run of characters
)

type token struct {
	kind    tokenKind
	literal string
	line    int
}

type lexer struct {
	input []rune
	pos   int
	line  int
}

func newLexer(input string) *lexer {
	return &lexer{input: []rune(input), line: 1}
}

func (l *lexer) next() token {
	l.skipInsignificant()
	if l.pos >= len(l.input) {
		return token{kind: tokEOF, line: l.line}
	}
	ch := l.input[l.pos]
	line := l.line
	switch ch {
	case '\n':
		l.pos++
		l.line++
		return token{kind: tokNewline, line: line}
	case '{':
		l.pos++
		return token{kind: tokLBrace, line: line}
	case '}':
		l.pos++
		return token{kind: tokRBrace, line: line}
	case '[':
		l.pos++
		return token{kind: tokLBracket, line: line}
	case ']':
		l.pos++
		return token{kind: tokRBracket, line: line}
	case ',':
		l.pos++
		return token{kind: tokComma, line: line}
	case ':', '=':
		l.pos++
		return token{kind: tokAssign, line: line}
	case '"':
		return l.lexString()
	default:
		return l.lexBareword()
	}
}

// skipInsignificant skips spaces/tabs and comments (# and // to end of
// line), but NOT newlines, which are a significant member separator.
func (l *lexer) skipInsignificant() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		switch {
		case ch == ' ' || ch == '\t' || ch == '\r':
			l.pos++
		case ch == '#':
			l.skipToLineEnd()
		case ch == '/' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '/':
			l.skipToLineEnd()
		default:
			return
		}
	}
}

func (l *lexer) skipToLineEnd() {
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.pos++
	}
}

func (l *lexer) lexString() token {
	line := l.line
	l.pos++ // opening quote
	var sb strings.Builder
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '"' {
			l.pos++
			return token{kind: tokString, literal: sb.String(), line: line}
		}
		if ch == '\\' && l.pos+1 < len(l.input) {
			l.pos++
			sb.WriteRune(unescape(l.input[l.pos]))
			l.pos++
			continue
		}
		if ch == '\n' {
			l.line++
		}
		sb.WriteRune(ch)
		l.pos++
	}
	return token{kind: tokString, literal: sb.String(), line: line}
}

func unescape(ch rune) rune {
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

func isBarewordBoundary(ch rune) bool {
	if unicode.IsSpace(ch) {
		return true
	}
	switch ch {
	case '{', '}', '[', ']', ':', '=', ',', '"', '#':
		return true
	}
	return false
}

func (l *lexer) lexBareword() token {
	line := l.line
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '/' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '/' {
			break
		}
		if isBarewordBoundary(ch) {
			break
		}
		l.pos++
	}
	literal := strings.TrimSpace(string(l.input[start:l.pos]))
	return token{kind: tokBareword, literal: literal, line: line}
}
