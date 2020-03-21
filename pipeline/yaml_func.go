package pipeline

import (
	"errors"
	"fmt"

	"github.com/goreflect/gostructor/infra"
)

/*YamlConfig - source yaml for configuring*/
type YamlConfig struct {
}

/*GetComplexType - getting from yaml slices, maps, arrays...*/
func (yaml YamlConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Yaml configurator source run")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getcomplex type from yaml not implemented"))
}

/*GetBaseType - getting from yaml string, int, float32 ...*/
func (yaml YamlConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	fmt.Println("Level: Debug. Yaml configurator source run")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("get base type from yaml not implemented"))
}
