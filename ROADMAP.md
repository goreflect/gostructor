# gostructor roadmap

Living plan for gostructor from v1.0 **up to and including v2.0** — everything
in this document is scoped to what we intend to land by the 2.0 release; work
beyond that lives elsewhere. Grouped by theme; each item has a motivation, an
API sketch, prior art where relevant, and a rough value/effort/dependency note.
Nothing here is a commitment to a date — it's the shortlist of what we think is
worth building on the road to 2.0, and what we deliberately won't.

## Where we are (v1.0)

Shipped and stable — the baseline everything below builds on:

- **Tag-driven fill** from any mix of sources, with **per-field priority**
  (`cf_priority`) — the core differentiator vs. merge-everything-into-one-map
  libraries.
- **Strict, lossless conversion**: fractional-float-into-int, width overflow,
  negative-into-unsigned, and non-finite floats are hard errors, never silent
  truncation/wraparound. Named scalar types, `time.Duration`,
  `encoding.TextUnmarshaler` (`time.Time`, `net.IP`, ...), pointers, arrays,
  nested slices, and `[]Struct`/`map[string]Struct` are supported.
- **Typed error taxonomy**: `ErrInvalidTarget`, `*NotResolvedError`
  (wraps `ErrFieldNotResolved`), `*SourceError`, `*ConvertError`,
  `*HookError`, all satisfying `FieldError` and unwrapping to the real cause.
- **Zero-dependency core**; YAML/TOML/HOCON/Vault each in their own module.
- **`Source` interface** (`Tag()` + `Resolve()`) as the public extension
  point, registered via `WithSources`.

## Guiding principles (what fits this library)

These are the yardsticks for every proposal below:

1. **Zero-dependency core.** New third-party deps live in their own module,
   never in the root. If a feature needs `fsnotify`/`go-git`/an HTTP client,
   it ships as `gostructor/<feature>`.
2. **Explicit, no global singleton.** Everything is a per-call `Option`. We do
   not adopt viper's package-level global state.
3. **Per-field priority stays the spine.** Features compose with the
   source-ordering model rather than around it.
4. **Typed and predictable.** No case-insensitive key magic, no silent
   coercion. Surprises are errors.

---

## Theme 1 — Observability & debugging (priority)

**Motivation.** A config layer is a black box exactly when you most need to
trust it ("why is `Port` 8080 and not what's in my file?"). Users should be
able to see, in detail, *what* each field resolved to, *where* it came from,
*in what order* sources were tried, and *why* the losing ones lost — with
secrets safely masked so the trace can be logged or pasted into a ticket.

### 1a. Resolution trace / "explain" mode

Capture a structured, per-field account of resolution and expose it both as a
returned report and as `slog` output.

```go
// Opt into capturing a trace; nil-cost when off.
cfg, report, err := gostructor.ConfigureWithReport(&Config{}, opts...)

type Report struct{ Fields []FieldResolution }

type FieldResolution struct {
    Field    string        // struct field name
    Type     string        // Go type
    Attempts []Attempt     // every source considered, in the order tried
    Winner   string        // tag that produced the value, or ""
    Raw      any           // value the winning source returned (masked if secret)
    Value    any           // converted, field-typed value  (masked if secret)
    Outcome  string        // resolved | default | unresolved | error
}

type Attempt struct {
    Tag    string // e.g. "cf_env"
    Status string // not-tagged | not-found | used | error
    Detail string // env var name / file path / error summary
}

func (r *Report) String() string // pretty tree/table for humans
```

- Human view is a per-field tree: `Port int  ⇐ cf_env(APP_PORT)=8080  [cf_json: not-found, cf_default: skipped]`.
- Machine view: the `Report` struct, or emitted as structured `slog` records
  when `WithLogger` is set and tracing is on.
- **Provenance map** (`field → winning source`) for the current environment
  is a trivial projection of the report — the "dependency view" of how *this*
  environment assembled the config.

### 1b. Debug flags — toggle verbosity without code changes

Let operators crank up detail via environment/CLI, not just code:

```go
gostructor.WithTrace()                 // programmatic
// GOSTRUCTOR_DEBUG=trace              // full resolution trace at runtime
// GOSTRUCTOR_DEBUG=1                  // debug-level slog
gostructor.RegisterFlags(flag.CommandLine) // wires -gostructor.debug etc.
```

Prior art: cleanenv's `GetDescription()`/`FUsage()` for config *shape*; this
extends that idea to config *resolution*.

### 1c. Data masking (secrets-aware output)

Sensitive values must never leak into logs, the resolution report, or error
messages. Mark fields sensitive and mask everywhere their value would print.

```go
type Config struct {
    APIKey string `cf_vault:"svc/prod#api-key" cf_secret:""`
}

gostructor.WithMasker(func(field gostructor.FieldContext, v any) string {
    return "••••" // default: full redaction; e.g. reveal last 4
})
```

Masking must cover **all** value-printing paths:

- `slog` debug/trace records,
- the `Report` (`Raw`/`Value` of secret fields),
- **`*ConvertError.Value`** — today it prints `%#v` of the raw value, which
  for a secret field would expose it; the masker has to apply here too.

