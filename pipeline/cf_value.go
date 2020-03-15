package pipeline

import "reflect"

type (
	/*
		GoStructorValue - it's main type which using by this library. In this type setuping current preparing field and if field doesn't preparing in this type will be inserted special noValue interface.
	*/
	GoStructorValue struct {
		Value     reflect.Value
		notAValue *NotAValue
	}

	// NotAValue - specials setuping value
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

/*NewGoStructorTrueValue - generate new GoStructorValue with completed preparing field
 */
func NewGoStructorTrueValue(value reflect.Value) GoStructorValue {
	return GoStructorValue{
		Value: value,
	}
}

/*NewGoStructorNoValue - generate new GoStructorValue with error handling value
 */
func NewGoStructorNoValue(value interface{}, err error) GoStructorValue {
	return GoStructorValue{
		notAValue: NewNotAValue(value, err),
	}
}

/*CheckIsValue - check that inserted in GoStructorValue is valid value*/
func (gostructvalue GoStructorValue) CheckIsValue() bool {
	return gostructvalue.Value.Kind() != reflect.Invalid
}

func (goStructorValue GoStructorValue) GetNotAValue() *NotAValue {
	return goStructorValue.notAValue
}
