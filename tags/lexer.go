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
		r               *bufio.Reader
		CurrentPosition int
	}

	Token int
)

const (
	WRONG Token = iota
	EOF
	WHITESPACE

	VALUE // value which recognized by name field in environment or any other sources
	FUNCTION
	PARAMFUNCTION
	PATH
	NODE
	TYPE

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
		r:               bufio.NewReader(r),
		CurrentPosition: 0,
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
func (scanner *Scanner) Scan() (Token, string, int, int) {
	char := scanner.read()
	scanner.CurrentPosition++

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
		return EOF, "", scanner.CurrentPosition, scanner.CurrentPosition
	case DefineComma:
		return COMMA, string(","), scanner.CurrentPosition, scanner.CurrentPosition
	case DefineEqual:
		return EQUAL, string("="), scanner.CurrentPosition, scanner.CurrentPosition
	case DefineSemicolon:
		return SEMICOLON, string(";"), scanner.CurrentPosition, scanner.CurrentPosition
	case DefineLeftBracket:
		return LEFTBRACKET, string("("), scanner.CurrentPosition, scanner.CurrentPosition
	case DefienRightBracket:
		return RIGHTBRACKET, string(")"), scanner.CurrentPosition, scanner.CurrentPosition
	}

	return WRONG, string(char), scanner.CurrentPosition, scanner.CurrentPosition
}

func (scanner *Scanner) scanWhiteSpace() (Token, string, int, int) {
	var buf bytes.Buffer
	startPostition := scanner.CurrentPosition
	buf.WriteRune(scanner.read())
	scanner.CurrentPosition++

	for {
		if char := scanner.read(); char == eof {
			scanner.CurrentPosition++
			break
		} else if !isWhiteSpace(char) {
			scanner.unread()
			scanner.CurrentPosition--
			break
		} else {
			buf.WriteRune(char)
			scanner.CurrentPosition++
		}
	}
	return WHITESPACE, buf.String(), startPostition, scanner.CurrentPosition
}

func (scanner *Scanner) scanID() (Token, string, int, int) {
	var buf bytes.Buffer
	startPosition := scanner.CurrentPosition
	buf.WriteRune(scanner.read())
	scanner.CurrentPosition++

	for {
		if char := scanner.read(); char == eof {
			scanner.CurrentPosition++
			break
		} else if !isLetter(char) {
			scanner.unread()
			scanner.CurrentPosition--
			break
		} else {
			buf.WriteRune(char)
			scanner.CurrentPosition++
		}
	}

	switch strings.ToLower(buf.String()) {
	case DefineNameNode:
		return CUSTOMPARAMNODE, buf.String(), startPosition, scanner.CurrentPosition
	case DefineNamePath:
		return CUSTOMPARAMPATH, buf.String(), startPosition, scanner.CurrentPosition
	case DefineNameFunctions:
		return CUSTOMPARAMFUNCTIONS, buf.String(), startPosition, scanner.CurrentPosition
	case DefineNameType:
		return CUSTOMPARAMTYPE, buf.String(), startPosition, scanner.CurrentPosition
	default:
		return VALUE, buf.String(), startPosition, scanner.CurrentPosition
	}
}