This couples with the error taxonomy: a secret field's `ConvertError` reports
"cannot convert ••••" while still naming the field and cause.

**Value: high. Effort: medium (core, no deps). Milestone: next.**

---

## Theme 2 — Hot reload / live configuration

**Motivation.** Currently `Configure` runs once. Long-lived services want to
pick up changed values without a restart. This is also the prerequisite for
the git and config-server polling below.

```go
// Optional interface a Source may implement.
type Watchable interface {
    Watch(ctx context.Context, onChange func()) error
}

// Re-runs Configure whenever any watchable source signals a change.
func Watch[T any](ctx context.Context, target *T,
    onReload func(*T, error), opts ...Option) error

gostructor.WithDebounce(200 * time.Millisecond) // coalesce change bursts
gostructor.WithValidate(func(c *Config) error { … }) // gate every reload
```

- File sources watch via `fsnotify` — kept in a `gostructor/watch` (or the
  file-source) module so the core stays dep-free.
- Reloads flow through the same resolution + trace + masking path as the
  initial `Configure`, so observability works live too.

**Debounce (anti-flap).** A single logical change often arrives as a *burst*
of low-level events: an editor writes a file in several `write`s, `fsnotify`
emits `WRITE`+`CHMOD`+`RENAME` for one save, and cloud/config-server watchers
push a batch of key events at once. Without coalescing, we'd re-resolve the
whole struct dozens of times a second. `WithDebounce(d)` (default a small
non-zero window, e.g. 200ms) collapses a quiet-for-`d` burst into **one**
reload — the config is only re-read once the source has settled. This is the
Argus/konf lesson made explicit rather than incidental.

**Transactional reload (last-known-good).** A reload must never leave the app
running on a half-applied or invalid config. The sequence is atomic from the
caller's view:

1. Resolve into a **fresh** target — the live struct is never mutated in place.
2. Run `WithValidate` (whole-struct) + per-field hooks against that fresh copy.
3. **Only on success** publish it via `onReload(newCfg, nil)` (an atomic
   pointer swap under the hood); on failure, the **previously good config keeps
   serving** and the error is delivered — `onReload(nil, err)` and, if
   `WithLogger` is set, an `slog` error record — rather than crashing or
   swapping in a broken value.

So a bad edit to a watched file is a logged, non-fatal event; the service holds
its last-known-good configuration until a subsequent valid change arrives.

Prior art: viper `WatchConfig`/`OnConfigChange`, koanf `Watch`, Argus
(batching/debounce), konf (validated live reload).

**Value: high (explicitly requested). Effort: medium. Deps: fsnotify (in submodule).**

---

## Theme 3 — Git as source of truth

**Motivation.** Store configuration in a git repo, treat **a branch/tag/ref
as a config version**, and drive a running service from it: pull the ref,
snapshot the resolved config somewhere durable, then serve from the snapshot
and periodically re-check upstream for drift — with the ability to switch
version at runtime.

Shape (`gostructor/git` submodule, implementing `Source` + `Watchable`):

```go
src := git.New(git.Options{
    Repo:      "https://…/config.git",
    Ref:       "release/2025.10",     // branch/tag = version
    Path:      "services/api/config.yaml",
    Snapshot:  git.DirStore("/var/lib/app/config"), // or a user callback
    Poll:      30 * time.Second,      // re-check upstream for drift
})
cfg, _ := gostructor.Configure(&Config{}, gostructor.WithSources(src, gostructor.Default()))

src.SetVersion("release/2025.11")     // switch version at runtime → triggers reload
```

Design points to settle:

- **Snapshot store** as an interface: default a directory store ("файлопомойка"),
  but pluggable (user decides where — local dir, object store via their own
  impl). The snapshot is what the service actually reads, so a transient git
  outage doesn't take the app down.
- **Drift detection**: poll the remote ref's commit SHA on an interval; on
  change, re-snapshot and reload (via Theme 2). Compare by SHA, not content.
- **Runtime version switch**: `SetVersion(ref)` re-points and reloads; the
  currently-serving snapshot stays until the new version is fully fetched.
- **git access**: lean toward shelling out to the `git` CLI (zero Go deps,
  reuses the user's existing auth/credentials) vs. `go-git` (pure Go, heavier
  dep, no external binary). **Open decision** — CLI is the current lean.
- Provenance in the trace names the ref + commit SHA a value came from.

**Value: high (headline request). Effort: large. Deps: git CLI or go-git.**

---

## Theme 4 — Config servers & remote sources

**Motivation.** "Can I write an adapter for Spring Cloud Config Server / any
config server?" — **yes, that's exactly what the `Source` interface is for**
(Vault is precisely such an adapter). This theme is about proving the pattern
with first-class adapters and examples.

- **Spring Cloud Config Server** adapter: `GET /{app}/{profile}[/{label}]`,
  parse the returned property sources, resolve keys by tag. Own submodule.
- **Generic HTTP/kv** adapter and, later, **etcd/consul** (the roadmap's
  earlier `cf_server_kv`/`cf_server_file`), each a `Source` (+ optionally
  `Watchable` for push/poll updates).
- A documented **"writing an adapter" guide** + runnable example with a mock
  server, so users can target their own config service.

