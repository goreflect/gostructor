package pipeline

import (
	"errors"

	"github.com/goreflect/gostructor/infra"
	"github.com/sirupsen/logrus"
)

/*YamlConfig - source yaml for configuring*/
type YamlConfig struct {
}

/*GetComplexType - getting from yaml slices, maps, arrays...*/
func (yaml YamlConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	logrus.Debug("yaml configurator source run")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getcomplex type from yaml not implemented"))
}

/*GetBaseType - getting from yaml string, int, float32 ...*/
func (yaml YamlConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	logrus.Debug("yaml configurator source run")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("get base type from yaml not implemented"))
}
