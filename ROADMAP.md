# gostructor roadmap

Living plan for gostructor beyond v1.0. Grouped by theme; each item has a
motivation, an API sketch, prior art where relevant, and a rough
value/effort/dependency note. Nothing here is a commitment to a date — it's
the shortlist of what we think is worth building and why, and what we
deliberately won't.

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
```

- File sources watch via `fsnotify` — kept in a `gostructor/watch` (or the
  file-source) module so the core stays dep-free.
- Debounce rapid successive changes; deliver reload errors to `onReload`
  rather than crashing.
- Reloads flow through the same resolution + trace + masking path as the
  initial `Configure`, so observability works live too.

Prior art: viper `WatchConfig`/`OnConfigChange`, koanf `Watch`.

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

## Prioritization (rough)

| Theme | Value | Effort | New deps | Suggested order |
|---|---|---|---|---|
| 1 Observability & masking | High | Medium | none | 1 |
| 6 Examples (existing features) | High | Small | none | 1 (parallel) |
| 2 Hot reload | High | Medium | fsnotify (submodule) | 2 |
| 5 Ergonomics (desc/validate/layout/map) | Medium | Small each | none | opportunistic |
| 3 Git source | High | Large | git CLI / go-git | 3 |
| 4 Config-server adapters | Med-High | Medium | per-adapter | 3–4 |

## Explicitly out of scope

Guardrails so the library keeps its shape:

- **No global singleton / package-level state** (viper-style). Everything is
  an explicit `Option`.
- **No writing config back to disk** (`WriteConfig`). gostructor fills structs;
  it doesn't own or mutate the source of truth.
- **No case-insensitive / fuzzy key matching magic.** Tags address keys
  explicitly; a miss is an error, not a guess.