**Value: medium-high. Effort: medium per adapter. Deps: per-adapter (in submodules).**

---

## Theme 5 — Ergonomics worth borrowing (viper / cleanenv)

Small, high-leverage additions that fit the model:

- **Self-documenting config** — `cf_desc` tag + `Describe[T]()` / `Usage[T]()`
  that walks the struct and prints each field, its sources, default,
  required-ness, and description. Zero-dep, pairs naturally with Theme 1.
  (cleanenv `GetDescription`/`env-description`.)
- **Whole-struct validation** — `WithValidate(func(*T) error)` run after fill,
  complementing per-field `WithHook`. (cleanenv `Updater.Update()`.)
- **Custom time layout** — `cf_layout:"2006-01-02"` for non-RFC3339 dates,
  rounding out the `TextUnmarshaler`/`time.Time` support. (cleanenv `env-layout`.)
- **In-memory override source** — `gostructor.Map(map[string]any)` as a
  highest-priority `Source`, great for tests and programmatic overrides.
  (viper `Set`.)
- **Explicit optional/required semantics** — a `cf_optional` (leave zero when
  unresolved) to complement the current "tagged ⇒ required" default.

**Value: medium. Effort: small each. Deps: none (core).**

### 5a. Config-file template generation (`GenerateTemplate[T]`)

**Motivation.** When a service starts and its config file is *missing*, the
best possible UX for an operator is not a stack trace — it's a ready-to-edit,
fully-commented sample file with every field, its default, and a human
description already filled in. cleanenv does exactly this (writes a default
YAML from the struct); it's a small feature that infra engineers love. We take
the idea further by reusing the `cf_desc` descriptions (Theme 1/5) as inline
comments and supporting multiple output formats.

```go
// Emit a valid, commented config file straight from the struct + tags.
yaml, err := gostructor.GenerateTemplate[Config](gostructor.TemplateYAML)
toml, err := gostructor.GenerateTemplate[Config](gostructor.TemplateTOML)

// Common pattern: self-heal a missing file on first run.
if errors.Is(err, os.ErrNotExist) {
    _ = os.WriteFile("config.yaml",
        must(gostructor.GenerateTemplate[Config](gostructor.TemplateYAML)), 0o644)
}
```

Given:

```go
type Config struct {
    Port    int    `cf_json:"server.port" cf_default:"8080" cf_desc:"HTTP listen port"`
    DBURL   string `cf_json:"db.url" cf_desc:"Postgres DSN"`
    Timeout time.Duration `cf_json:"timeout" cf_default:"30s" cf_desc:"per-request timeout"`
}
```

`GenerateTemplate[Config](TemplateYAML)` yields:

```yaml
server:
  # HTTP listen port
  port: 8080
db:
  # Postgres DSN            (required — no default)
  url: ""
# per-request timeout
timeout: 30s
```

- Walks the struct the same way `Configure` does, so nesting, `cf_default`
  values, and `cf_desc` text stay in sync with real resolution — no second
  source of truth.
- Marks required fields (no `cf_default`, not `cf_optional`) in the comment so
  the operator knows what they *must* fill.
- Zero-dep for YAML/TOML emission when paired with the respective format
  submodule; the core exposes a neutral document model the formatters render.

Prior art: cleanenv's default-file generation; `GenerateTemplate` extends it
with descriptions-as-comments and multi-format output.

### 5b. First-class `cf_env` (feature parity with `caarlos0/env`)

**Motivation.** Today teams often pair a config library with a *dedicated* env
parser (`caarlos0/env`, kelseyhightower/envconfig) because the built-in env
handling is anemic. We want `cf_env` to be good enough — as fast and as
featureful — that reaching for a second library is never necessary. The env
source stays a clean, isolated `Source` (no coupling to file parsing), but
gains the ergonomics power users expect:

```go
type Config struct {
    // prefix + nested expansion → APP_DB_HOST, APP_DB_PORT
    DB struct {
        Host string `cf_env:"HOST"`
        Port int    `cf_env:"PORT" cf_default:"5432"`
    } `cf_env_prefix:"DB_"`

    // list/map parsing with a custom separator
    Hosts []string          `cf_env:"HOSTS" cf_env_sep:","`
    Meta  map[string]string `cf_env:"META"`               // k1:v1,k2:v2

    // read the value from a file whose path is in the env var (secrets)
    APIKey string `cf_env:"API_KEY" cf_env_file:"true"`

    // expand ${OTHER_VAR} references inside the value
    URL string `cf_env:"URL" cf_env_expand:"true"`
}

env := envsource.New(envsource.WithPrefix("APP_"))
```

Parity checklist we hold ourselves to vs. `caarlos0/env`: env prefixes,
nested-struct prefix composition, slices/maps with configurable separators,
`*_FILE` indirection for secrets, `${VAR}` expansion, and required/default
semantics — all while keeping gostructor's **strict, lossless conversion** (a
non-numeric `PORT` is a `*ConvertError`, not a silent `0`), which is precisely
where the third-party parsers are permissive and we are not. Performance: the
env source is allocation-lean and is a prime beneficiary of Theme 7 codegen.

Prior art: `caarlos0/env` (the feature bar), kelseyhightower/envconfig
(prefix/expand conventions).

