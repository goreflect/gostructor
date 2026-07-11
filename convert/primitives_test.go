package convert

import (
	"reflect"
	"testing"
)

func TestToPrimitiveSizedInts(t *testing.T) {
	cases := []struct {
		name string
		dest reflect.Value
	}{
		{"int8", reflect.ValueOf(int8(0))},
		{"int16", reflect.ValueOf(int16(0))},
		{"int32", reflect.ValueOf(int32(0))},
		{"int64", reflect.ValueOf(int64(0))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ToPrimitive(reflect.ValueOf("5"), tc.dest)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Kind() != tc.dest.Kind() {
				t.Errorf("got kind %s, want %s", got.Kind(), tc.dest.Kind())
			}
		})
	}
}

func TestToPrimitiveSizedIntsFromNumericSources(t *testing.T) {
	sources := []reflect.Value{
		reflect.ValueOf(int32(5)),
		reflect.ValueOf(uint32(5)),
		reflect.ValueOf(float64(5)),
	}
	for _, source := range sources {
		got, err := ToPrimitive(source, reflect.ValueOf(int32(0)))
		if err != nil {
			t.Fatalf("source kind %s: unexpected error: %v", source.Kind(), err)
		}
		if got.Interface().(int32) != 5 {
			t.Errorf("source kind %s: got %v, want 5", source.Kind(), got.Interface())
		}
	}
}

func TestToPrimitiveSizedIntBadString(t *testing.T) {
	_, err := ToPrimitive(reflect.ValueOf("not-a-number"), reflect.ValueOf(int32(0)))
	if err == nil {
		t.Fatal("expected an error for a non-numeric string")
	}
}

func TestToPrimitiveSizedUints(t *testing.T) {
	cases := []reflect.Value{
		reflect.ValueOf(uint8(0)),
		reflect.ValueOf(uint16(0)),
		reflect.ValueOf(uint32(0)),
		reflect.ValueOf(uint64(0)),
		reflect.ValueOf(uint(0)),
	}
	for _, dest := range cases {
		got, err := ToPrimitive(reflect.ValueOf("5"), dest)
		if err != nil {
			t.Fatalf("dest kind %s: unexpected error: %v", dest.Kind(), err)
		}
		if got.Kind() != dest.Kind() {
			t.Errorf("got kind %s, want %s", got.Kind(), dest.Kind())
		}
	}
}

func TestToPrimitiveSizedUintFromNumericSources(t *testing.T) {
	sources := []reflect.Value{
		reflect.ValueOf(int32(5)),
		reflect.ValueOf(uint32(5)),
		reflect.ValueOf(float64(5)),
	}
	for _, source := range sources {
		got, err := ToPrimitive(source, reflect.ValueOf(uint32(0)))
		if err != nil {
			t.Fatalf("source kind %s: unexpected error: %v", source.Kind(), err)
		}
		if got.Interface().(uint32) != 5 {
			t.Errorf("source kind %s: got %v, want 5", source.Kind(), got.Interface())
		}
	}
}

func TestToPrimitiveSizedUintBadString(t *testing.T) {
	_, err := ToPrimitive(reflect.ValueOf("nope"), reflect.ValueOf(uint32(0)))
	if err == nil {
		t.Fatal("expected an error for a non-numeric string")
	}
}

func TestToPrimitiveSizedUintUnsupportedSource(t *testing.T) {
	_, err := ToPrimitive(reflect.ValueOf(true), reflect.ValueOf(uint32(0)))
	if err == nil {
		t.Fatal("expected an error for an unsupported source kind")
	}
}

func TestToPrimitiveFloats(t *testing.T) {
	cases := []struct {
		dest reflect.Value
		want float64
	}{
		{reflect.ValueOf(float32(0)), 1.5},
		{reflect.ValueOf(float64(0)), 1.5},
	}
	for _, tc := range cases {
		got, err := ToPrimitive(reflect.ValueOf("1.5"), tc.dest)
		if err != nil {
			t.Fatalf("dest kind %s: unexpected error: %v", tc.dest.Kind(), err)
		}
		var gotFloat float64
		if tc.dest.Kind() == reflect.Float32 {
			gotFloat = float64(got.Interface().(float32))
		} else {
			gotFloat = got.Interface().(float64)
		}
		if gotFloat != tc.want {
			t.Errorf("got %v, want %v", gotFloat, tc.want)
		}
	}
}

func TestToPrimitiveFloatFromNumericSources(t *testing.T) {
	sources := []reflect.Value{
		reflect.ValueOf(int32(5)),
		reflect.ValueOf(uint32(5)),
	}
	for _, source := range sources {
		got, err := ToPrimitive(source, reflect.ValueOf(float64(0)))
		if err != nil {
			t.Fatalf("source kind %s: unexpected error: %v", source.Kind(), err)
		}
		if got.Interface().(float64) != 5 {
			t.Errorf("source kind %s: got %v, want 5", source.Kind(), got.Interface())
		}
	}
}

func TestToPrimitiveFloatBadString(t *testing.T) {
	_, err := ToPrimitive(reflect.ValueOf("nope"), reflect.ValueOf(float64(0)))
	if err == nil {
		t.Fatal("expected an error for a non-numeric string")
	}
}

func TestToPrimitiveFloatUnsupportedSource(t *testing.T) {
	_, err := ToPrimitive(reflect.ValueOf(true), reflect.ValueOf(float64(0)))
	if err == nil {
		t.Fatal("expected an error for an unsupported source kind")
	}
}

func TestToPrimitiveBoolFromBool(t *testing.T) {
	got, err := ToPrimitive(reflect.ValueOf(true), reflect.ValueOf(false))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Interface().(bool) != true {
		t.Errorf("got %v, want true", got.Interface())
	}
}

func TestToPrimitiveBoolBadString(t *testing.T) {
	_, err := ToPrimitive(reflect.ValueOf("nope"), reflect.ValueOf(false))
	if err == nil {
		t.Fatal("expected an error for a non-boolean string")
	}
}

func TestToPrimitiveBoolUnsupportedSource(t *testing.T) {
	_, err := ToPrimitive(reflect.ValueOf(5), reflect.ValueOf(false))
	if err == nil {
		t.Fatal("expected an error for an unsupported source kind")
	}
}

func TestToPrimitiveStringFromVariousSources(t *testing.T) {
	cases := []struct {
		source reflect.Value
		want   string
	}{
		{reflect.ValueOf(uint(7)), "7"},
		{reflect.ValueOf(float32(1.5)), "1.5"},
		{reflect.ValueOf(float64(1.5)), "1.5"},
		{reflect.ValueOf(true), "true"},
	}
	for _, tc := range cases {
		got, err := ToPrimitive(tc.source, reflect.ValueOf(""))
		if err != nil {
			t.Fatalf("source kind %s: unexpected error: %v", tc.source.Kind(), err)
		}
		if got.Interface().(string) != tc.want {
			t.Errorf("source kind %s: got %q, want %q", tc.source.Kind(), got.Interface(), tc.want)
		}
	}
}

func TestToPrimitiveStringUnsupportedSource(t *testing.T) {
	_, err := ToPrimitive(reflect.ValueOf(struct{}{}), reflect.ValueOf(""))
	if err == nil {
		t.Fatal("expected an error for an unsupported source kind")
	}
}
