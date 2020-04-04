package syntax

import (
	"io"

	"github.com/goreflect/gostructor/tags"
)

type SyntaxAnalys struct {
	Parser *tags.Parser
}

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
