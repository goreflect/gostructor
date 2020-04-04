package tags

import "io"

type (
	Parser struct {
		Lexer  *Scanner
		Buffer struct {
			Tok           Token
			Literal       string
			StartPosition int
			EndPosition   int
			AmountLetters int
		}
	}

	ReturnSlice struct {
		Tok            Token
		Literal        string
		StartPositioin int
		EndPosition    int
	}
)

func NewParser(r io.Reader) *Parser {
	return &Parser{
		Lexer: NewScanner(r),
	}
}

func (parser *Parser) Scan() (Token, string, int, int) {
	if parser.Buffer.AmountLetters != 0 {
		parser.Buffer.AmountLetters = 0
		return parser.Buffer.Tok, parser.Buffer.Literal, 0, 0
	}
	token, literal, startPos, endPos := parser.Lexer.Scan()
	parser.Buffer.Tok = token
	parser.Buffer.Literal = literal
	parser.Buffer.StartPosition = startPos
	parser.Buffer.EndPosition = endPos
	return parser.Buffer.Tok, parser.Buffer.Literal, parser.Buffer.StartPosition, parser.Buffer.EndPosition
}

func (parser *Parser) Parse() []ReturnSlice {
	result := []ReturnSlice{}

	for {
		token, literal, start, end := parser.Scan()
		if token == EOF {
			break
		}
		result = append(result, ReturnSlice{
			Tok:            token,
			Literal:        literal,
			StartPositioin: start,
			EndPosition:    end,
		})
	}
	return result
}
