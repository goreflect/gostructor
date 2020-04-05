package tags

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
		KeyNode    *Key
		Middleware *MiddleWareNode
	}

	Node struct {
		TerminalSymbol
		NodeName TerminalSymbol
	}

	Path struct {
		TerminalSymbol
		PathName TerminalSymbol
	}

	Type struct {
		TerminalSymbol
		TypeName TerminalSymbol
	}

	Function struct {
		Params       []FunctionParam
		FunctionName TerminalSymbol
	}

	FunctionParam struct {
		TerminalSymbol
	}

	MiddleWareNode struct {
		TerminalSymbol
		Functions []Function
	}

	Key struct {
		NodeExist Node
		PathExist Path
		TypeExist Type
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
	token, literal, startPos, endPos := parser.Lexer.Scan()
	parser.Term = TerminalSymbol{
		Tok:            token,
		Literal:        literal,
		StartPositioin: startPos,
		EndPosition:    endPos,
	}
}

/*
Parse - parsing expression into ast tree
*/
func (parser *Parser) Parse() (*AST, error) {
	result := AST{}
	parsedKeyNode := false
	parsedMiddlewareNode := false
	parser.Scan()
	for {

		switch parser.Term.Tok {
		case WHITESPACE:
			parser.Scan()
			continue
		case EOF:
			return &result, nil
		case CUSTOMPARAMNODE, CUSTOMPARAMPATH, CUSTOMPARAMTYPE, VALUE:
			if !parsedKeyNode {
				keyNode, err := parser.parseValueNode()
				if err != nil {
					return nil, err
				}
				result.KeyNode = &keyNode
				parsedKeyNode = true
				continue
			}
			return nil, errors.New("in expression should be at least once key node")
		case CUSTOMPARAMFUNCTIONS:
			if !parsedMiddlewareNode {
				middleNode, err := parser.parseMiddlewareNode()
				if err != nil {
					return nil, err
				}
				result.Middleware = &middleNode
				parsedMiddlewareNode = true
				continue
			}
			return nil, errors.New("in expression should be at least once middleware node")
		default:
			return nil, errors.New("can not parsed current expression")
		}
	}
}

func (parser *Parser) parseValueNode() (result Key, err error) {
	result = Key{}
	parsedNode := false
	parsedPath := false
	parsedType := false
	for {
		switch parser.Term.Tok {
		case WHITESPACE:
			parser.Scan()
			continue
		case CUSTOMPARAMNODE:
			if parsedNode {
				err = errors.New("have more that 1 node token")
				return
			}
			node, errParsing := parser.parseNode()
			if errParsing != nil {
				err = errParsing
				return
			}
			result.NodeExist = node
			parsedNode = true
			parser.Scan()
		case CUSTOMPARAMPATH:
			if parsedPath {
				err = errors.New("already parsed path ")
				return
			}
			path, errParsing := parser.parsePath()
			if errParsing != nil {
				err = errParsing
				return
			}
			result.PathExist = path
			parsedPath = true
			parser.Scan()
		case CUSTOMPARAMTYPE:
			if parsedType {
				err = errors.New("already parsed type source")
				return
			}
			typ, errParsing := parser.parseType()
			if errParsing != nil {
				err = errParsing
				return
			}
			result.TypeExist = typ
			parsedType = true
			parser.Scan()
		case VALUE:
			if parsedPath {
				err = errors.New("already parsed path ")
				return
			}
			path, errParsing := parser.parsePathAsSingleValue()
			if errParsing != nil {
				err = errParsing
				return
			}
			result.PathExist = path
			parser.Scan()
		case SEMICOLON:
			parser.Scan()
			continue
		default:
			if parsedNode && parsedPath {
				return
			}
			if parsedPath && !parsedNode {
				return
			}
			if parsedType && parsedPath {
				return
			}
			err = errors.New("not valid syntax for key node")
			return
		}
	}
}

func (parser *Parser) parseNode() (node Node, err error) {
	node = Node{}
	node.Literal = parser.Term.Literal
	node.Tok = parser.Term.Tok
	node.StartPositioin = parser.Term.StartPositioin
	node.EndPosition = parser.Term.EndPosition
	parser.Scan()
	for {
		switch parser.Term.Tok {
		case WHITESPACE:
			parser.Scan()
			continue
		case EQUAL:
			parser.Scan()
			if parser.Term.Tok == VALUE {
				node.NodeName = parser.Term
				return
			}
			err = errors.New("error while parse NodeValue. Check That valid value")
			return
		default:
			err = errors.New("not valid expression. After literal node should be =")
			return
		}
	}
}

