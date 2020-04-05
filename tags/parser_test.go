package tags

import (
	"strings"
	"testing"
)

func TestParser_Scan(t *testing.T) {
	reader := strings.NewReader(`node=context1   ;   path = test1;  functions=function1(0,1),  function2(sda)`)
	ast, errParsing := NewParser(reader).Parse()
	if errParsing != nil {
		t.Error(errParsing)
	}
	t.Log(ast)
	// t.Error()
	// assert.Equal(t,
	// 	[]TerminalSymbol{
	// 		TerminalSymbol{
	// 			Tok:            CUSTOMPARAMNODE,
	// 			Literal:        DefineNameNode,
	// 			StartPositioin: 1,
	// 			EndPosition:    4,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            EQUAL,
	// 			Literal:        string(DefineEqual),
	// 			StartPositioin: 5,
	// 			EndPosition:    5,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            VALUE,
	// 			Literal:        string("context1"),
	// 			StartPositioin: 6,
	// 			EndPosition:    13,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            SEMICOLON,
	// 			Literal:        string(DefineSemicolon),
	// 			StartPositioin: 14,
	// 			EndPosition:    14,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            CUSTOMPARAMPATH,
	// 			Literal:        DefineNamePath,
	// 			StartPositioin: 15,
	// 			EndPosition:    18,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            EQUAL,
	// 			Literal:        string(DefineEqual),
	// 			StartPositioin: 19,
	// 			EndPosition:    19,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            VALUE,
	// 			Literal:        string("test1"),
	// 			StartPositioin: 20,
	// 			EndPosition:    24,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            SEMICOLON,
	// 			Literal:        string(DefineSemicolon),
	// 			StartPositioin: 25,
	// 			EndPosition:    25,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            CUSTOMPARAMFUNCTIONS,
	// 			Literal:        DefineNameFunctions,
	// 			StartPositioin: 26,
	// 			EndPosition:    34,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            EQUAL,
	// 			Literal:        string(DefineEqual),
	// 			StartPositioin: 35,
	// 			EndPosition:    35,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            VALUE,
	// 			Literal:        string("function1"),
	// 			StartPositioin: 36,
	// 			EndPosition:    44,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            LEFTBRACKET,
	// 			Literal:        string(DefineLeftBracket),
	// 			StartPositioin: 45,
	// 			EndPosition:    45,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            VALUE,
	// 			Literal:        string("0"),
	// 			StartPositioin: 46,
	// 			EndPosition:    46,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            COMMA,
	// 			Literal:        string(DefineComma),
	// 			StartPositioin: 47,
	// 			EndPosition:    47,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            VALUE,
	// 			Literal:        string("1"),
	// 			StartPositioin: 48,
	// 			EndPosition:    48,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            RIGHTBRACKET,
	// 			Literal:        string(DefienRightBracket),
	// 			StartPositioin: 49,
	// 			EndPosition:    49,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            COMMA,
	// 			Literal:        string(DefineComma),
	// 			StartPositioin: 50,
	// 			EndPosition:    50,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            VALUE,
	// 			Literal:        string("function2"),
	// 			StartPositioin: 51,
	// 			EndPosition:    59,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            LEFTBRACKET,
	// 			Literal:        string(DefineLeftBracket),
	// 			StartPositioin: 60,
	// 			EndPosition:    60,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            VALUE,
	// 			Literal:        string("sda"),
	// 			StartPositioin: 61,
	// 			EndPosition:    63,
	// 		},
	// 		TerminalSymbol{
	// 			Tok:            RIGHTBRACKET,
	// 			Literal:        string(DefienRightBracket),
	// 			StartPositioin: 64,
	// 			EndPosition:    64,
	// 		},
	// 	},
	// 	parsedSlice,
	// )
}
