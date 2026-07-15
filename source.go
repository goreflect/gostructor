package gostructor

import (
	"context"
	"reflect"

	"github.com/goreflect/gostructor/internal/structplan"
)

// Watchable is an optional interface a Source may implement to signal that its
// backing data can change while the program runs. Watch (see watch.go)
// subscribes to every configured source that implements it and re-fills the
// target struct whenever any of them reports a change.
//
// Watch (the method) blocks until ctx is cancelled, calling onChange each time
// the underlying data changes. It must honour ctx — returning ctx.Err() once
// cancelled — and return any fatal error otherwise. onChange may be invoked
// from any goroutine and must be safe to call repeatedly; the Watch driver
// coalesces bursts through WithDebounce, so a source need not debounce itself.
//
// A source that caches a snapshot for Resolve should refresh that snapshot
// *before* calling onChange, so the re-fill triggered by onChange reads the
// new data rather than the old.
type Watchable interface {
	Watch(ctx context.Context, onChange func()) error
}

// Source resolves one struct field's value from a single configuration
// origin: an environment variable, a parsed file, a secret store, and so on.
//
// Resolve returns found=false when nothing in this source applies to the
// field (for example, the field has no key for this source), so the engine
// falls through to the next source instead of treating it as an error.
//
// Name is the source's short identity ("env", "json", "vault", ...), and the
// key used to target it from a cfg per-source override such as
// `cfg:"port,env:SERVER_PORT"`.
//
// Sources outside the core module (gostructor/yaml, gostructor/toml,
// gostructor/hocon, gostructor/vault, or your own) implement this interface
// to plug into Configure via WithSources.
type Source interface {
	// Name is the source's short identity, e.g. "env". It doubles as the
	// per-source override key in a cfg tag.
	Name() string
	// Resolve returns the raw value for field, or found=false if this
	// source has nothing to contribute for it.
	Resolve(field FieldContext) (value any, found bool, err error)
}

// FieldContext describes the struct field being resolved, exposing its parsed
// cfg (base name and per-source overrides) and gos (defaults, flags) metadata
// so a Source never has to parse struct tags itself.
type FieldContext struct {
	reflect.StructField
	// plan is the pre-parsed metadata the engine attaches for cache reuse.
	// It is nil for a FieldContext built by hand (e.g. in a source's own
	// tests); the accessors then parse the embedded StructField's tags on
	// demand.
	plan *structplan.Field
}

// newFieldContext wraps a planned field for resolution, carrying its cached
// parse so the accessors don't re-parse the struct tags on every source turn.
func newFieldContext(field *structplan.Field) FieldContext {
	return FieldContext{StructField: field.Struct, plan: field}
}

// parsed returns the field's metadata, preferring the engine-attached cache
// and falling back to parsing the embedded struct tags for hand-built contexts.
func (f FieldContext) parsed() (base string, overrides, meta map[string]string, flags map[string]bool) {
	if f.plan != nil {
		return f.plan.Base, f.plan.Overrides, f.plan.Meta, f.plan.Flags
	}
	base, overrides = structplan.ParseCfg(f.Tag.Get(structplan.CfgTag))
	meta, flags = structplan.ParseGos(f.Tag.Get(structplan.GosTag))
	return base, overrides, meta, flags
}

// configured reports whether the field carries any cfg/gos configuration. An
// unconfigured leaf field (no base name, override, meta, or flag) is skipped
// silently rather than treated as unresolved.
func (f FieldContext) configured() bool {
	base, overrides, meta, flags := f.parsed()
	return base != "" || len(overrides) > 0 || len(meta) > 0 || len(flags) > 0
}

// TagValue is a convenience for field.Tag.Get(tagName), for the rare source
// that still wants a raw struct tag value.
func (f FieldContext) TagValue(tagName string) string {
	return f.Tag.Get(tagName)
}

// Base returns the field's base name (the first cfg element), from which each
// source derives its key when there is no explicit override.
func (f FieldContext) Base() string {
	base, _, _, _ := f.parsed()
	return base
}

// Override returns the explicit key a cfg tag set for source, as in
// `cfg:"port,env:SERVER_PORT"` -> Override("env") == ("SERVER_PORT", true).
func (f FieldContext) Override(source string) (string, bool) {
	_, overrides, _, _ := f.parsed()
	v, ok := overrides[source]
	return v, ok
}

// SourceKey returns the key source should look up for this field: the explicit
// per-source override if present, otherwise naming(Base()). It returns "" when
// there is no override and no base name, signalling the source does not apply.
func (f FieldContext) SourceKey(source string, naming func(base string) string) string {
	if v, ok := f.Override(source); ok {
		return v
	}
	base := f.Base()
	if base == "" {
		return ""
	}
	if naming == nil {
		return base
	}
	return naming(base)
}

// Meta returns a gos meta value, as in `gos:"default:8080"` ->
// Meta("default") == ("8080", true).
func (f FieldContext) Meta(key string) (string, bool) {
	_, _, meta, _ := f.parsed()
	v, ok := meta[key]
	return v, ok
}

// Flag reports whether a bare gos flag is present, as in `gos:"secret"` ->
// Flag("secret") == true.
func (f FieldContext) Flag(name string) bool {
	_, _, _, flags := f.parsed()
	return flags[name]
}

// IsSecret reports whether the field carries the gos secret flag; its value is
// masked everywhere it would otherwise print (report, trace, ConvertError).
func (f FieldContext) IsSecret() bool { return f.Flag("secret") }

// Optional reports whether the field carries the gos optional flag; an
// optional field left unresolved by every source is not an error.
func (f FieldContext) Optional() bool { return f.Flag("optional") }

// Separator returns the string used to split a flat value into slice elements:
// the gos sep meta if set, otherwise a comma.
func (f FieldContext) Separator() string {
	if s, ok := f.Meta("sep"); ok && s != "" {
		return s
	}
	return ","
}
