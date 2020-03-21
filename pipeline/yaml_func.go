package pipeline

import (
	"errors"
	"fmt"

	"github.com/goreflect/gostructor/infra"
)

type YamlConfig struct {
}

func (yaml YamlConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Yaml configurator source run")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getcomplex type from yaml not implemented"))
}

func (yaml YamlConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Yaml configurator source run")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("get base type from yaml not implemented"))
}
