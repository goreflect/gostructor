package pipeline

import "reflect"

type (
	GoStructorValue struct {
		Value     reflect.Value
		notAValue *NotAValue
	}

	// NotAValue - value for check setup in fields
	NotAValue struct {
		ValueAddress interface{}
		Error        error
	}
)

// NewNotAValue - create new not a value field interface for check true or false setuping value from source
func NewNotAValue(field interface{}, err error) *NotAValue {
	return &NotAValue{
		ValueAddress: field,
		Error:        err,
	}
}

func NewGoStructorTrueValue(value reflect.Value) GoStructorValue {
	return GoStructorValue{
		Value: value,
	}
}

func NewGoStructorNoValue(value interface{}, err error) GoStructorValue {
	return GoStructorValue{
		notAValue: NewNotAValue(value, err),
	}
}

// func (gostructvalue GoStructorValue) CheckIsValue() bool {
// 	if GoStructorValue.Value.Kind() == reflect.Invalid) {
// 		return false
// 	}
// 	return true
// }
