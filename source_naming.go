package gostructor

import (
	"strings"
	"unicode"
)

// Source names for the core module's built-in sources. These strings are what
// a cfg tag targets in a per-source override, e.g. `cfg:"port,env:DB_PORT"`.
const (
	SourceEnv     = "env"
	SourceJSON    = "json"
	SourceINI     = "ini"
	SourceDefault = "default"
)

// Identity returns base unchanged. It is the naming strategy for file sources
// (JSON, YAML, TOML, HOCON, INI), whose keys are the base names as written;
// nested paths are given explicitly via a per-source override.
func Identity(base string) string { return base }

// ScreamingSnake converts a base name to SCREAMING_SNAKE_CASE, the naming
// strategy for environment variables: "port" -> "PORT", "maxConns" ->
// "MAX_CONNS", "HTTPServer" -> "HTTP_SERVER". Hyphens, dots and spaces become
// underscores; camelCase and acronym boundaries get an underscore inserted.
func ScreamingSnake(base string) string {
	var b strings.Builder
	runes := []rune(base)
	for i, r := range runes {
		if r == '-' || r == '.' || r == ' ' {
			b.WriteByte('_')
			continue
		}
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			switch {
			case unicode.IsLower(prev) || unicode.IsDigit(prev):
				// camelCase boundary: maxConns -> MAX_CONNS
				b.WriteByte('_')
			case unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1]):
				// acronym boundary: HTTPServer -> HTTP_SERVER
				b.WriteByte('_')
			}
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}
