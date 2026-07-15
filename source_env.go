package gostructor

import "os"

type envSource struct{}

// Env resolves fields from environment variables. The variable name is the
// field's base name in SCREAMING_SNAKE_CASE (`cfg:"port"` -> PORT), unless the
// cfg tag overrides it (`cfg:"port,env:DB_PORT_LEGACY"` -> DB_PORT_LEGACY).
func Env() Source { return envSource{} }

func (envSource) Name() string { return SourceEnv }

func (envSource) Resolve(field FieldContext) (any, bool, error) {
	name := field.SourceKey(SourceEnv, ScreamingSnake)
	if name == "" {
		return nil, false, nil
	}
	value, set := os.LookupEnv(name)
	if !set || value == "" {
		return nil, false, nil
	}
	return splitIfSlice(field, value), true, nil
}
