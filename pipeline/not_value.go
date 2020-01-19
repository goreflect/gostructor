package pipeline

// NotAValue - value for check setup in fields
type NotAValue struct {
	ValueAddress interface{}
}

// NewNotAValue - create new not a value field interface for check true or false setuping value from source
func NewNotAValue(field interface{}) *NotAValue {
	return &NotAValue{
		ValueAddress: field,
	}
}
