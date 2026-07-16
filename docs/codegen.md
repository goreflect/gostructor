# Code generation (`gostructor-gen`)

The reflective engine walks `reflect.Value`s on every `Configure` call. That's
convenient but it boxes scalars into `any` and allocates on the hot path. If you
reload config often or fill it during a latency-sensitive startup, that adds up.

`gostructor-gen` parses your struct at `go generate` time and writes a typed
`Fill` method that resolves each field straight into its Go type: no per-fill
reflection, no boxing, near-zero allocation. It uses the same sources, the same
priority order, and the same conversion and error rules as the reflective
engine.

The reflective path stays the default. The generated one is opt-in, and a
generated test keeps the two in sync (see below).

## Quick start

There are two ways to tell `gostructor-gen` which structs to generate for.

**By name** — a `//go:generate` directive naming the type:

```go
//go:generate go run github.com/goreflect/gostructor/cmd/gostructor-gen -type Config

type Config struct {
    Host    string        `cfg:"host" gos:"default:0.0.0.0"`
    Port    int           `cfg:"port" gos:"default:8080"`
    DBURL   string        `cfg:"dbURL,env:DB_URL"`
    Timeout time.Duration `cfg:"timeout" gos:"default:30s"`
}
```

**By marker** — annotate the struct with `//gostructor:gen` and let the tool
find it. Handy when you have several config structs across several packages:
mark the ones you want and point the tool at the tree.

```go
//gostructor:gen
type Config struct { /* ... */ }
```

```sh
go generate ./...                    # marker discovery in each package
# or, once, for the whole project:
go run github.com/goreflect/gostructor/cmd/gostructor-gen -dir . -recursive
```

(To use marker discovery from a `//go:generate` line, drop the `-type`:
`//go:generate go run .../cmd/gostructor-gen -dir .`.)

Either way you get two files next to the source, named after the struct
(lower-cased) with a `.gs.go` suffix:

- `config.gs.go` — the `func (c *Config) Fill(opts ...gostructor.Option) error`,
  with each field's parsing in its own small helper rather than one long body.
- `config.gs_test.go` — `TestFillMatchesReflection_Config`, which fills the
  struct both ways and fails if the results or errors differ.

Your call site doesn't change:

```go
cfg, err := gostructor.Configure(&Config{}, gostructor.WithSources(env.New()))
```

`Configure` detects the generated `Fill` (any target implementing
`interface{ Fill(...Option) error }`) and calls it. That's the default engine,
`EngineAdaptive`: fast where generated, correct everywhere. There's a runnable
program in [`examples/codegen`](../examples/codegen).

## Engine modes

Codegen is never required — a package that hasn't run `go generate` works as
before. `WithEngine` picks how `Configure` resolves:

| Mode | Behaviour | Use it when |
|---|---|---|
| `EngineAdaptive` *(default)* | Call the generated `Fill` if the target has one, otherwise reflect. Same result either way. | Almost always. |
| `EngineReflection` | Always reflect, even if a `Fill` exists. | Debugging a suspected divergence, or working around a stale generated file. |
| `EngineCodegen` | Require a generated `Fill`; if there's none, return `ErrNoGeneratedFiller` instead of reflecting. | Builds that must guarantee the reflection-free path and want to fail if someone forgot `go generate`. |

```go
// Require the fast path; fail loudly if `go generate` wasn't run:
cfg, err := gostructor.Configure(&Config{},
    gostructor.WithEngine(gostructor.EngineCodegen),
    gostructor.WithSources(env.New()))
if errors.Is(err, gostructor.ErrNoGeneratedFiller) {
    log.Fatal("run `go generate ./...` — this build requires the codegen fast path")
}
```

`ConfigureWithReport` always uses the reflective engine, since the report is
that engine's view. It's populated even for a type that has a generated `Fill`.

## How the two paths stay equivalent

The generated `Fill` doesn't hardcode which sources exist. It iterates the same
`[]Source` you pass to `WithSources`, in the same order, calling each one's
`Resolve`. What it drops is the per-fill reflect walk (the fields are known at
generate time) and the reflective conversion: each resolved value goes through a
direct typed call into [`gostructor/gen`](../gen) (`gen.Int`, `gen.Duration`,
`gen.StringSlice`, …) and is assigned straight into the field.

Those typed conversions are the reflection-free counterparts of the reflective
`convert.Value` and share its overflow, integral-float, and base-10 rules, so a
value converts (or fails, with the same `*ConvertError`) the same way on both
paths. Hooks, secret masking, the not-resolved decision, and the
`SourceError`/`NotResolvedError`/`HookError` types all run through the shared
runtime the reflective engine uses.

The generated `TestFillMatchesReflection_<Type>` is the check: it fills the same
struct through both engines and asserts identical values and identical error
classification. Commit the generated files and run `go generate ./...` in CI (or
diff-check them) so a struct change without a regenerate gets caught.

## Supported field types

Generation is total: if a struct has a field the generator doesn't support, it
fails at `go generate` time with a clear message rather than emitting a partial
`Fill`. You get either a fully correct fast path or an error telling you to keep
that type on the reflective engine.

There are two conversion tiers, both equivalent to the reflective engine:

**Reflection-free** — a dedicated typed `gen.*` call:

- string, bool, and the sized numerics (`int`/`int8`…`int64`,
  `uint`/`uint8`…`uint64`, `float32`, `float64`),
- `time.Duration`,
- `[]string`.

**Via `gen.Reflective[T]`** — routed through the same reflective core, so the
reflect cost is paid only for that one field:

- other slices and arrays (`[]int`, `[3]bool`),
- maps (`map[string]int`),
- in-package named types over any of the above.

**Nested structs** are flattened as the reflective engine flattens them. An
untagged nested struct is recursed into and its leaves resolved under their own
keys (`c.Service.Name`). A nested struct carrying a `cfg`/`gos` tag is atomic and
filled via `gen.Reflective` from an object source.

Not yet supported (these fail generation): embedded fields, pointers,
out-of-package named types other than `time.Duration`, and
`encoding.TextUnmarshaler` types like `time.Time`. Keep structs containing them
on the reflective engine — under `EngineAdaptive` they just reflect, no code
change needed.

See [`examples/codegen`](../examples/codegen) for a config that exercises nested
structs, a map, and an `[]int`.

## Flags

```
gostructor-gen [-type Name[,Name2,...]] [-dir .] [-recursive]
```

- `-type` — comma-separated struct names to generate a `Fill` for. Omit it to
  discover structs marked `//gostructor:gen` in the package instead.
- `-dir` — directory holding the package (project root with `-recursive`;
  defaults to the current directory).
- `-recursive` — discover `//gostructor:gen` structs in `-dir` and its
  subdirectories and generate each in place. Can't be combined with `-type`.

The output file is `<lowercased-type>.gs.go` (plus `<type>.gs_test.go`),
wherever the struct is defined.
