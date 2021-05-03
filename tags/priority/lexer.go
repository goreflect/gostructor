package priority

import (
	"bufio"
	"bytes"
	"io"
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

	PRIORITY // value which recognized by name field in environment or any other sources
	VALUE

	// MISC characters
	COMMA     // ,
	COLON     // :
	SEMICOLON // ;

	DefineComma     = ','
	DefineColon     = ':'
	DefineSemicolon = ';'
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
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_'
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
	case DefineSemicolon:
		return SEMICOLON, string(";"), scanner.CurrentPosition, scanner.CurrentPosition
	case DefineColon:
		return COLON, string(":"), scanner.CurrentPosition, scanner.CurrentPosition
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

	return VALUE, buf.String(), startPosition, scanner.CurrentPosition
}
