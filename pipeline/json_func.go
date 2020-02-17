package pipeline

import (
	"errors"
	"fmt"
)

type JsonConfig struct {
}

func (json JsonConfig) GetComplexType(context *structContext) GoStructorValue {
	fmt.Println("Level: Debug. Json configurator source start.")
	return NewGoStructorNoValue(context.Value.Interface(), errors.New("getcomplext type from json not implement"))
}

func (json JsonConfig) GetBaseType(context *structContext) GoStructorValue {
	fmt.Println("Level: Debug. Json configurator source start.")
	return NewGoStructorNoValue(context.Value.Interface(), errors.New("getbase type from json not implement"))
}
