package tags

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_Scan(t *testing.T) {
	reader := strings.NewReader(`node=context1;path=test1;functions=function1(0,1),function2(sda)`)
	parsedSlice := NewParser(reader).Parse()
	assert.Equal(t,
		[]ReturnSlice{
			ReturnSlice{
				Tok:     CUSTOMPARAMNODE,
				Literal: DefineNameNode,
			},
			ReturnSlice{
				Tok:     EQUAL,
				Literal: string(DefineEqual),
			},
			ReturnSlice{
				Tok:     VALUE,
				Literal: string("context1"),
			},
			ReturnSlice{
				Tok:     SEMICOLON,
				Literal: string(DefineSemicolon),
			},
			ReturnSlice{
				Tok:     CUSTOMPARAMPATH,
				Literal: DefineNamePath,
			},
			ReturnSlice{
				Tok:     EQUAL,
				Literal: string(DefineEqual),
			},
			ReturnSlice{
				Tok:     VALUE,
				Literal: string("test1"),
			},
			ReturnSlice{
				Tok:     SEMICOLON,
				Literal: string(DefineSemicolon),
			},
			ReturnSlice{
				Tok:     CUSTOMPARAMFUNCTIONS,
				Literal: DefineNameFunctions,
			},
			ReturnSlice{
				Tok:     EQUAL,
				Literal: string(DefineEqual),
			},
			ReturnSlice{
				Tok:     VALUE,
				Literal: string("function1"),
			},
			ReturnSlice{
				Tok:     LEFTBRACKET,
				Literal: string(DefineLeftBracket),
			},
			ReturnSlice{
				Tok:     VALUE,
				Literal: string("0"),
			},
			ReturnSlice{
				Tok:     COMMA,
				Literal: string(DefineComma),
			},
			ReturnSlice{
				Tok:     VALUE,
				Literal: string("1"),
			},
			ReturnSlice{
				Tok:     RIGHTBRACKET,
				Literal: string(DefienRightBracket),
			},
			ReturnSlice{
				Tok:     COMMA,
				Literal: string(DefineComma),
			},
			ReturnSlice{
				Tok:     VALUE,
				Literal: string("function2"),
			},
			ReturnSlice{
				Tok:     LEFTBRACKET,
				Literal: string(DefineLeftBracket),
			},
			ReturnSlice{
				Tok:     VALUE,
				Literal: string("sda"),
			},
			ReturnSlice{
				Tok:     RIGHTBRACKET,
				Literal: string(DefienRightBracket),
			},
		},
		parsedSlice,
	)
}
