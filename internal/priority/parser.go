package priority

import (
	"errors"
	"io"
)

type (
	/*Parser - parser sequence token*/
	Parser struct {
		Lexer  *Scanner
		Buffer struct {
			AmountLetters int
		}
		Term TerminalSymbol
	}

	TerminalSymbol struct {
		Tok            Token
		Literal        string
		StartPositioin int
		EndPosition    int
	}

	AST struct {
		Nodes []Node
	}

	Node struct {
		Priority  *TerminalSymbol
		TagOrders []TerminalSymbol
	}
)

func NewParser(r io.Reader) *Parser {
	return &Parser{
		Lexer: NewScanner(r),
	}
}

/*
Scan - scanning by little level abstraction
*/
func (parser *Parser) Scan() {
	if parser.Buffer.AmountLetters != 0 {
		parser.Buffer.AmountLetters = 0
		return
	}
	for {
		token, literal, startPos, endPos := parser.Lexer.Scan()
		if token == WHITESPACE {
			continue
		}
		parser.Term = TerminalSymbol{
			Tok:            token,
			Literal:        literal,
			StartPositioin: startPos,
			EndPosition:    endPos,
		}
		break
	}
}

/*
Parse - parsing expression into ast tree
*/
func (parser *Parser) Parse() (*AST, error) {
	result := AST{}
	resultOne := Node{}
	parsedPriority := false
	parser.Scan()
	for {
		switch parser.Term.Tok {
		case SEMICOLON:
			parsedPriority = false
			result.Nodes = append(result.Nodes, resultOne)
			resultOne = Node{}
			parser.Scan()
			continue
		case EOF:
			result.Nodes = append(result.Nodes, resultOne)
			return &result, nil
		case VALUE:
			if !parsedPriority {
				keyNode, err := parser.parsePriority()
				if err != nil {
					return nil, err
				}
				resultOne.Priority = &keyNode
				parsedPriority = true
				continue
			}
			keyNode, err := parser.parseOrderTag()
			if err != nil {
				return nil, err
			}
			resultOne.TagOrders = append(resultOne.TagOrders, keyNode)
			continue
		case COLON:
			parser.Scan()
			continue
		case COMMA:
			parser.Scan()
			continue
		default:
			return nil, errors.New("can not parsed current expression")
		}
	}
}

func (parser *Parser) parseOrderTag() (result TerminalSymbol, err error) {
	result = TerminalSymbol{}
	for {
		switch parser.Term.Tok {
		case VALUE:
			result = parser.Term
			parser.Scan()
			return
		case SEMICOLON:
			parser.Scan()
			continue
		case COLON:
			parser.Scan()
			continue
		default:
			err = errors.New("not valid syntax for key node")
			return
		}
	}
}

func (parser *Parser) parsePriority() (result TerminalSymbol, err error) {
	result = TerminalSymbol{}
	parsedPriority := false
	for {
		switch parser.Term.Tok {
		case VALUE:
			if parsedPriority {
				err = errors.New("already parsed path ")
				return
			}
			result = parser.Term
			result.Tok = PRIORITY
			parser.Scan()
			return
		case SEMICOLON:
			parser.Scan()
			continue
		default:
			err = errors.New("not valid syntax for key node")
			return
		}
	}
}
