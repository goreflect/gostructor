package gostructor

import "log/slog"

// SecretTag marks a field whose value is sensitive: `cf_secret:""`. A tagged
// field's value is masked everywhere it would otherwise print - the
// resolution Report, slog trace records, and *ConvertError messages.
const SecretTag = "cf_secret"

type config struct {
	sources    []Source
	sourcesSet bool
	logger     *slog.Logger
	hooks      []Hook
	masker     Masker
	trace      bool
}

// Masker renders a sensitive field's value for display. It is called for any
// field carrying the cf_secret tag before its value reaches a log, the
// resolution report, or an error message. The default fully redacts.
type Masker func(field FieldContext, value any) string

// defaultMasker fully redacts, revealing nothing about the value.
func defaultMasker(FieldContext, any) string { return "••••••" }

// isSecret reports whether field carries the cf_secret tag (present with any
// value, including empty: `cf_secret:""`).
func (f FieldContext) isSecret() bool {
	_, ok := f.Tag.Lookup(SecretTag)
	return ok
}

// display returns v for a normal field, or the masked rendering for a secret
// one - the value safe to put in a report or log.
func (c *config) display(field FieldContext, v any) any {
	if field.isSecret() {
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

// WithMasker overrides how cf_secret fields are rendered for display. The
// default fully redacts ("••••••"); pass your own to, say, reveal the last
// four characters. The masker is only ever called for fields carrying the
// cf_secret tag.
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

func newConfig(opts []Option) *config {
	c := &config{
		logger: slog.New(slog.DiscardHandler),
		masker: defaultMasker,
	}
	for _, opt := range opts {
		opt(c)
	}
	if !c.sourcesSet {
		c.sources = []Source{Env(), JSON(), INI(), Default()}
	}
	return c
}
