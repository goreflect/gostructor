package gostructor

import (
	"errors"
	"testing"
)

// fakeFiller stands in for a gostructor-gen-generated type: its Fill writes a
// value the reflective engine never would, so a test can tell which path ran.
type fakeFiller struct {
	Value  string `cfg:"value" gos:"default:reflected"`
	filled bool
}

func (f *fakeFiller) Fill(opts ...Option) error {
	f.filled = true
	f.Value = "generated"
	return nil
}

// noFiller has no Fill method.
type noFiller struct {
	Value string `cfg:"value" gos:"default:reflected"`
}

func TestConfigure_AdaptiveDispatchesToFill(t *testing.T) {
	var f fakeFiller
	if _, err := Configure(&f); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	if !f.filled || f.Value != "generated" {
		t.Fatalf("adaptive engine did not dispatch to Fill: %+v", f)
	}
}

func TestConfigure_ReflectionIgnoresFill(t *testing.T) {
	var f fakeFiller
	if _, err := Configure(&f, WithEngine(EngineReflection)); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	if f.filled {
		t.Fatalf("EngineReflection called Fill, should have reflected")
	}
	if f.Value != "reflected" {
		t.Fatalf("EngineReflection did not use the reflective path: %+v", f)
	}
}

func TestConfigure_CodegenRequiresFill(t *testing.T) {
	var n noFiller
	_, err := Configure(&n, WithEngine(EngineCodegen))
	if !errors.Is(err, ErrNoGeneratedFiller) {
		t.Fatalf("EngineCodegen without a Fill: got %v, want ErrNoGeneratedFiller", err)
	}

	var f fakeFiller
	if _, err := Configure(&f, WithEngine(EngineCodegen)); err != nil {
		t.Fatalf("EngineCodegen with a Fill: %v", err)
	}
	if !f.filled {
		t.Fatalf("EngineCodegen did not dispatch to Fill: %+v", f)
	}
}

func TestConfigureWithReport_UsesReflection(t *testing.T) {
	// The report is the reflective engine's view, so ConfigureWithReport must
	// reflect even when a generated Fill exists.
	var f fakeFiller
	_, rep, err := ConfigureWithReport(&f)
	if err != nil {
		t.Fatalf("ConfigureWithReport: %v", err)
	}
	if f.filled {
		t.Fatalf("ConfigureWithReport dispatched to Fill; it should reflect for the report")
	}
	if rep == nil {
		t.Fatalf("ConfigureWithReport returned no report")
	}
	if f.Value != "reflected" {
		t.Fatalf("ConfigureWithReport did not reflect: %+v", f)
	}
}

func TestConfigure_NilTargetStillReports(t *testing.T) {
	// A nil target must reach the reflective engine's ErrInvalidTarget rather
	// than panic in a generated Fill.
	_, err := Configure[fakeFiller](nil, WithEngine(EngineCodegen))
	if !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("nil target: got %v, want ErrInvalidTarget", err)
	}
}
