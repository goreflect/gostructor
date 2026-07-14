package gostructor

import "os"

// EnvTag is the struct tag Env responds to: `cf_env:"MY_VAR"`.
const EnvTag = "cf_env"

type envSource struct{}

// Env resolves fields from environment variables named by the cf_env tag.
func Env() Source { return envSource{} }

func (envSource) Tag() string { return EnvTag }

func (envSource) Resolve(field FieldContext) (any, bool, error) {
	name := field.TagValue(EnvTag)
	if name == "" {
		return nil, false, nil
	}
	value, set := os.LookupEnv(name)
	if !set || value == "" {
		return nil, false, nil
	}
	return splitIfSlice(field, value), true, nil
}
