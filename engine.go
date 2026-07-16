package gostructor

import (
	"errors"
	"fmt"
)

// Filler is the interface a generated fast-path fills implement. gostructor-gen
// (Theme 7) emits a `func (c *T) Fill(opts ...Option) error` method on the
// target type; Configure detects it and dispatches to it instead of the
// reflective engine, so callers keep one entry point and get the reflection-free
// path for free.
//
// A Fill must produce byte-for-byte-equivalent results to Configure(&c, opts...)
// on the reflective engine — same field values, same error classification. The
// generator also emits a golden TestFillMatchesReflection to hold that promise.
type Filler interface {
	Fill(opts ...Option) error
}

// Engine selects which resolution path Configure uses. The default,
// EngineAdaptive, transparently uses a generated Fill when the target has one
// and falls back to reflection otherwise, so a project that has not run
// `go generate` keeps working exactly as before.
type Engine int

const (
	// EngineAdaptive dispatches to the generated Fill iff the target implements
	// Filler; otherwise it uses the reflective engine. Same result either way.
	// This is the default and the right choice almost always.
	EngineAdaptive Engine = iota
	// EngineReflection always uses the reflective engine, even if a generated
	// Fill exists. Use it to debug a suspected codegen/reflection divergence or
	// to route around a stale generated file.
	EngineReflection
	// EngineCodegen requires a generated Fill: if the target does not implement
	// Filler, Configure returns ErrNoGeneratedFiller instead of silently
	// reflecting. Use it in perf-critical builds that must guarantee the
	// reflection-free path and want a hard failure if `go generate` was skipped.
	EngineCodegen
)

// ErrNoGeneratedFiller is returned (wrapped, naming the target type) when
// Configure is called with WithEngine(EngineCodegen) but the target has no
// generated Fill method. Match it with errors.Is.
var ErrNoGeneratedFiller = errors.New("gostructor: EngineCodegen requires a generated Fill method, but the target has none (did you run go generate?)")

// WithEngine selects the resolution engine for this Configure call. The default
// is EngineAdaptive; see the Engine constants for when to override it.
func WithEngine(e Engine) Option {
	return func(c *config) {
		c.engine = e
	}
}

// dispatchFiller decides whether a Configure call should hand off to a generated
// Fill, honoring the configured engine. It returns handled=true when it has
// produced the final result (via a generated Fill or a hard EngineCodegen
// error); handled=false means the caller should run the reflective engine.
//
// A report is the reflective engine's view of resolution, so a caller that asked
// for one (withReport) always falls through to reflection even when a Fill
// exists — otherwise the fast path is used whenever it is available.
func dispatchFiller[T any](cfg *config, target *T, withReport bool, opts []Option) (handled bool, err error) {
	if cfg.engine == EngineReflection || target == nil {
		// A nil target is a caller bug; let the reflective engine report it as
		// ErrInvalidTarget rather than panic in a generated Fill's body.
		return false, nil
	}
	filler, ok := any(target).(Filler)
	if ok && !withReport {
		return true, filler.Fill(opts...)
	}
	if !ok && cfg.engine == EngineCodegen {
		return true, fmt.Errorf("%w: target type is %T", ErrNoGeneratedFiller, target)
	}
	return false, nil
}
