package tags

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

type (
	/*
		Scanner - lexical scanner
	*/
	Scanner struct {
		r *bufio.Reader
	}

	Token int
)

const (
	WRONG Token = iota
	EOF
	WHITESPACE

	VALUE           // value which recognized by name field in environment or any other sources
	CUSTOMPARAMNODE // node, path, functions
	CUSTOMPARAMPATH
	CUSTOMPARAMFUNCTIONS
	CUSTOMPARAMFUNCTION
	CUSTOMPARAMTYPE

	// MISC characters
	COMMA        // ,
	EQUAL        // =
	SEMICOLON    // ;
	LEFTBRACKET  // (
	RIGHTBRACKET // )

	DefineNameFunctions = "functions"
	DefineNameFunction  = "function"
	DefineNameNode      = "node"
	DefineNamePath      = "path"
	DefineNameType      = "type"
	DefineComma         = ','
	DefineEqual         = '='
	DefineSemicolon     = ';'
	DefineLeftBracket   = '('
	DefienRightBracket  = ')'
)

var eof = rune(1)

/*
NewScanner - initialize new scanner
*/
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r: bufio.NewReader(r),
	}
}

func (scanner *Scanner) read() rune {
	ch, _, err := scanner.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (scanner *Scanner) unread() {
	_ = scanner.r.UnreadRune()
}

func isWhiteSpace(char rune) bool {
	return char == ' ' || char == '\t' || char == '\n'
}

func isLetter(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')
}

/*
Scan - start lexer scanner anbd tokenize all sequences
*/
func (scanner *Scanner) Scan() (Token, string) {
	char := scanner.read()

	if isWhiteSpace(char) {
		scanner.unread()
		return scanner.scanWhiteSpace()
	}
	if isLetter(char) {
		scanner.unread()
		return scanner.scanID()
	}

	switch char {
	case eof:
		return EOF, ""
	case DefineComma:
		return COMMA, string(",")
	case DefineEqual:
		return EQUAL, string("=")
	case DefineSemicolon:
		return SEMICOLON, string(";")
	case DefineLeftBracket:
		return LEFTBRACKET, string("(")
	case DefienRightBracket:
		return RIGHTBRACKET, string(")")
	}

	return WRONG, string(char)
}

func (scanner *Scanner) scanWhiteSpace() (Token, string) {
	var buf bytes.Buffer
	buf.WriteRune(scanner.read())

	for {
		if char := scanner.read(); char == eof {
			break
		} else if !isWhiteSpace(char) {
			scanner.unread()
			break
		} else {
			buf.WriteRune(char)
		}
	}
	return WHITESPACE, buf.String()
}

func (scanner *Scanner) scanID() (Token, string) {
	var buf bytes.Buffer
	buf.WriteRune(scanner.read())

	for {
		if char := scanner.read(); char == eof {
			break
		} else if !isLetter(char) {
			scanner.unread()
			break
		} else {
			buf.WriteRune(char)
		}
	}

	switch strings.ToLower(buf.String()) {
	case DefineNameNode:
		return CUSTOMPARAMNODE, buf.String()
	case DefineNamePath:
		return CUSTOMPARAMPATH, buf.String()
	case DefineNameFunctions:
		return CUSTOMPARAMFUNCTIONS, buf.String()
	case DefineNameFunction:
		return CUSTOMPARAMFUNCTION, buf.String()
	case DefineNameType:
		return CUSTOMPARAMTYPE, buf.String()
	}
	return VALUE, buf.String()
}
