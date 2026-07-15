# Migration & v1.0 changes

## What changed in v1.0

- **Two tags instead of many.** The per-source tag family (`cf_env`,
  `cf_json`, `cf_default`, `cf_secret`, ...) is replaced by one routing tag
  `cfg` and one behavior tag `gos`. Each source derives its key from the base
  name via a naming strategy (env → `SCREAMING_SNAKE`, files → the name as
  written), overridable per source in the `cfg` tag.
- **Priority is composition.** The old `cf_priority` tag and the global
  `GOSTRUCTOR_PRIORITY` selector are gone. Priority is simply the order of the
  sources you pass to `WithSources` — first source that resolves a field wins.
- `Configure[T](cfg *T, opts ...Option) (*T, error)` replaces
  `ConfigureSmart`/`ConfigureSetup`/`ConfigureEasy` and their
  `(interface{}, error)` + type-assert calling convention.
- The core module has no third-party dependencies: env vars, defaults, JSON,
  and INI are handled with the standard library and a small hand-written
  parser. YAML, TOML, HOCON, and Vault support each live in their own module,
  added via `go get` only when you need them.
- INI, HOCON, and TOML are hand-written parsers for a practical subset of
  each format (see [limitations.md](limitations.md)); YAML stays on
  [`goccy/go-yaml`](https://github.com/goccy/go-yaml).
- Logging goes through `log/slog` via `WithLogger`; nothing is logged
  unless you pass one.
- `Source` is a two-method interface (`Name`, `Resolve`) any package can
  implement and register via `WithSources` — see `gostructor/yaml`'s
  `source.go` for a short example.

## Migrating from v0.x

The whole tag family (`cf_env`, `cf_default`, `cf_json`, `cf_secret`,
`cf_priority`, ...) is replaced by two tags, `cfg` and `gos`. Mechanical
mapping:

| Old | New |
|---|---|
| `cf_env:"DB_PORT"` | `cfg:"port,env:DB_PORT"` (or just `cfg:"port"` → `PORT`) |
| `cf_json:"server.host"` | `cfg:"host,json:server.host"` |
| `cf_ini:"db#pw"` | `cfg:"pw,ini:db#pw"` |
| `cf_yaml`/`cf_toml`/`cf_hocon`/`cf_vault:"..."` | `cfg:"name,<source>:..."` |
| `cf_default:"8080"` | `gos:"default:8080"` |
| `cf_secret:""` | `gos:"secret"` |
| `cf_priority:"..."` + `GOSTRUCTOR_PRIORITY` | source order in `WithSources(...)` |

Other changes:

- `ConfigureSmart(cfg)` / `myStruct.(*Config)` → `cfg, err := gostructor.Configure(&Config{})`,
  no type assertion needed.
- `ConfigureSetup(cfg, prefix, []infra.FuncType{...})` →
  `gostructor.Configure(&Config{}, gostructor.WithSources(...))`.
- YAML, TOML, HOCON, and Vault now require importing their own module (see
  [Install](../README.md#install)) and passing their source explicitly via
  `WithSources` — they're no longer bundled into the core import.
- The `Source` interface method `Tag()` is now `Name()`; error fields
  `SourceError.Tag`/`NotResolvedError.Tags` are now `.Source`/`.Sources`
  (they name sources, not struct tags).
- Nested file paths are explicit dotted overrides from the document root
  (`json:server.host`); the old implicit struct-name-based prefixing is gone.
- `ChangeLogLevel`/`ChangeLogFormatter` (global `logrus` config) →
  `gostructor.WithLogger(*slog.Logger)`, passed per call.
- **Numeric conversions are now strict.** Where v0.x silently truncated a
  fractional number into an integer field or let an out-of-range value wrap
  around, v1.0 returns a `*ConvertError`. Configs that relied on `3.9` landing
  in an `int` as `3` will now fail loudly; make the field a float, or round
  the value in a `WithHook`, to keep the old behavior explicitly.
