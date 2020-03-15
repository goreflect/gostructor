package pipeline

import (
	"errors"
	"fmt"
)

type EnvironmentConfig struct {
}

func (config EnvironmentConfig) GetComplexType(context *structContext) GoStructorValue {
	fmt.Println("Level: Debug. Message: environment values sources start")

	return NewGoStructorNoValue(context.Value.Interface(), errors.New("getComplexType not implemented for environment configuring"))
}

func (config EnvironmentConfig) GetBaseType(context *structContext) GoStructorValue {
	fmt.Println("Level: Debug. Message: environment values sources start")

	return NewGoStructorNoValue(context.Value.Interface(), errors.New("getBaseType not implemented for environment configuring"))
}
