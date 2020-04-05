package syntax

import (
	"io"

	"github.com/goreflect/gostructor/tags"
)

type (
	/*SyntaxAnalys - syntax analyser
	 */
	SyntaxAnalys struct {
		Parser *tags.Parser
	}

	/*AST - model for ast tree for syntax analys
	 */
	AST struct {
		Chidls []ASTChild
	}

	/*ASTChild - children for current node
	 */
	ASTChild struct {
		OneEntity tags.TerminalSymbol
		Childs    []ASTChild
	}
)

/*
NewSyntaxAnalyser - initialise new syntax analyser
*/
func NewSyntaxAnalyser(r io.Reader) *SyntaxAnalys {
	return &SyntaxAnalys{
		Parser: tags.NewParser(r),
	}
}

/*
Analys - analysing and add errors
*/
func (syntax *SyntaxAnalys) Analys(str string) []error {
	return nil
}

/*
BuildASTTree - building ast tree for analysing
*/
func (syntax *SyntaxAnalys) BuildASTTree(entries []tags.TerminalSymbol) AST {
	return AST{}
}
