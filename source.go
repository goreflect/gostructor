package gostructor

import "reflect"

// Source resolves one struct field's value from a single configuration
// origin: an environment variable, a parsed file, a secret store, and so on.
//
// Resolve returns found=false when nothing in this source applies to field
// (for example, the field doesn't carry this source's struct tag) so the
// engine can fall through to the next source in priority order, rather than
// treating "not tagged for me" as an error.
//
// Implementations living outside the core module (gostructor/yaml,
// gostructor/toml, gostructor/hocon, gostructor/vault, or your own) satisfy
// this interface to plug into Configure via WithSources.
type Source interface {
	// Tag is the struct tag name this source responds to, e.g. "cf_env".
	Tag() string
	// Resolve returns the raw value for field, or found=false if this
	// source has nothing to contribute for it.
	Resolve(field FieldContext) (value any, found bool, err error)
}

// FieldContext describes the struct field currently being resolved.
type FieldContext struct {
	reflect.StructField
}

// TagValue is a convenience for field.Tag.Get(tagName), which is what most
// Source implementations need first.
func (f FieldContext) TagValue(tagName string) string {
	return f.Tag.Get(tagName)
}
