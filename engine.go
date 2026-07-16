package gostructor

import (
	"errors"
	"fmt"
)

// Filler is the interface a generated Fill implements. gostructor-gen emits a
// func (c *T) Fill(opts ...Option) error method on the target type; Configure
// detects it and calls it instead of running the reflective engine, so the call
// site stays the same.
//
// A Fill must return the same result as Configure(&c, opts...) on the reflective
// engine: same field values, same error classification. The generator emits a
// TestFillMatchesReflection that checks this.
type Filler interface {
	Fill(opts ...Option) error
}

// Engine selects which resolution path Configure uses. The default,
// EngineAdaptive, calls a generated Fill when the target has one and reflects
// otherwise, so a package that hasn't run go generate still works.
type Engine int

const (
	// EngineAdaptive calls the generated Fill if the target implements Filler,
	// otherwise it reflects. Same result either way. This is the default.
	EngineAdaptive Engine = iota
	// EngineReflection always reflects, even if a generated Fill exists. Use it
	// to debug a suspected divergence, or to work around a stale generated file.
	EngineReflection
	// EngineCodegen requires a generated Fill: if the target doesn't implement
	// Filler, Configure returns ErrNoGeneratedFiller instead of reflecting. Use
	// it in builds that must guarantee the reflection-free path and want to fail
	// if go generate was skipped.
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
// produced the final result (a generated Fill ran, or EngineCodegen errored);
// handled=false means the caller should run the reflective engine.
//
// A report is the reflective engine's view of resolution, so a caller that asked
// for one (withReport) always reflects, even when a Fill exists. A configured
// debug dumper needs that same report, so it reflects too. Otherwise the Fill is
// used whenever it's available.
func dispatchFiller[T any](cfg *config, target *T, withReport bool, opts []Option) (handled bool, err error) {
	if cfg.engine == EngineReflection || target == nil {
		// A nil target is a caller bug; let the reflective engine report it as
		// ErrInvalidTarget rather than panic in a generated Fill's body.
		return false, nil
	}
	needsReport := withReport || cfg.dumper != nil
	filler, ok := any(target).(Filler)
	if ok && !needsReport {
		return true, filler.Fill(opts...)
	}
	if !ok && cfg.engine == EngineCodegen {
		return true, fmt.Errorf("%w: target type is %T", ErrNoGeneratedFiller, target)
	}
	return false, nil
}
