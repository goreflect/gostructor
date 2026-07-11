package converters

import (
	"reflect"
	"testing"
)

func TestConvertBetweenComplexTypesSliceOfMixedSourceKinds(t *testing.T) {
	// Mirrors what encoding/json and go-yaml hand back for a parsed list:
	// []interface{} with numbers decoded as float64, not string.
	source := []interface{}{float64(1), float64(2), float64(3)}
	destination := reflect.ValueOf([]int{})

	got := ConvertBetweenComplexTypes(reflect.ValueOf(source), destination)
	if got.GetNotAValue() != nil {
		t.Fatalf("unexpected error: %v", got.GetNotAValue().Error)
	}
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got.Value.Interface(), want) {
		t.Errorf("got %v, want %v", got.Value.Interface(), want)
	}
}

func TestConvertBetweenComplexTypesMap(t *testing.T) {
	source := map[string]interface{}{
		"a": "1",
		"b": "2",
	}
	destination := reflect.ValueOf(map[string]int{})

	got := ConvertBetweenComplexTypes(reflect.ValueOf(source), destination)
	if got.GetNotAValue() != nil {
		t.Fatalf("unexpected error: %v", got.GetNotAValue().Error)
	}
	want := map[string]int{"a": 1, "b": 2}
	if !reflect.DeepEqual(got.Value.Interface(), want) {
		t.Errorf("got %v, want %v", got.Value.Interface(), want)
	}
}

func TestConvertBetweenComplexTypesMapWrongSourceKind(t *testing.T) {
	destination := reflect.ValueOf(map[string]int{})
	got := ConvertBetweenComplexTypes(reflect.ValueOf([]string{"a"}), destination)
	if got.GetNotAValue() == nil {
		t.Fatal("expected an error when source is not a map")
	}
}
