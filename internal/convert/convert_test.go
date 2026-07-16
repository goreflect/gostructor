package convert

import (
	"math"
	"reflect"
	"testing"
	"time"
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

func TestToPrimitiveIntFromIntegralFloat64(t *testing.T) {
	// encoding/json and go-yaml hand back float64 for numbers; an integral
	// value must convert exactly.
	got, err := ToPrimitive(reflect.ValueOf(float64(42)), reflect.ValueOf(int(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Interface().(int) != 42 {
		t.Errorf("got %v, want 42", got.Interface())
	}
}

func TestToPrimitiveIntFromFractionalFloatIsError(t *testing.T) {
	// A fractional float into an int field is data loss and must be rejected,
	// not silently truncated.
	if _, err := ToPrimitive(reflect.ValueOf(float64(42.9)), reflect.ValueOf(int(0))); err == nil {
		t.Fatal("expected an error converting 42.9 into int")
	}
}

func TestToPrimitiveNumericOverflowIsError(t *testing.T) {
	cases := []struct {
		name   string
		source reflect.Value
		dest   reflect.Value
	}{
		{"int64 300 into int8", reflect.ValueOf(int64(300)), reflect.ValueOf(int8(0))},
		{"int -1 into uint", reflect.ValueOf(int(-1)), reflect.ValueOf(uint(0))},
		{"float -1 into uint8", reflect.ValueOf(float64(-1)), reflect.ValueOf(uint8(0))},
		{"uint64 max into int64", reflect.ValueOf(uint64(math.MaxUint64)), reflect.ValueOf(int64(0))},
		{"float 1e39 into float32", reflect.ValueOf(float64(1e39)), reflect.ValueOf(float32(0))},
		{"NaN into int", reflect.ValueOf(math.NaN()), reflect.ValueOf(int(0))},
		{"Inf into int", reflect.ValueOf(math.Inf(1)), reflect.ValueOf(int(0))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ToPrimitive(tc.source, tc.dest); err == nil {
				t.Fatalf("expected an error for %s", tc.name)
			}
		})
	}
}

type celsius int

func TestToPrimitiveNamedType(t *testing.T) {
	got, err := ToPrimitive(reflect.ValueOf("21"), reflect.ValueOf(celsius(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Type() != reflect.TypeOf(celsius(0)) {
		t.Fatalf("got type %s, want celsius", got.Type())
	}
	if got.Interface().(celsius) != 21 {
		t.Errorf("got %v, want 21", got.Interface())
	}
}

func TestValueDuration(t *testing.T) {
	got, err := Value(reflect.ValueOf("1h30m"), reflect.TypeOf(time.Duration(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Interface().(time.Duration) != 90*time.Minute {
		t.Errorf("got %v, want 1h30m", got.Interface())
	}
}

func TestValueTextUnmarshaler(t *testing.T) {
	got, err := Value(reflect.ValueOf("2020-01-02T03:04:05Z"), reflect.TypeOf(time.Time{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	if !got.Interface().(time.Time).Equal(want) {
		t.Errorf("got %v, want %v", got.Interface(), want)
	}
}

func TestValueWithLayout(t *testing.T) {
	got, err := ValueWithLayout(reflect.ValueOf("2020-01-02"), timeType, "2006-01-02")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	if !got.Interface().(time.Time).Equal(want) {
		t.Errorf("got %v, want %v", got.Interface(), want)
	}
}

func TestValueWithLayoutEmptyFallsBackToRFC3339(t *testing.T) {
	// An empty layout must behave exactly like Value: RFC3339 via TextUnmarshaler.
	got, err := ValueWithLayout(reflect.ValueOf("2020-01-02T03:04:05Z"), timeType, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	if !got.Interface().(time.Time).Equal(want) {
		t.Errorf("got %v, want %v", got.Interface(), want)
	}
	// The custom layout string, not RFC3339, is now what a value must match.
	if _, err := ValueWithLayout(reflect.ValueOf("2020-01-02T03:04:05Z"), timeType, "2006-01-02"); err == nil {
		t.Error("expected an error parsing an RFC3339 string with a date-only layout")
	}
}

func TestValueWithLayoutPropagatesToSliceAndPointer(t *testing.T) {
	// []time.Time honors the field's layout for every element.
	sliceType := reflect.TypeOf([]time.Time(nil))
	got, err := ValueWithLayout(reflect.ValueOf([]any{"2020-01-02", "2021-03-04"}), sliceType, "2006-01-02")
	if err != nil {
		t.Fatalf("unexpected slice error: %v", err)
	}
	times := got.Interface().([]time.Time)
	if len(times) != 2 || !times[0].Equal(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)) || !times[1].Equal(time.Date(2021, 3, 4, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("got %v, want [2020-01-02 2021-03-04]", times)
	}

	// *time.Time honors it too.
	ptr, err := ValueWithLayout(reflect.ValueOf("2020-01-02"), reflect.TypeOf((*time.Time)(nil)), "2006-01-02")
	if err != nil {
		t.Fatalf("unexpected pointer error: %v", err)
	}
	if p := ptr.Interface().(*time.Time); p == nil || !p.Equal(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("got %v, want *2020-01-02", ptr.Interface())
	}
}

func TestValuePointer(t *testing.T) {
	got, err := Value(reflect.ValueOf("7"), reflect.TypeOf((*int)(nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p := got.Interface().(*int)
	if p == nil || *p != 7 {
		t.Errorf("got %v, want *7", got.Interface())
	}
}

func TestValueMultiLevelPointerIsError(t *testing.T) {
	if _, err := Value(reflect.ValueOf("7"), reflect.TypeOf((**int)(nil))); err == nil {
		t.Fatal("expected an error for a multi-level pointer destination")
	}
}

func TestToComplexArray(t *testing.T) {
	source := []interface{}{"1", "2", "3"}
	got, err := ToComplex(reflect.ValueOf(source), reflect.ValueOf([3]int{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Interface().([3]int) != [3]int{1, 2, 3} {
		t.Errorf("got %v, want [1 2 3]", got.Interface())
	}
}

func TestToComplexArrayLengthMismatchIsError(t *testing.T) {
	source := []interface{}{"1", "2", "3"}
	if _, err := ToComplex(reflect.ValueOf(source), reflect.ValueOf([2]int{})); err == nil {
		t.Fatal("expected an error for an array length mismatch")
	}
}

func TestToComplexNestedSlice(t *testing.T) {
	source := []interface{}{[]interface{}{"1", "2"}, []interface{}{"3"}}
	got, err := ToComplex(reflect.ValueOf(source), reflect.ValueOf([][]int{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := [][]int{{1, 2}, {3}}
	if !reflect.DeepEqual(got.Interface(), want) {
		t.Errorf("got %v, want %v", got.Interface(), want)
	}
}

type point struct {
	X int
	Y int
}

func TestValueSliceOfStructs(t *testing.T) {
	source := []interface{}{
		map[string]interface{}{"x": "1", "y": "2"},
		map[string]interface{}{"x": "3", "y": "4"},
	}
	got, err := Value(reflect.ValueOf(source), reflect.TypeOf([]point(nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []point{{1, 2}, {3, 4}}
	if !reflect.DeepEqual(got.Interface(), want) {
		t.Errorf("got %v, want %v", got.Interface(), want)
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
