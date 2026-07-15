package gostructor

type defaultSource struct{}

// Default resolves fields from the literal value on the gos default meta,
// `gos:"default:8080"`. It has no external dependency and always applies when
// the meta is present, so it's typically placed last in a source list as the
// fallback. For a slice/array field the value is split on the field's separator
// (gos sep, default comma).
func Default() Source { return defaultSource{} }

func (defaultSource) Name() string { return SourceDefault }

func (defaultSource) Resolve(field FieldContext) (any, bool, error) {
	raw, ok := field.Meta("default")
	if !ok {
		return nil, false, nil
	}
	return splitIfSlice(field, raw), true, nil
}
