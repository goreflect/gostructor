package hocon_test

import (
	"reflect"
	"testing"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/hocon"
)

type myStruct struct {
	Field1 []string  `cfg:"field1,hocon:MyStruct.Field1"`
	Field2 []int32   `cfg:"field2,hocon:MyStruct.field2"`
	Field3 []float32 `cfg:"field3,hocon:MyStruct.Field3"`
	Field4 []bool    `cfg:"field4,hocon:MyStruct.field4"`
}

func TestHOCONSourceEndToEndSlices(t *testing.T) {
	cfg, err := gostructor.Configure(&myStruct{}, gostructor.WithSources(hocon.File("../testdata/test1.hocon")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(cfg.Field1, []string{"test1", "test2", "test3"}) {
		t.Errorf("Field1 = %v", cfg.Field1)
	}
	if !reflect.DeepEqual(cfg.Field2, []int32{112312323, 2, 123123123, 4}) {
		t.Errorf("Field2 = %v", cfg.Field2)
	}
	if !reflect.DeepEqual(cfg.Field3, []float32{1.2, 1.5, 1.7, 11.2}) {
		t.Errorf("Field3 = %v", cfg.Field3)
	}
	if !reflect.DeepEqual(cfg.Field4, []bool{true, false, false, true}) {
		t.Errorf("Field4 = %v", cfg.Field4)
	}
}

type nestedFromHocon struct {
	Nested   string         `cfg:"nested,hocon:planC.tururu.tratatat.planZ.Field1"`
	MapValue map[string]int `cfg:"mapValue,hocon:TestHocon.myMap"`
	Base     int            `cfg:"base,hocon:TestHocon.myBaseType"`
}

func TestHOCONSourceEndToEndDeepNestingAndMap(t *testing.T) {
	cfg, err := gostructor.Configure(&nestedFromHocon{}, gostructor.WithSources(hocon.File("../testdata/testmap.hocon")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Nested != "testValueByTest" {
		t.Errorf("Nested = %q", cfg.Nested)
	}
	want := map[string]int{"test1": 1, "test2": 2}
	if !reflect.DeepEqual(cfg.MapValue, want) {
		t.Errorf("MapValue = %v, want %v", cfg.MapValue, want)
	}
	if cfg.Base != 1 {
		t.Errorf("Base = %d", cfg.Base)
	}
}

func TestHOCONSourceMissingFileEnvVar(t *testing.T) {
	type cfgT struct {
		Value string `cfg:"value,hocon:x"`
	}
	_, err := gostructor.Configure(&cfgT{}, gostructor.WithSources(hocon.New()))
	if err == nil {
		t.Fatal("expected an error when GOSTRUCTOR_HOCON is unset and no explicit path was given")
	}
}
