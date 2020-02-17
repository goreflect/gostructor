package pipeline

import (
	"errors"
	"fmt"
)

type DefaultConfig struct {
}

func (config DefaultConfig) GetComplexType(context *structContext) GoStructorValue {
	fmt.Println("Level: Debug. default values sources start")
	return NewGoStructorNoValue(context.Value.Interface(), errors.New("getComplexType not implement for default configuring"))
}

func (config DefaultConfig) GetBaseType(context *structContext) GoStructorValue {
	fmt.Println("Level: Debug. default values sources start")
	return NewGoStructorNoValue(context.Value.Interface(), errors.New("getBaseType not implement for default configuring"))
}
