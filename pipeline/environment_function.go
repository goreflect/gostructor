package pipeline

import (
	"errors"
	"fmt"

	"github.com/goreflect/gostructor/infra"
)

type EnvironmentConfig struct {
}

func (config EnvironmentConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Message: environment values sources start")

	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getComplexType not implemented for environment configuring"))
}

func (config EnvironmentConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Message: environment values sources start")

	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getBaseType not implemented for environment configuring"))
}
