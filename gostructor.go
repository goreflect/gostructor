// Package gostructor fills the fields of a Go struct from any mix of
// configuration sources - environment variables, files, secret stores - all
// driven by struct tags. See the README for the full guide; this file holds
// the engine: the generic Configure entry point and the field-resolution
// loop that walks a cached structplan.Plan and asks each configured Source,
// in order, for a value.
package gostructor

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/goreflect/gostructor/internal/convert"
	"github.com/goreflect/gostructor/internal/priority"
	"github.com/goreflect/gostructor/internal/structplan"
)

// PriorityEnvVar is the environment variable read to pick which cf_priority
// stage applies, e.g. GOSTRUCTOR_PRIORITY=prod.
const PriorityEnvVar = "GOSTRUCTOR_PRIORITY"

// PriorityTag is the struct tag used to declare a per-field source order,
// e.g. `cf_priority:"prod:cf_env,cf_default;dev:cf_default,cf_env"`.
const PriorityTag = "cf_priority"

// Configure fills target's fields in place from the configured sources and
// returns target back for convenient chaining. With no options, it uses the
// core module's built-in sources: Env, JSON, then Default. Bring in
// external sources (yaml.New(), vault.New(), ...) via WithSources.
func Configure[T any](target *T, opts ...Option) (*T, error) {
	result, _, err := configure(target, false, opts)
	return result, err
}

// ConfigureWithReport is Configure plus a structured resolution trace: for
// every field it records which sources were tried, in what order, which one
// won, and the raw and converted values (secrets masked). The Report is the
// machine view; its String() renders a human-readable per-field tree. On
// error the report is returned as far as resolution got, alongside the error.
func ConfigureWithReport[T any](target *T, opts ...Option) (*T, *Report, error) {
	return configure(target, true, opts)
}

func configure[T any](target *T, withReport bool, opts []Option) (result *T, report *Report, err error) {
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
	cfg := newConfig(opts)

	structValue := reflect.ValueOf(target).Elem()
	if structValue.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("%w: got *%s", ErrInvalidTarget, structValue.Kind())
	}

	// A report is built when explicitly requested or when tracing to a
	// logger is on; otherwise it stays nil for zero overhead.
	var rep *Report
	if withReport || cfg.trace {
		rep = &Report{}
	}

	plan := structplan.For(structValue.Type())
	cfg.logger.Debug("configuring struct", "type", structValue.Type().String(), "fields", len(plan.Fields))

	for _, field := range plan.Fields {
		if err := resolveField(cfg, field, structValue, rep); err != nil {
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

// reportIf returns rep only to callers that asked for it; a trace-only report
// (built for logging) is not surfaced through the return value.
func reportIf(withReport bool, rep *Report) *Report {
	if withReport {
		return rep
	}
	return nil
}

func resolveField(cfg *config, field structplan.Field, structValue reflect.Value, rep *Report) error {
	fieldCtx := FieldContext{StructField: field.Struct}
	sources := sourcesForField(cfg, fieldCtx)

	var fr *FieldResolution
	if rep != nil {
		fr = &FieldResolution{Field: field.Struct.Name, Type: field.Struct.Type.String()}
	}
	record := func(outcome string) {
		if fr != nil {
			fr.Outcome = outcome
			rep.Fields = append(rep.Fields, *fr)
		}
	}

	var presentTags []string
	for i, source := range sources {
		key := fieldCtx.TagValue(source.Tag())
		if key == "" {
			continue
		}
		presentTags = append(presentTags, source.Tag())
		raw, found, err := source.Resolve(fieldCtx)
		if err != nil {
			if fr != nil {
				fr.Attempts = append(fr.Attempts, Attempt{Tag: source.Tag(), Status: statusError, Detail: err.Error()})
			}
			record(outcomeError)
			return &SourceError{Field: field.Struct.Name, Tag: source.Tag(), Cause: err}
		}
		if !found {
			if fr != nil {
				fr.Attempts = append(fr.Attempts, Attempt{Tag: source.Tag(), Status: statusNotFound, Detail: key})
			}
			continue
		}
		destination := structValue.FieldByIndex(field.Index)
		converted, err := convert.Value(reflect.ValueOf(raw), destination.Type())
		if err != nil {
			if fr != nil {
				fr.Attempts = append(fr.Attempts, Attempt{Tag: source.Tag(), Status: statusError, Detail: key})
			}
			record(outcomeError)
			// Mask the offending value for secret fields so the error is
			// safe to log or surface to a user. The underlying cause
			// (e.g. *strconv.NumError) embeds the raw value in its own
			// message, so for a secret field we drop it entirely rather
			// than leak it through Unwrap - security wins over the errors.As
			// chain here.
			if fieldCtx.isSecret() {
				return &ConvertError{Field: field.Struct.Name, Value: cfg.masker(fieldCtx, raw), Target: destination.Type(), Cause: nil}
			}
			return &ConvertError{Field: field.Struct.Name, Value: raw, Target: destination.Type(), Cause: err}
		}

		// Hooks run on the converted, field-typed value (an actual int,
		// not the string "80" cf_default produced it from) since that's
		// what's actually useful to validate or transform.
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
			fr.Attempts = append(fr.Attempts, Attempt{Tag: source.Tag(), Status: statusUsed, Detail: key})
			markSkipped(fr, fieldCtx, sources[i+1:])
			fr.Winner = source.Tag()
			fr.Raw = cfg.display(fieldCtx, raw)
			fr.Value = cfg.display(fieldCtx, hookValue)
			if source.Tag() == DefaultTag {
				record(outcomeDefault)
			} else {
				record(outcomeResolved)
			}
		}
		cfg.logger.Debug("resolved field", "field", field.Struct.Name, "source", source.Tag())
		return nil
	}

	if len(presentTags) > 0 {
		record(outcomeUnresolved)
		return &NotResolvedError{Field: field.Struct.Name, Tags: presentTags}
	}
	cfg.logger.Debug("skipping untagged field", "field", field.Struct.Name)
	record(outcomeSkipped)
	return nil
}

// markSkipped records the still-tagged sources that never got a turn because
// an earlier source already won, so the report shows the full source order.
func markSkipped(fr *FieldResolution, fieldCtx FieldContext, remaining []Source) {
	for _, s := range remaining {
		if key := fieldCtx.TagValue(s.Tag()); key != "" {
			fr.Attempts = append(fr.Attempts, Attempt{Tag: s.Tag(), Status: statusSkipped, Detail: key})
		}
	}
}

// sourcesForField returns the source order to use for one field: the
// cf_priority-selected order if the field carries that tag and it resolves
// against the current PriorityEnvVar selection, otherwise cfg.sources as-is.
func sourcesForField(cfg *config, field FieldContext) []Source {
	tag := field.TagValue(PriorityTag)
	if tag == "" {
		return cfg.sources
	}
	ast, err := priority.NewParser(strings.NewReader(tag)).Parse()
	if err != nil {
		cfg.logger.Error("could not parse cf_priority tag", "field", field.Name, "error", err)
		return cfg.sources
	}
	selected := priority.GetPriorityChains(ast, os.Getenv(PriorityEnvVar))
	if len(selected) == 0 {
		return cfg.sources
	}
	ordered := make([]Source, 0, len(selected))
	for _, tagName := range selected {
		for _, source := range cfg.sources {
			if source.Tag() == tagName {
				ordered = append(ordered, source)
				break
			}
		}
	}
	if len(ordered) == 0 {
		return cfg.sources
	}
	return ordered
}
