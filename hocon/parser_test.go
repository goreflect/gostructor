package hocon

import (
	"reflect"
	"testing"
)

func TestParseImplicitRootObject(t *testing.T) {
	data := []byte(`
MyStruct = {
    Field1 = ["test1", "test2", "test3"]
    field2 = [112312323,2,123123123,4]
    Field3 = [1.2, 1.5, 1.7, 11.2]
    field4 = [true, false, no, on]
}`)
	got, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	myStruct, ok := got["MyStruct"].(map[string]any)
	if !ok {
		t.Fatalf("MyStruct not parsed as an object: %v", got["MyStruct"])
	}
	if !reflect.DeepEqual(myStruct["Field1"], []any{"test1", "test2", "test3"}) {
		t.Errorf("Field1 = %v", myStruct["Field1"])
	}
	if !reflect.DeepEqual(myStruct["field2"], []any{float64(112312323), float64(2), float64(123123123), float64(4)}) {
		t.Errorf("field2 = %v", myStruct["field2"])
	}
	if !reflect.DeepEqual(myStruct["Field3"], []any{1.2, 1.5, 1.7, 11.2}) {
		t.Errorf("Field3 = %v", myStruct["Field3"])
	}
	if !reflect.DeepEqual(myStruct["field4"], []any{true, false, false, true}) {
		t.Errorf("field4 = %v (want [true false false true] from [true false no on])", myStruct["field4"])
	}
}

func TestParseNestedObjectsAndMixedSeparators(t *testing.T) {
	data := []byte(`
MyStruct2 = {
    MyMap = {
        1= test,
        2= test2,
        3= test3
    }
    tururur = {
        test1 = "testvalueInNestedStructure"
    }
}

planB = {
    test1 = "testValueByNodeInTag"
}

planC = {
    tururu = {
        tratatat = {
            planZ = {
                Field1 = "testValueByTest"
            }
        }
    }
}

TestHocon = {
    myField1 = ["test1","test2"]
    myMap = {
        "test1": 1,
        "test2": 2
    }
    myBaseType = 1
    myBools = [on, off, on]
}`)
	got, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	myStruct2 := got["MyStruct2"].(map[string]any)
	myMap := myStruct2["MyMap"].(map[string]any)
	if myMap["1"] != "test" || myMap["2"] != "test2" || myMap["3"] != "test3" {
		t.Errorf("MyMap = %v", myMap)
	}
	tururur := myStruct2["tururur"].(map[string]any)
	if tururur["test1"] != "testvalueInNestedStructure" {
		t.Errorf("tururur.test1 = %v", tururur["test1"])
	}

	planB := got["planB"].(map[string]any)
	if planB["test1"] != "testValueByNodeInTag" {
		t.Errorf("planB.test1 = %v", planB["test1"])
	}

	planC := got["planC"].(map[string]any)
	tururu := planC["tururu"].(map[string]any)
	tratatat := tururu["tratatat"].(map[string]any)
	planZ := tratatat["planZ"].(map[string]any)
	if planZ["Field1"] != "testValueByTest" {
		t.Errorf("planC...Field1 = %v", planZ["Field1"])
	}

	testHocon := got["TestHocon"].(map[string]any)
	if !reflect.DeepEqual(testHocon["myField1"], []any{"test1", "test2"}) {
		t.Errorf("myField1 = %v", testHocon["myField1"])
	}
	myMap2 := testHocon["myMap"].(map[string]any)
	if myMap2["test1"] != float64(1) || myMap2["test2"] != float64(2) {
		t.Errorf("myMap = %v", myMap2)
	}
	if testHocon["myBaseType"] != float64(1) {
		t.Errorf("myBaseType = %v", testHocon["myBaseType"])
	}
	if !reflect.DeepEqual(testHocon["myBools"], []any{true, false, true}) {
		t.Errorf("myBools = %v", testHocon["myBools"])
	}
}

func TestParseExplicitRootBraces(t *testing.T) {
	got, err := Parse([]byte(`{ a = 1, b = "two" }`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["a"] != float64(1) || got["b"] != "two" {
		t.Errorf("got %v", got)
	}
}

func TestParseComments(t *testing.T) {
	data := []byte(`
# a hash comment
a = 1 // a trailing comment
// a full-line slash comment
b = 2
`)
	got, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["a"] != float64(1) || got["b"] != float64(2) {
		t.Errorf("got %v", got)
	}
}

func TestParseColonSeparator(t *testing.T) {
	got, err := Parse([]byte(`a: 1`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["a"] != float64(1) {
		t.Errorf("got %v", got)
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name string
		data string
	}{
		{"unterminated object", "a = { b = 1"},
		{"missing assign", "a 1"},
		{"unterminated array", "a = [1, 2"},
		{"trailing content after root object", "{ a = 1 } b = 2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse([]byte(tc.data)); err == nil {
				t.Errorf("expected an error for input %q", tc.data)
			}
		})
	}
}
