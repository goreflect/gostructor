package tools

import (
	"reflect"
	"testing"
)

func TestLookupPathLeafScalar(t *testing.T) {
	data := map[string]interface{}{
		"server": map[string]interface{}{
			"host": "0.0.0.0",
		},
	}
	got, ok := LookupPath(data, "server.host")
	if !ok || got != "0.0.0.0" {
		t.Errorf("got %v, %v", got, ok)
	}
}

func TestLookupPathReturnsWholeMapUnflattened(t *testing.T) {
	data := map[string]interface{}{
		"limits": map[string]interface{}{
			"a": "1",
			"b": "2",
		},
	}
	got, ok := LookupPath(data, "limits")
	if !ok {
		t.Fatal("expected limits to be found")
	}
	want := map[string]interface{}{"a": "1", "b": "2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v (LookupPath must not flatten nested maps)", got, want)
	}
}

func TestLookupPathReturnsList(t *testing.T) {
	data := map[string]interface{}{"tags": []interface{}{"a", "b"}}
	got, ok := LookupPath(data, "tags")
	if !ok {
		t.Fatal("expected tags to be found")
	}
	if !reflect.DeepEqual(got, []interface{}{"a", "b"}) {
		t.Errorf("got %v", got)
	}
}

func TestLookupPathMissing(t *testing.T) {
	data := map[string]interface{}{"a": "1"}
	if _, ok := LookupPath(data, "missing"); ok {
		t.Error("expected ok=false for a missing top-level key")
	}
	if _, ok := LookupPath(data, "a.b"); ok {
		t.Error("expected ok=false when descending into a non-map value")
	}
}
