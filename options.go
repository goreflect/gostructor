package gostructor

import (
	"fmt"
	"log/slog"
	"time"
)

// A field marked with the gos secret flag (`gos:"secret"`) is masked
// everywhere its value would otherwise print: the Report, slog trace records,
// and *ConvertError messages. See FieldContext.IsSecret.

type config struct {
	sources    []Source
	sourcesSet bool
	logger     *slog.Logger
	hooks      []Hook
	masker     Masker
	trace      bool
	// debounce coalesces a burst of change signals during Watch into a single
	// reload; zero disables debouncing. See WithDebounce.
	debounce time.Duration
	// validate gates a live reload against the freshly filled struct before it
	// is published; nil disables it. See WithValidate. It is stored type-erased
	// (the *T assertion lives in the closure WithValidate builds).
	validate func(any) error
}

// Masker renders a sensitive field's value for display. It is called for any
// field carrying the gos secret flag before its value reaches a log, the
// resolution report, or an error message. The default fully redacts.
type Masker func(field FieldContext, value any) string

// defaultMasker fully redacts, revealing nothing about the value.
func defaultMasker(FieldContext, any) string { return "••••••" }

// display returns v for a normal field, or the masked rendering for a secret
// one: the value safe to put in a report or log.
func (c *config) display(field FieldContext, v any) any {
	if field.IsSecret() {
		return c.masker(field, v)
	}
	return v
}

// Hook runs after a value has been resolved for a field but before it is
// set on the struct, letting you validate or transform it. Returning an
// error aborts Configure with that error.
type Hook func(field FieldContext, value any) (any, error)

// Option configures a single Configure call.
type Option func(*config)

// WithSources replaces the default source list with an explicit, ordered
// one. Sources are tried in the order given; the first one that reports
// found=true for a field wins. This slice order is how priority is expressed:
// put the source that should win first. Use this to bring in external sources
// such as yaml.New() or vault.New(), since the core module only ships Env,
// JSON, INI, and Default.
func WithSources(sources ...Source) Option {
	return func(c *config) {
		c.sources = sources
		c.sourcesSet = true
	}
}

// WithLogger sets the logger Configure uses for diagnostic output. The
// default is a no-op logger; pass slog.Default() (or your own) to see
// what gostructor is doing.
func WithLogger(l *slog.Logger) Option {
	return func(c *config) {
		c.logger = l
	}
}

// WithHook registers a hook run after each field is resolved, in the order
// added. Use it for validation (return an error to reject a value) or
// transformation (return a different value).
func WithHook(h Hook) Option {
	return func(c *config) {
		c.hooks = append(c.hooks, h)
	}
}

// WithMasker overrides how secret fields (gos:"secret") are rendered for
// display. The default fully redacts ("••••••"); pass your own to, say, reveal
// the last four characters. The masker is only ever called for fields carrying
// the gos secret flag.
func WithMasker(m Masker) Option {
	return func(c *config) {
		if m != nil {
			c.masker = m
		}
	}
}

// WithTrace makes Configure log the full resolution report at debug level
// through the configured logger (see WithLogger) once resolution completes.
// It has no effect without a logger. To capture the report programmatically
// instead, use ConfigureWithReport.
func WithTrace() Option {
	return func(c *config) {
		c.trace = true
	}
}

// WithDebounce sets how long Watch waits for a source's change signals to go
// quiet before running one reload, collapsing a burst (an editor's several
// writes for one save, a config server's batch of key events) into a single
// re-fill. Zero disables debouncing (every signal reloads immediately). It has
// no effect on a plain Configure call.
func WithDebounce(d time.Duration) Option {
	return func(c *config) {
		c.debounce = d
	}
}

// WithValidate registers a whole-struct validation run during Watch after a
// reload has filled a fresh copy but before it is published. Returning an error
// rejects that reload: the previously good config keeps serving and the error
// is delivered to onReload rather than swapping in a broken value. It
// complements per-field WithHook (which runs during every fill) with a check
// that can see the whole struct at once. It has no effect on a plain Configure
// call; validate there is the caller's own concern.
func WithValidate[T any](fn func(*T) error) Option {
	return func(c *config) {
		if fn == nil {
			return
		}
		c.validate = func(v any) error {
			target, ok := v.(*T)
			if !ok {
				// Watch always passes the *T it just filled, so a mismatch here
				// means WithValidate's T differs from the Watch target's T.
				return fmt.Errorf("gostructor: WithValidate type %T does not match target %T", (*T)(nil), v)
			}
			return fn(target)
		}
	}
}

func newConfig(opts []Option) *config {
	c := &config{
		logger: slog.New(slog.DiscardHandler),
		masker: defaultMasker,
	}
	for _, opt := range opts {
		opt(c)
	}
	if !c.sourcesSet {
		// Minimal, dependency-free default: env vars then gos defaults.
		// File and secret sources are opt-in via WithSources, so a bare cfg
		// base name never triggers an unexpected file load.
		c.sources = []Source{Env(), Default()}
	}
	return c
}
