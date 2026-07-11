// Package gostructor fills the fields of a Go struct from any mix of
// configuration sources - environment variables, files, secret stores - all
// driven by struct tags. See the README for the full guide; this file holds
// the engine: the generic Configure entry point and the field-resolution
// loop that walks a cached structplan.Plan and asks each configured Source,
// in order, for a value.
package gostructor

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/goreflect/gostructor/convert"
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
func Configure[T any](target *T, opts ...Option) (result *T, err error) {
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
		return nil, errors.New("gostructor: target must not be nil")
	}
	cfg := newConfig(opts)

	structValue := reflect.ValueOf(target).Elem()
	if structValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("gostructor: target must point to a struct, got *%s", structValue.Kind())
	}

	plan := structplan.For(structValue.Type())
	cfg.logger.Debug("configuring struct", "type", structValue.Type().String(), "fields", len(plan.Fields))

	for _, field := range plan.Fields {
		if err := resolveField(cfg, field, structValue); err != nil {
			return nil, err
		}
	}
	return target, nil
}

func resolveField(cfg *config, field structplan.Field, structValue reflect.Value) error {
	fieldCtx := FieldContext{StructField: field.Struct}
	sources := sourcesForField(cfg, fieldCtx)
	anyTagPresent := false

	for _, source := range sources {
		if fieldCtx.TagValue(source.Tag()) == "" {
			continue
		}
		anyTagPresent = true
		raw, found, err := source.Resolve(fieldCtx)
		if err != nil {
			return fmt.Errorf("gostructor: field %q via %s: %w", field.Struct.Name, source.Tag(), err)
		}
		if !found {
			continue
		}
		destination := structValue.FieldByIndex(field.Index)
		converted, err := setValue(destination, raw)
		if err != nil {
			return fmt.Errorf("gostructor: field %q: %w", field.Struct.Name, err)
		}

		// Hooks run on the converted, field-typed value (an actual int,
		// not the string "80" cf_default produced it from) since that's
		// what's actually useful to validate or transform.
		hookValue := converted.Interface()
		for _, hook := range cfg.hooks {
			hookValue, err = hook(fieldCtx, hookValue)
			if err != nil {
				return fmt.Errorf("gostructor: field %q hook: %w", field.Struct.Name, err)
			}
		}
		finalValue := reflect.ValueOf(hookValue)
		if !finalValue.Type().AssignableTo(destination.Type()) {
			return fmt.Errorf("gostructor: field %q: hook returned %s, want %s", field.Struct.Name, finalValue.Type(), destination.Type())
		}
		destination.Set(finalValue)
		cfg.logger.Debug("resolved field", "field", field.Struct.Name, "source", source.Tag())
		return nil
	}

	if anyTagPresent {
		return fmt.Errorf("gostructor: field %q: no configured source produced a value", field.Struct.Name)
	}
	cfg.logger.Debug("skipping untagged field", "field", field.Struct.Name)
	return nil
}

func setValue(destination reflect.Value, raw any) (reflect.Value, error) {
	source := reflect.ValueOf(raw)
	switch destination.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return convert.ToComplex(source, destination)
	default:
		return convert.ToPrimitive(source, destination)
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
