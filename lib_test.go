package gostructor

import (
	"os"
	"testing"

	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/pipeline"
	"github.com/goreflect/gostructor/tags"
	"github.com/sirupsen/logrus"
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

	// MyStruct3 struct {
	// 	NestedStruct2 struct {
	// 		Field1 string `cf_hocon:"test1"`
	// 	} `cf_hocon:"node=planB"`
	// }

	// MyStruct4 struct {
	// 	NestedStruct4 struct {
	// 		Field1 string
	// 	} `cf_hocon:"node=planC.tururu.tratatat.planZ"`
	// }

	EnvStruct struct {
		Field1 int16   `cf_env:"myField1"`
		Field2 string  `cf_env:"myField2"`
		Field3 bool    `cf_env:"myField3"`
		Field4 float32 `cf_env:"myField4"`
		Field5 []bool  `cf_env:"myField5"`
	}

	ManySourceStrategies struct {
		Field1 int16  `cf_env:"myField1" cf_hocon:"field2" cf_default:"14"`
		Field2 string `cf_hocon:"field2" cf_default:"test_tratata" cf_env:"test_tururu"`
	}
)

func Test_parseHocon1(t *testing.T) {
	os.Setenv(tags.HoconFile, "./test_configs/test1.hocon")
	myStruct, err := ConfigureSetup(&MyStruct{}, pipeline.EmptyAdditionalPrefix, []infra.FuncType{infra.FunctionSetupHocon})
	if err != nil {
		t.Error("error while configurig: ", err)
		return
	}
	assert.Equal(t, &MyStruct{
		Field1: []string{"test1", "test2", "test3"},
		Field2: []int32{112312323, 2, 123123123, 4},
		Field3: []float32{1.2, 1.5, 1.7, 11.2},
		Field4: []bool{true, false, false, true},
	}, myStruct.(*MyStruct))
}

func Test_parseHocon(t *testing.T) {
	os.Setenv(tags.HoconFile, "./test_configs/testmap.hocon")
	myStruct, err := ConfigureSetup(&MyStruct2{}, pipeline.EmptyAdditionalPrefix, []infra.FuncType{infra.FunctionSetupHocon})
	if err != nil {
		t.Error("error while configuring: ", err)
		return
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

func Test_getValueFromEnvironment(t *testing.T) {
	os.Setenv("myField1", "12")
	os.Setenv("myField2", "test")
	os.Setenv("myField3", "true")
	os.Setenv("myField4", "12.2")
	os.Setenv("myField5", "true,false,true")
	defer func() {
		os.Unsetenv("myField1")
		os.Unsetenv("myField2")
		os.Unsetenv("myField3")
		os.Unsetenv("myField4")
		os.Unsetenv("myField5")
	}()
	myStruct, err := ConfigureSmart(&EnvStruct{})
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, &EnvStruct{
		Field1: 12,
		Field2: "test",
		Field3: true,
		Field4: 12.2,
		Field5: []bool{true, false, true},
	}, myStruct.(*EnvStruct))
}

func Test_configureEasy(t *testing.T) {
	os.Setenv("myField1", "12")
	os.Setenv("myField2", "test")
	os.Setenv("myField3", "true")
	os.Setenv("myField4", "12.2")
	os.Setenv("myField5", "true,false,true")
	defer func() {
		os.Unsetenv("myField1")
		os.Unsetenv("myField2")
		os.Unsetenv("myField3")
		os.Unsetenv("myField4")
		os.Unsetenv("myField5")
	}()
	myStruct, err := ConfigureEasy(&EnvStruct{})
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, &EnvStruct{
		Field1: 12,
		Field2: "test",
		Field3: true,
		Field4: 12.2,
		Field5: []bool{true, false, true},
	}, myStruct.(*EnvStruct))

}

func TestChangeLogsParams(t *testing.T) {
	ChangeLogLevel(logrus.DebugLevel)
	ChangeLogFormatter(&logrus.JSONFormatter{})
}

func Test_covarianceSources(t *testing.T) {
	os.Setenv("test_tururu", "tururum")
	defer func() {
		os.Unsetenv("test_tururu")
	}()

	myStruct, err := ConfigureSmart(&ManySourceStrategies{})
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(myStruct)
}
