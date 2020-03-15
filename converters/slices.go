package converters

import (
	"errors"
	"reflect"

	"github.com/goreflect/gostructor/infra"
)

func convertSlice(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	return infra.NewGoStructorNoValue(nil, errors.New(""))
}
