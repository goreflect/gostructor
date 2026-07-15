// Package gostructor fills the fields of a Go struct from a mix of
// configuration sources (environment variables, files, secret stores),
// driven by two struct tags: cfg for routing/naming and gos for behavior.
// See the README for the full guide. This file holds the engine: the
// Configure entry point and the resolution loop that walks a cached
// structplan.Plan and asks each Source, in slice order, for a value.
//
// Sources are tried in the order passed to WithSources; the first that
// reports found=true wins. There is no per-field priority tag and no global
// selector: the slice order decides priority.
package gostructor

import (
	"fmt"
	"reflect"

	"github.com/goreflect/gostructor/internal/convert"
	"github.com/goreflect/gostructor/internal/structplan"
)

// Configure fills target's fields in place and returns target for chaining.
// With no options it uses a minimal default source list: Env then Default.
// Add file and secret sources (JSON(), INI(), yaml.New(), vault.New(), ...)
// via WithSources; they are opt-in, so a bare cfg base name never triggers
// an unexpected file load.
func Configure[T any](target *T, opts ...Option) (*T, error) {
	result, _, err := configure(target, false, opts)
	return result, err
}

// ConfigureWithReport is Configure plus a resolution trace: for every field
// it records which sources were tried, in what order, which won, and the raw
// and converted values (secrets masked). Report is the machine view; its
// String() renders a human summary. On error the report is returned as far
// as resolution got.
func ConfigureWithReport[T any](target *T, opts ...Option) (*T, *Report, error) {
	return configure(target, true, opts)
}

func configure[T any](target *T, withReport bool, opts []Option) (result *T, report *Report, err error) {
	cfg := newConfig(opts)
	return runConfigure(cfg, target, withReport)
}

// runConfigure fills target using an already-built config. It is the shared
// resolution path behind both Configure (which builds cfg from opts once) and
// Watch (which reuses one cfg across many re-fills), so a live reload takes the
// exact same trace + masking + hook path as the initial fill.
func runConfigure[T any](cfg *config, target *T, withReport bool) (result *T, report *Report, err error) {
	defer func() {
		if r := recover(); r != nil {
			if asErr, ok := r.(error); ok {
				err = asErr
			} else {
				err = fmt.Errorf("gostructor: panic while configuring: %v", r)
			}
		}
	}()

	if target == nil {
		return nil, nil, fmt.Errorf("%w: got nil", ErrInvalidTarget)
	}

	structValue := reflect.ValueOf(target).Elem()
	if structValue.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("%w: got *%s", ErrInvalidTarget, structValue.Kind())
	}

	// Build a report when the caller asked for one or tracing is on;
	// otherwise leave it nil for zero overhead.
	var rep *Report
	if withReport || cfg.trace {
		rep = &Report{Type: structValue.Type().String()}
	}

	plan := structplan.For(structValue.Type())
	cfg.logger.Debug("configuring struct", "type", structValue.Type().String(), "fields", len(plan.Fields))

	for i := range plan.Fields {
		if err := resolveField(cfg, &plan.Fields[i], structValue, rep); err != nil {
			if cfg.trace && rep != nil {
				cfg.logger.Debug("resolution trace (partial)", "trace", rep.String())
			}
			return nil, reportIf(withReport, rep), err
		}
	}
	if cfg.trace && rep != nil {
		cfg.logger.Debug("resolution trace", "trace", rep.String())
	}
	return target, reportIf(withReport, rep), nil
}

// reportIf returns rep only when the caller asked for it, so a trace-only
// report built for logging is not leaked through the return value.
func reportIf(withReport bool, rep *Report) *Report {
	if withReport {
		return rep
	}
	return nil
}

