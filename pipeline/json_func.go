package pipeline

import (
	"errors"

	"github.com/goreflect/gostructor/infra"
	"github.com/sirupsen/logrus"
)

/*JSONConfig - source json configuring*/
type JSONConfig struct {
}

/*GetComplexType - get complex types like arrays, slices, maps from json source*/
func (json JSONConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	logrus.Debug("json configurator source start.")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getcomplext type from json not implemented"))
}

/*GetBaseType - gettin base type like string, int, float32...*/
func (json JSONConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	logrus.Debug("json configurator source start.")
	return infra.NewGoStructorNoValue(context.Value.Interface(), errors.New("getbase type from json not implemented"))
}
