package convert

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
	"time"
)

// checkParity asserts a typed helper's (value, error) result is identical to
// what the reflective Value would return for the same input and destination
// type — same value, same error text — which is the equivalence the generated
// fast path relies on.
func checkParity(t *testing.T, name string, in any, gotVal any, gotErr error, destType reflect.Type) {
	t.Helper()
	wantV, wantErr := Value(reflect.ValueOf(in), destType)
	if (gotErr == nil) != (wantErr == nil) {
		t.Errorf("%s(%#v): error presence mismatch: typed=%v reflective=%v", name, in, gotErr, wantErr)
		return
	}
	if gotErr != nil {
		if gotErr.Error() != wantErr.Error() {
			t.Errorf("%s(%#v): error text mismatch:\n typed=%q\n reflective=%q", name, in, gotErr, wantErr)
		}
		return
	}
	if !reflect.DeepEqual(gotVal, wantV.Interface()) {
		t.Errorf("%s(%#v): value mismatch: typed=%#v reflective=%#v", name, in, gotVal, wantV.Interface())
	}
}

func TestTypedParity(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		for _, in := range []any{"42", "-7", "notint", int64(5), float64(3.0), float64(3.5), uint64(math.MaxUint64), json.Number("99")} {
			v, err := Int(in)
			checkParity(t, "Int", in, v, err, intType)
		}
	})
	t.Run("Int8", func(t *testing.T) {
		for _, in := range []any{"100", "200", "-200", 42} {
			v, err := Int8(in)
			checkParity(t, "Int8", in, v, err, int8Type)
		}
	})
	t.Run("Uint", func(t *testing.T) {
		for _, in := range []any{"5", "-1", int(-3), float64(-2), float64(4)} {
			v, err := Uint(in)
			checkParity(t, "Uint", in, v, err, uintType)
		}
	})
	t.Run("Float64", func(t *testing.T) {
		for _, in := range []any{"3.5", "bad", int64(7), float64(1.25)} {
			v, err := Float64(in)
			checkParity(t, "Float64", in, v, err, float64Type)
		}
	})
	t.Run("Float32", func(t *testing.T) {
		for _, in := range []any{"3.5", "1e40", float64(2.5)} {
			v, err := Float32(in)
			checkParity(t, "Float32", in, v, err, float32Type)
		}
	})
	t.Run("String", func(t *testing.T) {
		for _, in := range []any{"hi", int64(9), true, float64(2.5), json.Number("12")} {
			v, err := String(in)
			checkParity(t, "String", in, v, err, stringType)
		}
	})
	t.Run("Bool", func(t *testing.T) {
		for _, in := range []any{"true", "0", "nope", true} {
			v, err := Bool(in)
			checkParity(t, "Bool", in, v, err, boolType)
		}
	})
	t.Run("Duration", func(t *testing.T) {
		for _, in := range []any{"1h30m", "100", "bad", int64(time.Second)} {
			v, err := Duration(in)
			checkParity(t, "Duration", in, v, err, durationType)
		}
	})
	t.Run("StringSlice", func(t *testing.T) {
		for _, in := range []any{[]string{"a", "b"}, []any{"a", int64(1)}, []any{"x"}} {
			v, err := StringSlice(in)
			checkParity(t, "StringSlice", in, v, err, stringSlice)
		}
	})
}
