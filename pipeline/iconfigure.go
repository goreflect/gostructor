package pipeline

import "github.com/goreflect/gostructor/infra"

// IConfigure - configurer interface for chain pipeline configuration
type IConfigure interface {
	GetComplexType(*structContext) infra.GoStructorValue
	GetBaseType(*structContext) infra.GoStructorValue
}
