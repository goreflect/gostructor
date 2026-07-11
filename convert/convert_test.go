package convert

import (
	"reflect"
	"testing"
)

func TestToPrimitiveIntFromString(t *testing.T) {
	got, err := ToPrimitive(reflect.ValueOf("42"), reflect.ValueOf(int(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Interface().(int) != 42 {
		t.Errorf("got %v, want 42", got.Interface())
	}
}

func TestToPrimitiveIntFromFloat64(t *testing.T) {
	// This is what encoding/json and go-yaml hand back for numbers.
	got, err := ToPrimitive(reflect.ValueOf(float64(42.9)), reflect.ValueOf(int(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Interface().(int) != 42 {
		t.Errorf("got %v, want 42 (truncated)", got.Interface())
	}
}

func TestToPrimitiveIntFromUint64(t *testing.T) {
	// goccy/go-yaml decodes positive integer literals as uint64.
	got, err := ToPrimitive(reflect.ValueOf(uint64(7)), reflect.ValueOf(int(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Interface().(int) != 7 {
		t.Errorf("got %v, want 7", got.Interface())
	}
}

func TestToPrimitiveStringFromInt(t *testing.T) {
	got, err := ToPrimitive(reflect.ValueOf(7), reflect.ValueOf(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Interface().(string) != "7" {
		t.Errorf("got %v, want \"7\"", got.Interface())
	}
}

func TestToPrimitiveBoolFromString(t *testing.T) {
	got, err := ToPrimitive(reflect.ValueOf("true"), reflect.ValueOf(false))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Interface().(bool) != true {
		t.Errorf("got %v, want true", got.Interface())
	}
}

func TestToPrimitiveUnsupportedSource(t *testing.T) {
	_, err := ToPrimitive(reflect.ValueOf(struct{}{}), reflect.ValueOf(0))
	if err == nil {
		t.Fatal("expected an error for an unsupported source kind")
	}
}

func TestToPrimitiveUnsupportedDestination(t *testing.T) {
	_, err := ToPrimitive(reflect.ValueOf("x"), reflect.ValueOf(struct{}{}))
	if err == nil {
		t.Fatal("expected an error for an unsupported destination kind")
	}
}

func TestToComplexSliceMixedSourceKinds(t *testing.T) {
	source := []interface{}{float64(1), float64(2), float64(3)}
	got, err := ToComplex(reflect.ValueOf(source), reflect.ValueOf([]int{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got.Interface(), want) {
		t.Errorf("got %v, want %v", got.Interface(), want)
	}
}

func TestToComplexMap(t *testing.T) {
	source := map[string]interface{}{"a": "1", "b": "2"}
	got, err := ToComplex(reflect.ValueOf(source), reflect.ValueOf(map[string]int{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]int{"a": 1, "b": 2}
	if !reflect.DeepEqual(got.Interface(), want) {
		t.Errorf("got %v, want %v", got.Interface(), want)
	}
}

func TestToComplexMapWrongSourceKind(t *testing.T) {
	_, err := ToComplex(reflect.ValueOf([]string{"a"}), reflect.ValueOf(map[string]int{}))
	if err == nil {
		t.Fatal("expected an error when source is not a map")
	}
}

func TestToComplexUnsupportedDestination(t *testing.T) {
	_, err := ToComplex(reflect.ValueOf([]interface{}{1}), reflect.ValueOf(0))
	if err == nil {
		t.Fatal("expected an error for a non-slice/map/array destination")
	}
}
