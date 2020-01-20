package gostructor

import (
	"testing"

	"github.com/goreflect/gostructor/pipeline"
)

type MyStruct struct {
	Field1 []*string `cf_hocon:"field1"`
	Field2 []int32   `cf_hocon:"field2"`
	Field3 []float32 `cf_hocon:"field3"`
	Field4 []bool    `cf_hocon:"field4"`
}

func Test_parseHocon(t *testing.T) {
	ConfigureSetup(&MyStruct{}, "./test_configs/test1.hocon", "", []pipeline.FuncType{pipeline.FunctionSetupHocon})
	// fmt.Println("parsed strcture: ", myStruct)
}
