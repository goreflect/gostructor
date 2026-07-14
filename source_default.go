package gostructor

// DefaultTag is the struct tag Default responds to: `cf_default:"8080"`.
// The tag's value is the literal default; for slice/array fields, it's a
// comma-separated list.
const DefaultTag = "cf_default"

type defaultSource struct{}

// Default resolves fields from the literal value on the cf_default tag
// itself. It has no external dependency and always applies if the tag is
// present, so it's typically placed last in a source list as the fallback.
func Default() Source { return defaultSource{} }

func (defaultSource) Tag() string { return DefaultTag }

func (defaultSource) Resolve(field FieldContext) (any, bool, error) {
	raw := field.TagValue(DefaultTag)
	if raw == "" {
		return nil, false, nil
	}
	return splitIfSlice(field, raw), true, nil
}
