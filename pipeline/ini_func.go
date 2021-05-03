package pipeline

import (
	"errors"

	"github.com/goreflect/gostructor/infra"
	"github.com/sirupsen/logrus"
)

/*IniConfig - source toml for configuring*/
type IniConfig struct {
	fileName string
}

/*GetComplexType - getting from yaml slices, maps, arrays...*/
func (yaml IniConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	logrus.Debug("yaml configurator source run")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getcomplex type from yaml not implemented"))
}

/*GetBaseType - getting from yaml string, int, float32 ...*/
func (yaml IniConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	logrus.Debug("yaml configurator source run")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("get base type from yaml not implemented"))
}