### 5c. Auto-generated CLI flags from the struct (aconfig-style)

**Motivation.** A struct already *is* a complete description of the config
surface; making the operator hand-write a `flag` per field is redundant.
aconfig derives `--db-port` from `DB.Port` automatically. We offer the same as
an opt-in flag `Source`, so CLI flags compose with env/file/default through the
normal per-field priority model instead of being a separate parallel system.

```go
// Derive one flag per leaf field: DB.Port → -db.port (or --db-port),
// help text from cf_desc, default from cf_default.
fs := flag.NewFlagSet("app", flag.ContinueOnError)
flags := flagsource.New(fs, flagsource.FromStruct[Config]())
_ = fs.Parse(os.Args[1:])

cfg, _ := gostructor.Configure(&Config{},
    gostructor.WithSources(flags, envsource.New(), gostructor.Default()))
// -h prints: -db.port int   HTTP listen port (default 8080)
```

- Flag names are derived from the field path (config-key-cased, e.g.
  `db.port`/`db-port`); `-h`/`--help` text comes from `cf_desc`, defaults from
  `cf_default` — one struct drives fill *and* usage.
- It's just another `Source`: a flag explicitly set on the command line beats
  env/file per `cf_priority`; an unset flag simply doesn't resolve, so defaults
  and other sources still apply. No global flag state.
- `RegisterFlags` (Theme 1b) and this share the derivation logic.

Prior art: aconfig (flags-from-struct), kong (struct-driven CLI), Go stdlib
`flag`.

**Value: medium-high (removes whole categories of second dependencies).
Effort: small–medium each. Deps: none for template/flags; format submodules for
YAML/TOML template output.**

---

## Theme 6 — Richer examples

Runnable, zero-setup examples so the library is easy to poke at:

- One per core source (`cf_env`, `cf_default`, `cf_json`, `cf_ini`) that
  writes its own sample config to a temp dir.
- Combined sources + `cf_priority` (order switching via `GOSTRUCTOR_PRIORITY`).
- Hooks (validation + transform) and the error taxonomy (trigger each error
  type and classify it with `errors.As`).
- File formats (yaml/toml/hocon) and Vault (illustrative).
- Once built: hot-reload, git source, and config-server adapter examples.

**Value: high (low-friction onboarding). Effort: small–medium. Deps: none for core examples.**

---

## Theme 7 — Code generation (`gostructor-gen`)

**Motivation.** The core fill path is reflection-driven: fast to *build*, but
every `Configure` walks `reflect.Value`s, boxes scalars into `any`, and
allocates on the hot path. For services that reload config frequently (Theme
2/3) or embed gostructor in latency-sensitive startup, reflection is the tax
we pay for genericity. Code generation removes it: parse the target struct at
`go generate` time and emit a bespoke, fully-typed `Fill` method that reads
each source directly into each field — **no runtime reflection, no `any`
boxing, near-zero allocation** — while reusing the exact same `Source`,
priority, conversion, and error-taxonomy semantics as the reflective engine.
The reflective path stays the default and the always-correct reference; the
generated path is an opt-in performance tier that must produce
byte-for-byte-equivalent results.

**API / Usage sketch.**

```go
//go:generate go run github.com/goreflect/gostructor/cmd/gostructor-gen -type Config

type Config struct {
    Port    int           `cf_env:"APP_PORT" cf_default:"8080"`
    DBURL   string        `cf_env:"DB_URL" cf_json:"db.url"`
    Timeout time.Duration `cf_env:"TIMEOUT" cf_default:"30s"`
}
```

`gostructor-gen` emits `config_gostructor.go`:

```go
// Code generated by gostructor-gen. DO NOT EDIT.

// Fill resolves every field of Config from the given sources, honoring
// per-field priority, without runtime reflection. Semantics match
// gostructor.Configure(&c, opts...) exactly.
func (c *Config) Fill(opts ...gostructor.Option) error { … }
```

- The generated `Fill` is a drop-in accelerator: `gostructor.Configure` detects
  a `Filler` (`interface{ Fill(...Option) error }`) on the target and dispatches
  to it, so callers keep one entry point and get the fast path for free.
- Conversions call into the same `internal/convert` primitives (generated as
  direct, monomorphized calls, e.g. `convert.StringToDurationStrict`), so the
  lossless/strict guarantees and `*ConvertError` taxonomy are identical.
- Generation is total: unsupported field shapes fail the generator (build-time
  error) rather than silently degrading — you never get a partial `Fill`.
- Tracing/masking (Theme 1) are honored via the same option plumbing; the
  generator emits trace hooks guarded by a compile-time-cheap branch.

**Three execution modes (adaptive by default).** Codegen must never be a
*requirement* — a project that hasn't run `go generate` has to keep working
exactly as today. So `Configure` supports an explicit engine mode, defaulting
to adaptive:

```go
type Engine int
const (
    EngineAdaptive   Engine = iota // default: use generated Fill if present, else reflect
    EngineReflection                // force the reflective path (ignore any Fill)
    EngineCodegen                   // require a generated Fill; error if absent
)

func WithEngine(Engine) Option
```