func (parser *Parser) parsePath() (path Path, err error) {
	path = Path{}
	path.Literal = parser.Term.Literal
	path.Tok = parser.Term.Tok
	path.StartPositioin = parser.Term.StartPositioin
	path.EndPosition = parser.Term.EndPosition
	parser.Scan()
	for {
		switch parser.Term.Tok {
		case WHITESPACE:
			parser.Scan()
			continue
		case EQUAL:
			parser.Scan()
			if parser.Term.Tok == WHITESPACE {
				parser.Scan()
			}
			if parser.Term.Tok == VALUE {
				path.PathName = parser.Term
				return
			}
			err = errors.New("error while parse PathValue. Check that valid value")
			return
		default:
			err = errors.New("not valid expression. After literal path should be =")
			return
		}
	}
}

func (parser *Parser) parseType() (typ Type, err error) {
	typ = Type{}
	typ.Literal = parser.Term.Literal
	typ.StartPositioin = parser.Term.StartPositioin
	typ.EndPosition = parser.Term.EndPosition
	typ.Tok = parser.Term.Tok
	parser.Scan()
	for {
		switch parser.Term.Tok {
		case WHITESPACE:
			parser.Scan()
			continue
		case EQUAL:
			parser.Scan()
			if parser.Term.Tok == VALUE {
				typ.TypeName = parser.Term
				return
			}
			err = errors.New("error while parse TypeName. Check that valid value")
			return
		default:
			err = errors.New("not valid expression. After literal type should be =")
			return
		}
	}
}

func (parser *Parser) parsePathAsSingleValue() (path Path, err error) {
	path = Path{}
	path.PathName = parser.Term
	path.Tok = parser.Term.Tok
	path.Literal = parser.Term.Literal
	path.StartPositioin = parser.Term.StartPositioin
	path.EndPosition = parser.Term.EndPosition
	return
}

func (parser *Parser) parseMiddlewareNode() (middle MiddleWareNode, err error) {
	middle = MiddleWareNode{}
	middle.Literal = parser.Term.Literal
	middle.Tok = parser.Term.Tok
	middle.StartPositioin = parser.Term.StartPositioin
	middle.EndPosition = parser.Term.EndPosition
	functions := []Function{}
	useEqual := false
	parser.Scan()
	for {
		switch parser.Term.Tok {
		case WHITESPACE:
			parser.Scan()
			continue
		case EQUAL:
			useEqual = true
			parser.Scan()
			function, errParseFunc := parser.parseFunction()
			if errParseFunc != nil {
				err = errParseFunc
				return
			}
			functions = append(functions, function)
		case COMMA:
			if useEqual {
				parser.Scan()
				function, errParseFunc := parser.parseFunction()
				if errParseFunc != nil {
					err = errParseFunc
					return
				}
				functions = append(functions, function)
				continue
			}
			err = errors.New("for parsing function should be after =")
			return
		case EOF:
			middle.Functions = functions
			return
		default:
			err = errors.New("error while parsing function. After token functions should be equal sym = ")
			return
		}
	}
}

func (parser *Parser) parseFunction() (function Function, err error) {
	function = Function{}
	if parser.Term.Tok == WHITESPACE {
		parser.Scan()
	}
	if parser.Term.Tok == VALUE {
		function.FunctionName = parser.Term
		parser.Scan()
		funcParams, errParseParams := parser.parseFunctionParams()
		if errParseParams != nil {
			err = errParseParams
			return
		}
		function.Params = funcParams
		return
	}
	err = errors.New("after comma or equal should by value name of function")
	return
}

func (parser *Parser) parseFunctionParams() (funcParams []FunctionParam, err error) {
	result := []FunctionParam{}
	useLeftBracket := false
	for {
		switch parser.Term.Tok {
		case WHITESPACE:
			parser.Scan()
			continue
		case LEFTBRACKET:
			useLeftBracket = true
			parser.Scan()
			funcParam, errParseFuncParam := parser.parseFunctionParam()
			if errParseFuncParam != nil {
				err = errParseFuncParam
				return
			}
			result = append(result, funcParam)
			parser.Scan()
		case COMMA:
			if useLeftBracket {
				parser.Scan()
				funcParam, errParseFuncParam := parser.parseFunctionParam()
				if errParseFuncParam != nil {
					err = errParseFuncParam
					return
				}
				result = append(result, funcParam)
				parser.Scan()
				continue
			}
			err = errors.New("before separating function params should be start from left brecket")
			return
		case RIGHTBRACKET:
			funcParams = result
			parser.Scan()
			return
		}
	}
}

func (parser *Parser) parseFunctionParam() (funcParam FunctionParam, err error) {
	funcParam = FunctionParam{}
	if parser.Term.Tok == WHITESPACE {
		parser.Scan()
	}
	if parser.Term.Tok == VALUE {
		funcParam.Tok = parser.Term.Tok
		funcParam.Literal = parser.Term.Literal
		funcParam.StartPositioin = parser.Term.StartPositioin
		funcParam.EndPosition = parser.Term.EndPosition
		return
	}
	err = errors.New("can not recognize value as function param")
	return
}
