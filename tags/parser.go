package tags

import "io"

type (
	Parser struct {
		Lexer  *Scanner
		Buffer struct {
			Tok           Token
			Literal       string
			AmountLetters int
		}
	}

	ReturnSlice struct {
		Tok     Token
		Literal string
	}
)

func NewParser(r io.Reader) *Parser {
	return &Parser{
		Lexer: NewScanner(r),
	}
}

func (parser *Parser) Scan() (Token, string) {
	if parser.Buffer.AmountLetters != 0 {
		parser.Buffer.AmountLetters = 0
		return parser.Buffer.Tok, parser.Buffer.Literal
	}
	parser.Buffer.Tok, parser.Buffer.Literal = parser.Lexer.Scan()
	return parser.Buffer.Tok, parser.Buffer.Literal
}

func (parser *Parser) Parse() []ReturnSlice {
	result := []ReturnSlice{}

	for {
		token, literal := parser.Scan()
		if token == EOF {
			break
		}
		result = append(result, ReturnSlice{
			Tok:     token,
			Literal: literal,
		})
	}
	return result
}