| Mode | Behaviour | Use it when |
|---|---|---|
| `EngineAdaptive` *(default)* | Dispatch to the generated `Fill` **iff** the target implements `Filler`; otherwise fall back to reflection — transparently, same result either way. | Almost always. Zero-config: fast where generated, correct everywhere. |
| `EngineReflection` | Always reflect, even if a `Fill` exists. | Debugging a suspected codegen/reflection divergence; a stale generated file. |
| `EngineCodegen` | Require a generated `Fill`; if the target doesn't implement `Filler`, return `ErrNoGeneratedFiller` instead of silently reflecting. | CI / perf-critical builds that must *guarantee* the reflection-free path (e.g. an allocation budget) and want a hard failure if someone forgets `go generate`. |

```go
// Default — works with or without generated code:
cfg, err := gostructor.Configure(&Config{}, gostructor.WithSources(env.New()))

// Guarantee the fast path in a latency-critical service; fail the build/boot
// loudly if `go generate` wasn't run:
cfg, err := gostructor.Configure(&Config{},
    gostructor.WithEngine(gostructor.EngineCodegen),
    gostructor.WithSources(env.New()))
```

To keep the generated and reflective paths from drifting, the generator also
emits a golden test (`TestFillMatchesReflection`) that fills the same struct
both ways across a fixture matrix and asserts byte-for-byte-equal results and
identical error classification — a divergence fails CI, so `EngineAdaptive`'s
"same result either way" promise is enforced, not assumed.

**Documentation & examples (ship with the feature).** Because codegen changes
the mental model, it gets first-class docs, not just a flag mention:

- A `docs/codegen.md` guide: when to bother (benchmarks showing the alloc/ns
  delta vs. reflection), the `go:generate` line, the three modes and how to
  pick, and how the golden test guards correctness.
- A runnable `examples/codegen/` module: a struct with `//go:generate`, the
  committed generated file, and a `Makefile`/`go generate ./...` target so a
  reader can regenerate and `go test -bench` locally.
- A "migrating an existing config to codegen" recipe (add the directive,
  generate, flip to `EngineCodegen` in prod builds) and a troubleshooting note
  for `ErrNoGeneratedFiller`.

Prior art: `easyjson`/`ffjson` (reflection-free JSON, with a reflective
fallback), `sqlc`, `ent`, protobuf generated marshalers, and Enflag (compile-
time-safe config without struct-tag reflection — codegen is our answer that
keeps the ergonomic tagged struct *and* gets the reflection-free speed).

**Value: high (headline perf story vs. viper; compile-time-safe answer to
Enflag). Effort: large. Deps: none at runtime; `go/ast`+`go/types` in the
generator tool.**

### 7a. Codegen-only ergonomics (the config pains codegen uniquely fixes)

