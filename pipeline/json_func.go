package pipeline

import (
	"errors"
	"fmt"

	"github.com/goreflect/gostructor/infra"
)

type JsonConfig struct {
}

func (json JsonConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Json configurator source start.")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getcomplext type from json not implemented"))
}

func (json JsonConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Json configurator source start.")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getbase type from json not implemented"))
}
