# Code generation (`gostructor-gen`)

The reflective engine walks `reflect.Value`s on every `Configure` call: fast to
build, but it boxes every scalar into `any` and allocates on the hot path. For
services that reload config often or fill it in a latency-sensitive startup, that
reflection is a tax. `gostructor-gen` removes it: it parses your struct at
`go generate` time and emits a bespoke, fully-typed `Fill` method that resolves
each field directly into its concrete Go type — **no per-fill reflection, no
`any` boxing of the destination, near-zero allocation** — while reusing the exact
same sources, slice-order priority, conversion rules, and error taxonomy as the
reflective engine.

The reflective path stays the default and the always-correct reference. The
generated path is an opt-in performance tier that must produce
byte-for-byte-equivalent results, and a generated golden test enforces that.

## Quick start

Two ways to tell `gostructor-gen` which structs to generate for.

**By name** — a `//go:generate` directive that names the type:

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
discover it, so you only ever hand it a path. This is the better fit when a
project has several config structs across several packages: mark the ones you
want and point the tool at the tree.

```go
//gostructor:gen
type Config struct { /* ... */ }
```

```sh
go generate ./...                    # marker discovery in each package
# or, once, for the whole project:
go run github.com/goreflect/gostructor/cmd/gostructor-gen -dir . -recursive
```

(For marker discovery from a `//go:generate` line, drop the `-type`:
`//go:generate go run .../cmd/gostructor-gen -dir .`.)

Either way you get two files next to the source, named after the struct (lower-
cased) with a short `.gs.go` postfix:

- `config.gs.go` — the `func (c *Config) Fill(opts ...gostructor.Option) error`,
  with each field's parsing split into its own small helper method rather than
  one monolithic body.
- `config.gs_test.go` — `TestFillMatchesReflection_Config`, which fills the
  struct both ways and fails if the results or errors diverge.

Your call site does not change:

```go
cfg, err := gostructor.Configure(&Config{}, gostructor.WithSources(env.New()))
```

`Configure` detects the generated `Fill` (any target implementing
`interface{ Fill(...Option) error }`) and dispatches to it automatically. This is
the default engine, `EngineAdaptive`: fast where generated, correct everywhere.
See a full runnable program in [`examples/codegen`](../examples/codegen).

## The three engine modes

Codegen is never a *requirement* — a package that hasn't run `go generate` keeps
working exactly as before. `WithEngine` chooses how `Configure` resolves:

| Mode | Behaviour | Use it when |
|---|---|---|
| `EngineAdaptive` *(default)* | Dispatch to the generated `Fill` **iff** the target implements the filler interface; otherwise reflect. Same result either way. | Almost always. Zero-config: fast where generated, correct everywhere. |
| `EngineReflection` | Always reflect, even if a `Fill` exists. | Debugging a suspected codegen/reflection divergence, or routing around a stale generated file. |
| `EngineCodegen` | Require a generated `Fill`; if the target has none, return `ErrNoGeneratedFiller` instead of silently reflecting. | CI / perf-critical builds that must *guarantee* the reflection-free path and want a hard failure if someone forgot `go generate`. |

```go
// Guarantee the fast path in a latency-critical service; fail loudly if
// `go generate` wasn't run:
cfg, err := gostructor.Configure(&Config{},
    gostructor.WithEngine(gostructor.EngineCodegen),
    gostructor.WithSources(env.New()))
if errors.Is(err, gostructor.ErrNoGeneratedFiller) {
    log.Fatal("run `go generate ./...` — this build requires the codegen fast path")
}
```

`ConfigureWithReport` always uses the reflective engine: the resolution report is
that engine's view, so it is populated even for a type that has a generated
`Fill`.

## How equivalence is guaranteed

The generated `Fill` is source-agnostic. It does **not** hardcode which sources
exist — it iterates the same `[]Source` you pass through `WithSources`, in the
same order, calling each source's `Resolve`. What it removes is the per-fill
reflect walk (fields are known at generate time) and the reflective conversion:
each resolved value is converted with a direct, typed call into
[`gostructor/gen`](../gen) (`gen.Int`, `gen.Duration`, `gen.StringSlice`, …) and
assigned straight into the field, with no `reflect.Value.Set`.

Those typed conversions are the reflection-free siblings of the reflective
`convert.Value`, sharing the exact overflow, integral-float, and base-10 rules,
so a value that converts (or fails to convert) one way does the same the other
way — including the `*ConvertError` text. Hooks, secret masking, the
not-resolved decision, and the `*SourceError`/`*NotResolvedError`/`*HookError`
taxonomy all run through the same shared runtime the reflective engine uses.

The emitted `TestFillMatchesReflection_<Type>` is the backstop: it fills the same
struct through both engines across a fixture and asserts identical values and
identical error classification. A divergence fails CI, so `EngineAdaptive`'s
"same result either way" promise is enforced, not assumed. Commit the generated
files and run `go generate ./...` in CI (or diff-check them) so a struct change
without a regenerate is caught.

## Supported field types

Generation is **total**: if a struct contains a field shape the generator does
not support, it fails at `go generate` time with a clear message rather than
emitting a partial `Fill`. You never get a silently-incomplete fast path — you
either get a fully-correct one or an error telling you to keep that type on the
reflective engine.

Two conversion tiers, both byte-for-byte equivalent to the reflective engine:

**Reflection-free** (a dedicated typed `gen.*` call, no reflection on the hot
path):

- the string, bool, and sized numeric primitives (`int`/`int8`…`int64`,
  `uint`/`uint8`…`uint64`, `float32`, `float64`),
- `time.Duration`,
- `[]string`.

**Via `gen.Reflective[T]`** (routed through the same reflective core the engine
uses, so parity holds; the reflect cost is paid only for that field, exactly as
the reflective engine would):

- other slices and arrays (`[]int`, `[3]bool`),
- maps (`map[string]int`),
- in-package named types over any of the above.

**Nested structs** are flattened exactly as the reflective engine flattens them:
an *untagged* nested struct field is recursed into and its leaves resolved with
their own keys (`c.Service.Name`). A nested struct that carries a `cfg`/`gos`
tag is atomic to the engine and filled via `gen.Reflective` from an object
source, matching the reflective path.

Still not supported (these fail generation): embedded fields, pointers,
out-of-package named/`SelectorExpr` types other than `time.Duration`, and
`encoding.TextUnmarshaler` types such as `time.Time`. Use the reflective engine
for structs containing them — `EngineAdaptive` means such a struct simply
reflects with no code change.

See [`examples/codegen`](../examples/codegen) for a runnable config that
exercises nested structs, a map, and an `[]int`.

## Flags

```
gostructor-gen [-type Name[,Name2,...]] [-dir .] [-recursive]
```

- `-type` — comma-separated struct type names to generate a `Fill` for. Omit it
  to discover structs marked `//gostructor:gen` in the package instead.
- `-dir` — directory holding the package (project root with `-recursive`;
  default: the current directory).
- `-recursive` — discover `//gostructor:gen`-marked structs in `-dir` and every
  subdirectory, generating each in place. Mutually exclusive with `-type`.

The output file is named `<lowercased-type>.gs.go` (plus `<type>.gs_test.go`),
regardless of which source file the struct lives in.
