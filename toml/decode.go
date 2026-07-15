package toml

// Decode parses TOML into a nested map, the shape gostructor.LookupKey
// addresses. It is a gostructor.Decoder (a named alias of Parse), so a
// remote/file source (gostructor/git, gostructor/watch) can read TOML:
//
//	git.New(git.Options{Repo: ..., Path: "config.toml", Decoder: toml.Decode})
func Decode(raw []byte) (map[string]any, error) {
	return Parse(raw)
}
