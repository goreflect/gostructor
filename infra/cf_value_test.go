package infra

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewGoStructorTrueValueCheckIsValue(t *testing.T) {
	value := NewGoStructorTrueValue(reflect.ValueOf("test"))
	if !value.CheckIsValue() {
		t.Error("expected CheckIsValue() to be true for a value built from a valid reflect.Value")
	}
	if value.GetNotAValue() != nil {
		t.Error("expected GetNotAValue() to be nil for a true value")
	}
}

func TestNewGoStructorNoValueCheckIsValue(t *testing.T) {
	wantErr := errors.New("boom")
	value := NewGoStructorNoValue("field-address", wantErr)
	if value.CheckIsValue() {
		t.Error("expected CheckIsValue() to be false for a no-value result")
	}
	notAValue := value.GetNotAValue()
	if notAValue == nil {
		t.Fatal("expected GetNotAValue() to be non-nil")
	}
	if notAValue.Error != wantErr {
		t.Errorf("got error %v, want %v", notAValue.Error, wantErr)
	}
	if notAValue.ValueAddress != "field-address" {
		t.Errorf("got ValueAddress %v, want %v", notAValue.ValueAddress, "field-address")
	}
}

func TestZeroValueGoStructorValueIsNotAValue(t *testing.T) {
	var value GoStructorValue
	if value.CheckIsValue() {
		t.Error("expected the zero-value GoStructorValue to report CheckIsValue() == false")
	}
}
