package gostructor

import (
	"fmt"
	"reflect"

	"github.com/goreflect/gostructor/internal/structplan"
)

// This file is runtime support for the Fill methods gostructor-gen (Theme 7)
// emits. A generated Fill drives resolution through a GenRuntime instead of the
// reflective engine in gostructor.go: it builds one runtime from the caller's
// options, then, per field, asks the runtime for the winning source value and
// converts it into the concrete field with the typed helpers in gostructor/gen.
// The runtime keeps the parts that must stay source-agnostic and identical to
// the reflective path — the ordered source loop, hook application, secret
// masking, and the not-resolved decision — so a generated Fill produces the same
// values and the same error taxonomy as Configure, just without the per-fill
// reflection.
//
// Everything here is exported only so generated code (which lives in the user's
// module) can call it; it is not meant to be used by hand.

// GenRuntime resolves fields for a generated Fill. Build one per Fill call with
// NewGenRuntime; it carries the same configured sources, hooks, and masker a
// reflective Configure would use for the same options.
type GenRuntime struct {
	cfg *config
}

// NewGenRuntime builds the resolution runtime for a generated Fill from the
// caller's options — the same option set Configure would receive.
func NewGenRuntime(opts ...Option) *GenRuntime {
	return &GenRuntime{cfg: newConfig(opts)}
}

// Fields returns the flattened, resolvable fields of struct type t as
// FieldContexts (using the cached structplan for t), for a generated file to
// build its per-field contexts once. Order matches declaration order.
func Fields(t reflect.Type) []FieldContext {
	plan := structplan.For(t)
	out := make([]FieldContext, len(plan.Fields))
	for i := range plan.Fields {
		out[i] = newFieldContext(&plan.Fields[i])
	}
	return out
}

// FieldsByName is Fields keyed by struct field name, for a generated file to
// look a field's context up by name. Generated code caches this once at package
// init, so the per-fill cost is a map read, not a reflect walk.
func FieldsByName(t reflect.Type) map[string]FieldContext {
	fields := Fields(t)
	out := make(map[string]FieldContext, len(fields))
	for _, fc := range fields {
		out[fc.Name] = fc
	}
	return out
}

// ResolveField runs the ordered source loop for one field and returns the raw
// value from the first source that has one (found=false if none do). A source
// error is wrapped as a *SourceError, exactly as the reflective engine does; the
// generated code converts the raw value and applies hooks itself.
func (r *GenRuntime) ResolveField(fc FieldContext) (raw any, found bool, err error) {
	for _, source := range r.cfg.sources {
		value, ok, e := source.Resolve(fc)
		if e != nil {
			return nil, false, &SourceError{Field: fc.Name, Source: source.Name(), Cause: e}
		}
		if ok {
			return value, true, nil
		}
	}
	return nil, false, nil
}

// NotResolved returns the error for a field that no source produced a value for,
// matching the reflective engine's terminal handling: a configured, non-optional
// field yields a *NotResolvedError naming every source tried; an optional or
// unconfigured field yields nil (it is simply left at its zero value).
func (r *GenRuntime) NotResolved(fc FieldContext) error {
	if !fc.configured() || fc.Optional() {
		return nil
	}
	names := make([]string, len(r.cfg.sources))
	for i, s := range r.cfg.sources {
		names[i] = s.Name()
	}
	return &NotResolvedError{Field: fc.Name, Sources: names}
}

// ApplyHooks runs the configured hooks over an already-converted, field-typed
// value in order, exactly as the reflective engine does, wrapping a hook failure
// as a *HookError. The returned value is what the generated code assigns to the
// field (after checking its type with HookTypeError).
func (r *GenRuntime) ApplyHooks(fc FieldContext, value any) (any, error) {
	var err error
	for _, hook := range r.cfg.hooks {
		value, err = hook(fc, value)
		if err != nil {
			return nil, &HookError{Field: fc.Name, Cause: err}
		}
	}
	return value, nil
}

// HookTypeError builds the *HookError the reflective engine returns when a hook
// hands back a value that is not assignable to the field's type. Generated code
// calls it when the type assertion after ApplyHooks fails.
func (r *GenRuntime) HookTypeError(fc FieldContext, value any, target reflect.Type) error {
	return &HookError{Field: fc.Name, Cause: fmt.Errorf("hook returned %s, want %s", reflect.TypeOf(value), target)}
}

// ConvertError builds the *ConvertError for a value that could not be converted
// into the field's type, reproducing the reflective engine's secret handling:
// for a secret field the raw value is masked and the underlying cause dropped so
// nothing sensitive leaks through the error or its Unwrap.
func (r *GenRuntime) ConvertError(fc FieldContext, raw any, target reflect.Type, cause error) error {
	if fc.IsSecret() {
		return &ConvertError{Field: fc.Name, Value: r.cfg.masker(fc, raw), Target: target, Cause: nil}
	}
	return &ConvertError{Field: fc.Name, Value: raw, Target: target, Cause: cause}
}
