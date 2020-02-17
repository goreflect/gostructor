package pipeline

import (
	"errors"
	"fmt"
)

type EnvironmentConfig struct {
}

func (config EnvironmentConfig) GetComplexType(context *structContext) GoStructorValue {
	fmt.Println("Level: Debug. environment values sources start")

	return NewGoStructorNoValue(context.Value.Interface(), errors.New("getComplexType not implement for environment configuring"))
}

func (config EnvironmentConfig) GetBaseType(context *structContext) GoStructorValue {
	fmt.Println("Level: Debug. environment values sources start")

	return NewGoStructorNoValue(context.Value.Interface(), errors.New("getBaseType not implement for environment configuring"))
}