**Motivation.** Four recurring complaints from teams doing config in Go are
awkward to solve well in the reflective engine but fall out almost for free once
we parse the struct at `go generate` time. Codegen "develops our hands": the
generator has the full AST + type info, so it can infer, check, and emit code
that reflection would have to pay for at runtime (or couldn't do at all). These
land **with** Theme 7's generator, reusing the same `Source`/priority/convert
semantics.

**1. Kill tag duplication — convention-first mapping (not one mega-tag).**
The pain is `cf_env:"DB" cf_json:"db" cf_dyn:"DB"` — visual noise, three keys to
keep in sync. **Decision (settled): keep several conventional `cf_*` tags, infer
their values from the field name, and require a tag only for exceptions.** We
deliberately reject folding everything into one `gostructor:"key=…,sources=…"`
tag, because a single shared key can't express gostructor's defining feature —
*a different key per source with a per-field priority order* (`cf_env:"DB_URL"`
vs `cf_json:"db.url"`, tried in a chosen order). A mega-tag would force an
immediate escape hatch and give us two syntaxes; it's also less `go vet`- and
IDE-legible than plain tag keys. Instead:

```go
//gostructor:naming=snake_case   // file-level default: DatabasePort → DATABASE_PORT / database_port

type Config struct {
    DatabasePort int                              // inferred for every enabled source, no tags
    LegacyDSN    string `cf_env:"OLD_DB_URL"`     // exception: override only the source that breaks the convention
    APIKey       string `cf_secret:"" cf_required:""`
}
```

- A file-level `//gostructor:naming=<snake_case|kebab|screaming_snake|…>` sets
  the default derivation per source (env → `SCREAMING_SNAKE`, yaml/json →
  `snake_case`, etc.). The generator emits the concrete key into the generated
  `Fill`, so there's no runtime name-mangling cost.
- An explicit `cf_*` tag always wins over the inferred key — you annotate only
  the fields whose real key deviates. Net effect: ~90% fewer tags, and every
  surviving tag stays a plain, tool-legible key.
- Reflective engine parity: the same inference can run in the reflective path
  behind an opt-in (`WithNaming(...)`), so behavior matches whether or not you
  ran `go generate`.

**2. `required` out of the box — a real check, not a convention.** Today
"tagged ⇒ required" is enforced only as `*NotResolvedError` when *no* source
produced anything; there's no way to say "must be non-zero after all sources
run." Add `cf_required:""`. The generated `Fill` emits a direct post-resolution
`if` per required field: if the field is still the zero value for its type after
every source (and default) was tried, return a precise error naming the field
and the sources checked:

```go
fmt.Errorf("gostructor: missing required config field %q (checked: cf_yaml, cf_env)", "DB")
```

No panics, no reflection at check time — just generated straight-line code. Maps
onto the existing taxonomy as a distinct `*RequiredError` (satisfies
`FieldError`), so callers classify it with `errors.As`.

**3. Auto-wire `Validate()` — never forget to call it.** The generator inspects
the target (and its nested structs) for a `Validate() error` method; if present,
it **emits the call at the tail of `Fill`**, after all fields resolve and all
`cf_required` checks pass. `cfg, err := LoadConfig()` with `err == nil` then
guarantees the struct is both fully assembled *and* business-rule-valid — the
`WithValidate` (Theme 5) hook made zero-cost and un-forgettable via codegen.
Reflective-engine equivalent stays the explicit `WithValidate(...)` option.

**4. Config/test drift — generate the canonical sample artifact.** This is
Theme 5a (`GenerateTemplate`) turned into a **build-time artifact** by the
generator: on `go generate`, also emit `config.sample.yaml` / `.env.example`
straight from the struct's fields, tags, `cf_default` values, and `cf_desc`
comments. Because it regenerates from the single source of truth, a test can
diff a committed golden sample against the freshly generated one and **fail CI
when someone adds a field but forgets the sample** — the "green tests, prod
panics on a missing key" class of bug becomes a build failure. The generator
also emits an optional `TestConfigSampleUpToDate` alongside the golden test from
Theme 7 to enforce this automatically.

**Value: high (hits the four most-cited Go-config pains). Effort: medium on top
of Theme 7's generator; small each. Deps: none at runtime; rides Theme 7's
`go/ast`+`go/types`.**

---

## Theme 8 — Native DI integration (`uber-go/fx`)

**Motivation.** In DI-structured apps, wiring config in is pure boilerplate:
write a provider that takes a `context`, constructs the target, calls
`Configure`, maps the error, and returns the struct — repeated per config type,
per service. gostructor already produces exactly the "one fully-resolved value"
that a DI graph wants to inject. A thin, well-typed bridge collapses that
boilerplate to a single line and, by composing with Theme 2, can keep the
injected value live across reloads.

**API / Usage sketch.** A dedicated submodule `gostructor/gofx` (keeps the
`go.uber.org/fx` dep out of the core), exposing a generic provider:

```go
import (
    "go.uber.org/fx"
    gostructfx "github.com/goreflect/gostructor/gofx"
)

app := fx.New(
    fx.Provide(
        gostructfx.New[*Config](
            gostructor.WithSources(env.New()),
            gostructfx.WithOnReload(func(cfg *Config) {
                log.Println("Config reloaded live!")
            }),
        ),
    ),
    fx.Invoke(func(cfg *Config) { /* fully-resolved, ready to use */ }),
)
```

- `gostructfx.New[*Config](opts...)` returns an `fx.Provide`-compatible
  constructor that builds, configures, and yields `*Config` into the graph;
  resolution errors surface as `fx` startup errors (fail-fast, no half-wired
  app).
- `WithOnReload` opts the provider into Theme 2: gostructfx runs `Watch` under
  an `fx.Lifecycle` hook (started `OnStart`, cancelled `OnStop`) and pushes new
  values through the callback — the DI-friendly way to get live config without
  leaking a goroutine.
- Ships a ready-made `gostructfx.Module(opts...)` (`fx.Module("gostructor", …)`)
  for apps that prefer module-level wiring over per-provider calls.

Prior art: `fx`'s own `fx.Provide`/`fx.Module` conventions; `zap`'s
`fx`-friendly constructors; the general "provider function" pattern.

**Value: high (removes real boilerplate for fx users). Effort: medium.
Deps: `go.uber.org/fx` (in `gostructor/gofx` submodule).**

---

## Theme 9 — Decentralized & P2P sources

**Motivation.** Not every deployment has a central config authority (Vault,
Consul, Spring Cloud Config). Edge, offline-first, and P2P nodes typically
carry their **own local store** — an embedded SQLite/LevelDB database — and
must react to changes written to it by a local agent, a sync process, or a
gossip/P2P event, with no network round-trip. The `Source` + `Watchable`
contracts already model exactly this: read on `Resolve`, signal on `Watch`.
This theme delivers first-class embedded-store sources.

**API / Usage sketch.** `gostructor/sqlitesource` (submodule), implementing
`Source` + `Watchable` (Theme 2):

```go
src := sqlitesource.New(db, "SELECT key, value FROM settings",
    sqlitesource.WithPoll(5*time.Second),   // or:
    sqlitesource.WithTrigger("settings"),    // hook a table trigger + notify
)

cfg, _ := gostructor.Configure(&Config{},
    gostructor.WithSources(src, gostructor.Default()))

// Live: when a row in `settings` changes, Watch fires and Configure re-runs.
gostructor.Watch(ctx, &Config{}, onReload, gostructor.WithSources(src))
```

```go
type Config struct {
    PeerLimit int `cf_sqlite:"peer.limit" cf_default:"32"`
}
```

- `Resolve` runs the key/value query once and indexes it; each `cf_sqlite` tag
  is a key lookup. The query is user-supplied, so any settings schema works.
- **Change detection, two strategies:** *polling* (re-run the query on an
  interval, diff against the last snapshot — universal, zero DB features
  required) or *trigger-driven* (install an `AFTER INSERT/UPDATE/DELETE`
  trigger on the settings table that bumps a version row/channel the source
  watches — lower latency, no busy-poll). `WithPoll`/`WithTrigger` pick.
- A LevelDB variant (`gostructor/leveldbsource`) follows the same shape for
  nodes that use an LSM store instead of SQLite.
- Because it goes through `Watchable`, reloads reuse Theme 2's debounce, error
  delivery, trace, and masking — a P2P node gets hot-reload "for free."

Prior art: `koanf`'s file/kv providers; viper's remote providers (but those
assume a network config server, not an embedded local DB).

**Value: medium-high (unlocks edge/P2P deployments). Effort: medium.
Deps: a SQLite/LevelDB driver (in the respective submodule).**

---

## Theme 10 — Migration CLI (`gostructor-migrate`)

**Motivation.** The single biggest barrier to adopting gostructor is not the
runtime — it's the **mechanical cost of migrating an existing codebase**:
hundreds of `mapstructure`/`env`/`yaml` struct tags to rewrite and dozens of
`viper.Get*` call sites and init blocks to replace. That work is tedious,
error-prone by hand, and easy to defer forever. A source-to-source tool that
does the rote 90% turns "someday" into an afternoon. Crucially, the user
should not have to *know or name* which config library they're on — the Go
config ecosystem is huge (viper, koanf, cleanenv, envconfig, `caarlos0/env`,
kong, fig, gonfig, aconfig, confita, onion, gookit/config, …). The tool's
first job is to **auto-detect the incumbent(s)** and only then rewrite.

**API / Usage sketch.** A Go-AST-based CLI (`gostructor/cmd/gostructor-migrate`)
that first *detects* the config library in use, then scans packages, rewrites
tags, and best-effort rewrites init calls:

```console
$ gostructor-migrate detect ./...          # just report what it found
detected config libraries in ./...:
  • spf13/viper            (imports: 3 pkgs, calls: 27)   confidence: high
  • kelseyhightower/envconfig (tags: `envconfig` on 4 structs) confidence: high
  • caarlos0/env/v11        (tags: `env`,`envDefault` on 9 structs) confidence: high
run `gostructor-migrate run ./... --dry-run` to preview the rewrite.

$ gostructor-migrate run ./...             # auto-detect + rewrite (all incumbents)
$ gostructor-migrate run ./... --dry-run   # preview unified diff only
$ gostructor-migrate run ./internal/config --from=viper --write  # pin one source
```

**Auto-detection.** `--from` becomes *optional* — an override, not a
requirement. Detection runs over `go/packages` (full type info, not regex) and
fuses three signals into a per-library confidence score:

- **Import paths** — the strongest signal. A registry maps module paths to a
  known incumbent: `github.com/spf13/viper`, `github.com/knadh/koanf/...`,
  `github.com/ilyakaznacheev/cleanenv`, `github.com/kelseyhightower/envconfig`,
  `github.com/caarlos0/env/v11`, `github.com/alecthomas/kong`,
  `github.com/kkyr/fig`, `github.com/cristalhq/aconfig`, `github.com/heetch/confita`,
  `github.com/gookit/config`, … (the awesome-go "configuration" set seeds it).
- **Struct-tag signatures** — which tag keys actually appear: `mapstructure`
  ⇒ viper/mapstructure decode, `env`+`envDefault` ⇒ `caarlos0/env`, `env`+
  `env-default`+`env-required` ⇒ cleanenv, `envconfig`/`default`/`required`
  ⇒ kelseyhightower, `koanf` ⇒ koanf, `conf`/`fig` ⇒ fig, etc.
- **Call-site fingerprints** — resolved selector calls: `viper.Unmarshal`,
  `viper.Get*`, `envconfig.Process`, `env.Parse`, `cleanenv.ReadConfig`,
  `koanf.Unmarshal`, `kong.Parse`, … each pins a specific library and its
  init pattern.

Multiple incumbents in one repo is common (e.g. viper for files + `caarlos0/env`
for env); detection reports **each** and the rewriter handles them together.
Anything below a confidence threshold is listed but left untouched unless
`--from` names it explicitly.

**Rewrite**, driven by the matched adapter(s):

- **Rewrites tags** to gostructor equivalents — `mapstructure`→`cf_json`/
  `cf_yaml` (by decode context), `env`→`cf_env`, `envDefault`/`env-default`/
  `default`→`cf_default`, `env-required`/`required`→(drop, since tagged ⇒
  required), `yaml`/`toml`→`cf_yaml`/`cf_toml` — preserving key names,
  ordering, and gofmt-clean formatting.
- **Rewrites init sites** where it can prove the shape: `viper.Unmarshal(&c)`
  / `envconfig.Process("", &c)` / `env.Parse(&c)` / the `viper.Get*` +
  manual-assign pattern → `gostructor.Configure(&c, …)`, emitting a
  `WithSources(...)` best-guess from the providers each incumbent implied.
- **Flags what it can't** with `// TODO(gostructor-migrate):` comments instead
  of guessing — anything ambiguous (dynamic keys, `viper.Sub`, runtime `Set`,
  a library with no gostructor equivalent yet) is surfaced for a human, never
  silently mistranslated.
- `--dry-run` prints a unified diff; `--write` applies in place; a summary
  reports detected libraries, tags rewritten, call sites migrated, TODOs left.

Each supported incumbent is a small **detector+rewriter plugin** behind one
interface, so covering a new library from the list above is an additive
contribution, not a core change.

Prior art: `gofmt -r`, `go fix`, `golang.org/x/tools/go/analysis` rewriters,
`eg` (example-based refactoring), and `go mod why`/dependency-graph tools for
the import-based detection idea — same "AST rewrite as a first-class tool"
lineage.

**Value: high (removes the #1 adoption barrier, zero guessing by the user).
Effort: large. Deps: `golang.org/x/tools` in the tool only; nothing at
runtime.**

---

## Theme 11 — Backward-compatibility adapters (viper / cleanenv)

**Motivation.** Theme 10 rewrites source; some teams can't or won't touch
working business logic on day one. For them the fastest path is a **shim**:
keep the old call sites verbatim, but swap the engine underneath. An adapter
that implements viper's read surface (or cleanenv's) on top of gostructor's
strict, no-global-state core lets a project "migrate in five minutes" and defer
the real refactor — while immediately gaining lossless conversion and the typed
error taxonomy. It's the low-risk on-ramp that Theme 10 later finishes.

**API / Usage sketch.** `gostructor/compat/viper` — a `*Viper`-shaped façade
backed by gostructor:

```go
import viper "github.com/goreflect/gostructor/compat/viper"

v := viper.New()                     // same constructor surface
v.SetConfigFile("config.yaml")
v.AutomaticEnv()
_ = v.ReadInConfig()

port := v.GetInt("server.port")      // strict conversion under the hood:
                                     // a non-numeric value is an error, not 0
_ = v.Unmarshal(&cfg)               // routes through gostructor.Configure
```

```go
import cleanenv "github.com/goreflect/gostructor/compat/cleanenv"

// existing cleanenv-tagged struct, unchanged
err := cleanenv.ReadEnv(&cfg)        // gostructor engine, cleanenv signature
```

- Implements the **common read subset** — `Get*`, `Unmarshal`, `ReadInConfig`,
  `AutomaticEnv`, `SetDefault`, `BindEnv` — enough to relink real apps. The long
  tail (`WriteConfig`, `Sub`, remote KV, live `Set` mutation) is deliberately
  **not** emulated: those either violate our out-of-scope guardrails or need an
  explicit port. Unsupported calls fail loudly, not silently.
- **Strictness is opt-in-to-lax:** by default the shim keeps gostructor's hard
  errors (so a bad value that viper would've swallowed as a zero now surfaces);
  a `viper.WithLegacyLax()` toggle restores viper's coercion for a staged
  migration where you fix call sites incrementally.
- No package-level global: `viper.GetViper()`-style singletons map to an
  explicit instance the adapter owns, documented as the one intentional
  deviation.

Prior art: `spf13/viper` and `ilyakaznacheev/cleanenv` public APIs (the
surfaces being shimmed); the general "compatibility façade over a new engine"
pattern (e.g. `log/slog`'s `slog.NewLogLogger` bridging `log`).

**Value: high (five-minute adoption, refactor later). Effort: medium–large.
Deps: none at runtime beyond the adapter submodule.**

---

## Prioritization (rough)

| Theme | Value | Effort | New deps | Suggested order |
|---|---|---|---|---|
| 1 Observability & masking | High | Medium | none | 1 |
| 6 Examples (existing features) | High | Small | none | 1 (parallel) |
| 2 Hot reload | High | Medium | fsnotify (submodule) | 2 |
| 5 Ergonomics (desc/validate/layout/map) | Medium | Small each | none | opportunistic |
| 3 Git source | High | Large | git CLI / go-git | 3 |
| 4 Config-server adapters | Med-High | Medium | per-adapter | 3–4 |
| 11 Compat adapters (viper/cleanenv) | High | Med–Large | none (submodule) | 3 (adoption on-ramp) |
| 8 DI integration (uber-go/fx) | High | Medium | fx (submodule) | 3–4 |
| 9 P2P/embedded sources (SQLite/LevelDB) | Med-High | Medium | DB driver (submodule) | 4 (needs Theme 2) |
| 10 Migration CLI (gostructor-migrate) | High | Large | x/tools (tool only) | 4 |
| 7 Code generation (gostructor-gen) | High | Large | go/ast (tool only) | 5 (after API stabilizes) |

## Explicitly out of scope

Guardrails so the library keeps its shape:

- **No global singleton / package-level state** (viper-style). Everything is
  an explicit `Option`.
- **No writing config back to disk** (`WriteConfig`). gostructor fills structs;
  it doesn't own or mutate the source of truth.
- **No case-insensitive / fuzzy key matching magic.** Tags address keys
  explicitly; a miss is an error, not a guess.
