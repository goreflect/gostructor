package gostructor

import (
	"fmt"
	"os"
	"testing"

	"github.com/goreflect/gostructor/infra"
	"github.com/stretchr/testify/assert"
)

type (
	MyStruct struct {
		Field1 []string
		Field2 []int32 `cf_hocon:"field2"`
		Field3 []float32
		Field4 []bool `cf_hocon:"field4"`
	}
	NestedStruct1 struct {
		Field1 string `cf_hocon:"test1"`
	}
	MyStruct2 struct {
		NestedStruct1 NestedStruct1  `cf_hocon:"tururur"`
		MyMap         map[int]string `cf_hocon:"MyMap"`
	}

	MyStruct3 struct {
		NestedStruct2 struct {
			Field1 string `cf_hocon:"test1"`
		} `cf_hocon:"node=planB"`
	}

	MyStruct4 struct {
		NestedStruct4 struct {
			Field1 string
		} `cf_hocon:"node=planC.tururu.tratatat.planZ"`
	}

	EnvStruct struct {
		Field1 int16   `cf_env:"myField1"`
		Field2 string  `cf_env:"myField2"`
		Field3 bool    `cf_env:"myField3"`
		Field4 float32 `cf_env:"myField4"`
		Field5 []bool  `cf_env:"myField5"`
	}
)

func Test_parseHocon1(t *testing.T) {
	myStruct, err := ConfigureSetup(&MyStruct{}, "./test_configs/test1.hocon", "", []infra.FuncType{infra.FunctionSetupHocon})
	if err != nil {
		t.Error("error while configuring: ", err)
	}
	assert.Equal(t, &MyStruct{
		Field1: []string{"test1", "test2", "test3"},
		Field2: []int32{112312323, 2, 123123123, 4},
		Field3: []float32{1.2, 1.5, 1.7, 11.2},
		Field4: []bool{true, false, false, true},
	}, myStruct.(*MyStruct))
}

func Test_parseHocon(t *testing.T) {
	myStruct, err := ConfigureSetup(&MyStruct2{}, "./test_configs/testmap.hocon", "", []infra.FuncType{infra.FunctionSetupHocon})
	if err != nil {
		t.Error("error while configuring: ", err)
	}
	assert.Equal(t, &MyStruct2{
		NestedStruct1: NestedStruct1{
			Field1: "testvalueInNestedStructure",
		},
		MyMap: map[int]string{
			1: "test",
			2: "test2",
			3: "test3",
		},
	}, myStruct.(*MyStruct2))
}

func Test_parseHoconWithNodeNotation(t *testing.T) {
	mystruct, err := ConfigureSetup(&MyStruct3{}, "./test_configs/testmap.hocon", "", []infra.FuncType{infra.FunctionSetupHocon})
	if err != nil {
		fmt.Println("error while configuring: ", err)
	}
	assert.Equal(t, &MyStruct3{
		NestedStruct2: struct {
			Field1 string "cf_hocon:\"test1\""
		}{
			Field1: "testValueByNodeInTag",
		},
	}, mystruct.(*MyStruct3))
}

func Test_parseHoconWithNodeNotation2(t *testing.T) {
	myStruct, err := ConfigureSetup(&MyStruct4{}, "./test_configs/testmap.hocon", "", []infra.FuncType{infra.FunctionSetupHocon})
	if err != nil {
		fmt.Println("error while configuring: ", err)
	}
	assert.Equal(t, &MyStruct4{
		NestedStruct4: struct{ Field1 string }{
			Field1: "testValueByTest",
		},
	}, myStruct.(*MyStruct4))
}

func Test_smartConfigure(t *testing.T) {
	myStruct, err := ConfigureSmart(&MyStruct4{}, "./test_configs/testmap.hocon")
	if err != nil {
		fmt.Println("error while configuring: ", err)
	}
	assert.Equal(t, &MyStruct4{
		NestedStruct4: struct{ Field1 string }{
			Field1: "testValueByTest",
		},
	}, myStruct.(*MyStruct4))
}

func Test_smartConfigure(t *testing.T) {
	myStruct, err := ConfigureSmart(&MyStruct4{}, "./test_configs/testmap.hocon")
	if err != nil {
		fmt.Println("error while configuring: ", err)
	}
	assert.Equal(t, &MyStruct4{
		NestedStruct4: struct{ Field1 string }{
			Field1: "testValueByTest",
		},
	}, myStruct.(*MyStruct4))
}

func Test_getValueFromEnvironment(t *testing.T) {
	os.Setenv("myField1", "12")
	os.Setenv("myField2", "test")
	os.Setenv("myField3", "true")
	os.Setenv("myField4", "12.2")
	os.Setenv("myField5", "true,false,true")
	defer func() {
		os.Remove("myField1")
		os.Remove("myField2")
		os.Remove("myField3")
		os.Remove("myField4")
		os.Remove("myField5")
	}()
	myStruct, err := ConfigureSmart(&EnvStruct{}, "")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, &EnvStruct{
		Field1: 12,
		Field2: "test",
		Field3: true,
		Field4: 12.2,
		Field5: []bool{true, false, true},
	}, myStruct.(*EnvStruct))
}
