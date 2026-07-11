package gostructor

import "log/slog"

type config struct {
	sources    []Source
	sourcesSet bool
	logger     *slog.Logger
	hooks      []Hook
}

// Hook runs after a value has been resolved for a field but before it is
// set on the struct, letting you validate or transform it. Returning an
// error aborts Configure with that error.
type Hook func(field FieldContext, value any) (any, error)

// Option configures a single Configure call.
type Option func(*config)

// WithSources replaces the default source list with an explicit, ordered
// one. Sources are tried in the order given; the first one that reports
// found=true for a field wins (unless overridden per-field by a cf_priority
// tag). Use this to bring in external sources such as yaml.New() or
// vault.New(), since the core module only ships Env, Default, and JSON.
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

func newConfig(opts []Option) *config {
	c := &config{
		logger: slog.New(slog.DiscardHandler),
	}
	for _, opt := range opts {
		opt(c)
	}
	if !c.sourcesSet {
		c.sources = []Source{Env(), JSON(), INI(), Default()}
	}
	return c
}
