package gostructor

import (
	"fmt"
	"testing"

	"github.com/goreflect/gostructor/pipeline"
)

type (
	MyStruct struct {
		Field1 []string
		Field2 []int32 `cf_hocon:"field2"`
		Field3 []float32
		Field4 []bool `cf_hocon:"field4"`
	}
	MyStruct2 struct {
		NestedStruct1 struct {
			Field1 string `cf_hocon:"test1"`
		} `cf_hocon:"tururur"`
		MyMap map[int]string `cf_hocon:"MyMap"`
	}

	MyStruct3 struct {
		NestedStruct2 struct {
			Field1 string `cf_hocon:"test1"`
		} `cf_hocon:"node=planB"`
	}

	MyStruct4 struct {
		NestedStruct4 struct {
			Field1 string
		} `cf_hocon:"node=planZ"`
	}
)

func Test_parseHocon1(t *testing.T) {
	myStruct, err := ConfigureSetup(&MyStruct{}, "./test_configs/test1.hocon", "", []pipeline.FuncType{pipeline.FunctionSetupHocon})
	if err != nil {
		t.Error("error while configuring: ", err)
	}
	t.Log("parsed structure: ", myStruct.(*MyStruct))
}

func Test_parseHocon(t *testing.T) {
	myStruct, err := ConfigureSetup(&MyStruct2{}, "./test_configs/testmap.hocon", "", []pipeline.FuncType{pipeline.FunctionSetupHocon})
	if err != nil {
		t.Error("error while configuring: ", err)
	}
	fmt.Println("parsed strcture: ", myStruct)
}

func Test_parseHoconWithNodeNotation(t *testing.T) {
	mystruct, err := ConfigureSetup(&MyStruct3{}, "./test_configs/testmap.hocon", "", []pipeline.FuncType{pipeline.FunctionSetupHocon})
	if err != nil {
		fmt.Println("error while configuring: ", err)
	}
	fmt.Println("parsed structure: ", mystruct)
}

func Test_parseHoconWithNodeNotation2(t *testing.T) {
	myStruct, err := ConfigureSetup(&MyStruct4{}, "./test_configs/testmap.hocon", "", []pipeline.FuncType{pipeline.FunctionSetupHocon})
	if err != nil {
		fmt.Println("error while configuring: ", err)
	}
	fmt.Println("parsed structure: ", myStruct)
}

// func Test_parseCustomStructureName(t *testing.T) {
// 	myStruct := struct {
// 		Field1 string `cf_hocon:"test1"`
// 	} `cf_hocon:"test2"`{}

// }
