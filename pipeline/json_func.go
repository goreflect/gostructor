package pipeline

import (
	"errors"
	"fmt"

	"github.com/goreflect/gostructor/infra"
)

/*JSONConfig - source json configuring*/
type JSONConfig struct {
}

/*GetComplexType - get complex types like arrays, slices, maps from json source*/
func (json JSONConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Json configurator source start.")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getcomplext type from json not implemented"))
}

/*GetBaseType - gettin base type like string, int, float32...*/
func (json JSONConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Json configurator source start.")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getbase type from json not implemented"))
}