func resolveField(cfg *config, field *structplan.Field, structValue reflect.Value, rep *Report) error {
	fieldCtx := newFieldContext(field)
	sources := cfg.sources

	var fr *FieldResolution
	if rep != nil {
		fr = &FieldResolution{Field: field.Struct.Name, Type: field.Struct.Type.String(), IsSecret: fieldCtx.IsSecret()}
	}
	record := func(outcome string) {
		if fr != nil {
			fr.Outcome = outcome
			rep.Fields = append(rep.Fields, *fr)
		}
	}

	var tried []string
	for i, source := range sources {
		raw, found, err := source.Resolve(fieldCtx)
		if err != nil {
			if fr != nil {
				fr.Attempts = append(fr.Attempts, Attempt{Source: source.Name(), Status: statusError, Detail: err.Error()})
			}
			record(outcomeError)
			return &SourceError{Field: field.Struct.Name, Source: source.Name(), Cause: err}
		}
		tried = append(tried, source.Name())
		if !found {
			if fr != nil {
				fr.Attempts = append(fr.Attempts, Attempt{Source: source.Name(), Status: statusNotFound, Detail: effectiveKey(fieldCtx, source.Name())})
			}
			continue
		}
		destination := structValue.FieldByIndex(field.Index)
		converted, err := convert.Value(reflect.ValueOf(raw), destination.Type())
		if err != nil {
			if fr != nil {
				fr.Attempts = append(fr.Attempts, Attempt{Source: source.Name(), Status: statusError, Detail: effectiveKey(fieldCtx, source.Name())})
			}
			record(outcomeError)
			// Mask the value for secret fields so the error is safe to
			// log. The cause (e.g. *strconv.NumError) embeds the raw value
			// in its own message, so for a secret field we drop the cause
			// rather than leak it through Unwrap.
			if fieldCtx.IsSecret() {
				return &ConvertError{Field: field.Struct.Name, Value: cfg.masker(fieldCtx, raw), Target: destination.Type(), Cause: nil}
			}
			return &ConvertError{Field: field.Struct.Name, Value: raw, Target: destination.Type(), Cause: err}
		}

		// Hooks run on the converted, field-typed value (a real int, not the
		// string "80" it was parsed from), which is what you validate.
		hookValue := converted.Interface()
		for _, hook := range cfg.hooks {
			hookValue, err = hook(fieldCtx, hookValue)
			if err != nil {
				record(outcomeError)
				return &HookError{Field: field.Struct.Name, Cause: err}
			}
		}
		finalValue := reflect.ValueOf(hookValue)
		if !finalValue.Type().AssignableTo(destination.Type()) {
			record(outcomeError)
			return &HookError{Field: field.Struct.Name, Cause: fmt.Errorf("hook returned %s, want %s", finalValue.Type(), destination.Type())}
		}
		destination.Set(finalValue)
		if fr != nil {
			fr.Attempts = append(fr.Attempts, Attempt{Source: source.Name(), Status: statusUsed, Detail: effectiveKey(fieldCtx, source.Name())})
			markSkipped(fr, sources[i+1:])
			fr.Winner = source.Name()
			fr.Raw = cfg.display(fieldCtx, raw)
			fr.Value = cfg.display(fieldCtx, hookValue)
			if source.Name() == SourceDefault {
				record(outcomeDefault)
			} else {
				record(outcomeResolved)
			}
		}
		cfg.logger.Debug("resolved field", "field", field.Struct.Name, "source", source.Name())
		return nil
	}

	if fieldCtx.configured() {
		if fieldCtx.Optional() {
			cfg.logger.Debug("optional field left unresolved", "field", field.Struct.Name)
			record(outcomeUnresolved)
			return nil
		}
		record(outcomeUnresolved)
		return &NotResolvedError{Field: field.Struct.Name, Sources: tried}
	}
	cfg.logger.Debug("skipping unconfigured field", "field", field.Struct.Name)
	record(outcomeSkipped)
	return nil
}

// effectiveKey is a key hint for the trace: the field's per-source override
// if set, else its base name. It is informational only (a source may
// transform the base further, e.g. env upper-cases it), never used for lookup.
func effectiveKey(field FieldContext, source string) string {
	if v, ok := field.Override(source); ok {
		return v
	}
	return field.Base()
}

// markSkipped records the sources that never got a turn because an earlier
// source already won, so the report shows the full source order.
func markSkipped(fr *FieldResolution, remaining []Source) {
	for _, s := range remaining {
		fr.Attempts = append(fr.Attempts, Attempt{Source: s.Name(), Status: statusSkipped})
	}
}
