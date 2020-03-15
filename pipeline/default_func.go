package pipeline

import (
	"errors"
	"fmt"
)

/*
DefaultConfig - one of most configuring source functions that should preparing fields in structures of data by default values setuped by special tag name
*/
type DefaultConfig struct {
}

/*
GetComplexType - get slices, maps, arrays or anything else hard types
*/
func (config DefaultConfig) GetComplexType(context *structContext) GoStructorValue {
	fmt.Println("Level: Debug. Message: default values sources start")
	return NewGoStructorNoValue(context.Value.Interface(), errors.New("getComplexType not implemented for default configuring"))
}

/*
GetBaseType - get base type from default values.
*/
func (config DefaultConfig) GetBaseType(context *structContext) GoStructorValue {
	fmt.Println("Level: Debug. Message: default values sources start")
	return NewGoStructorNoValue(context.Value.Interface(), errors.New("getBaseType not implemented for default configuring"))
}
