package infra

import "reflect"

type (
	/*GoStructorValue - it's main type which using by this library. Current preparing field contain in this type and if field structure doesn't prepared that will be contain special noValue interface.
	 */
	GoStructorValue struct {
		Value     reflect.Value
		notAValue *NotAValue
	}

	// NotAValue - special error interface
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

/*CheckIsValue - check that inserted in GoStructorValue is valid value
// TODO: Upgrade in future
*/
func (gostructvalue GoStructorValue) CheckIsValue() bool {
	return gostructvalue.Value.Kind() != reflect.Invalid
}

/*GetNotAValue - getting interface of can not install value*/
func (gostructvalue GoStructorValue) GetNotAValue() *NotAValue {
	return gostructvalue.notAValue
}
