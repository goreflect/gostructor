package gostructor

import (
	"fmt"
	"testing"

	"github.com/goreflect/gostructor/pipeline"
)

type (
	MyStruct struct {
		Field1 []string  `cf_hocon:"field1"`
		Field2 []int32   `cf_hocon:"field2"`
		Field3 []float32 `cf_hocon:"field3"`
		Field4 []bool    `cf_hocon:"field4"`
	}
	MyStruct2 struct {
		NestedStruct1 struct {
			Field1 string `cf_hocon:"test1"`
		} `cf_hocon:"tururur"`
	}
)

func Test_parseHocon(t *testing.T) {
	myStruct := MyStruct2{}
	ConfigureSetup(&myStruct, "./test_configs/testmap.hocon", "", []pipeline.FuncType{pipeline.FunctionSetupHocon})
	fmt.Println("parsed strcture: ", myStruct)
	t.Error()
}

// func Test_parseCustomStructureName(t *testing.T) {
// 	myStruct := struct {
// 		Field1 string `cf_hocon:"test1"`
// 	} `cf_hocon:"test2"`{}

// }
