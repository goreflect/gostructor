package pipeline

import (
	"errors"
	"fmt"
)

type YamlConfig struct {
}

func (yaml YamlConfig) GetComplexType(context *structContext) GoStructorValue {
	fmt.Println("Level: Debug. Yaml configurator source run")
	return NewGoStructorNoValue(context.Value.Interface(), errors.New("getcomplex type from yaml not implemented"))
}

func (yaml YamlConfig) GetBaseType(context *structContext) GoStructorValue {
	fmt.Println("Level: Debug. Yaml configurator source run")
	return NewGoStructorNoValue(context.Value.Interface(), errors.New("get base type from yaml not implemented"))
}
