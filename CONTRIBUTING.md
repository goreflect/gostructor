# Contributing

Issues and pull requests are welcome. If you're picking up one of the
[roadmap](README.md#roadmap) items or fixing a bug, a short description of
the approach in the PR body is appreciated so reviewers don't have to
reverse-engineer intent from the diff.

## Layout

- Root module (`.`): the core engine (`Configure`, `Source`, the built-in
  Env/Default/JSON/INI sources) and `internal/` (implementation details not
  part of the public API — type conversion, tag-priority parsing, struct
  planning).
- `yaml/`, `toml/`, `hocon/`, `vault/`: optional sources, each its own Go
  module with a `replace` directive back to the root module for local
  development.
- `examples/`: runnable programs demonstrating usage. `examples/basic` is
  part of the root module; `examples/multisource` has its own `go.mod`
  since it depends on `yaml/`.
- `testdata/`: fixture files used by tests across modules.

## Running checks

```sh
make build   # go build ./... in every module
make vet
make test    # go test ./... -race -cover in every module
make all     # build + vet + test
```

A new module (source implementation, or an `examples/` program with
external dependencies) needs to be added to `MODULES` in the `Makefile` and
to the `matrix` in `.github/workflows/ci.yaml`.

## Adding a Source

Implement the two-method `Source` interface (`Tag() string` and
`Resolve(field FieldContext) (value any, found bool, err error)`) in a new
module under the repo root, following `yaml/source.go` as a template.
`found=false` means "this source doesn't apply to this field," not an
error — return it whenever the field doesn't carry your tag so `Configure`
falls through to the next source.

## Known limitations before you file a bug

`gostructor/hocon` and `gostructor/toml` intentionally parse a subset of
each spec — check [Known limitations](README.md#known-limitations) before
assuming a parse failure is a bug.
