# gostructor

[![Actions Status](https://github.com/goreflect/gostructor/workflows/CI_dev/badge.svg)](https://github.com/goreflect/gostructor/actions?query=workflow%3ACI_dev)
[![Go Report Card](https://goreportcard.com/badge/github.com/goreflect/gostructor)](https://goreportcard.com/report/github.com/goreflect/gostructor)
[![Go Reference](https://pkg.go.dev/badge/github.com/goreflect/gostructor.svg)](https://pkg.go.dev/github.com/goreflect/gostructor)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

<img src="logo.svg" alt="gostructor" width="640"/>

**gostructor** fills the fields of a Go struct from any mix of configuration
sources — environment variables, files, HashiCorp Vault, or plain
struct-tag defaults — driven entirely by struct tags, in the priority order
you choose per field.

```go
type Config struct {
    Host  string `cf_env:"APP_HOST" cf_default:"0.0.0.0"`
    Port  int    `cf_env:"APP_PORT" cf_default:"8080"`
    Debug bool   `cf_env:"APP_DEBUG" cf_default:"false"`
}

cfg, err := gostructor.Configure(&Config{})
```

That one field having both `cf_env` and `cf_default` — tried in that order,
first one that resolves wins — is what gostructor actually gives you beyond
a plain unmarshaller: **one field, several possible sources, resolved with a
priority you control**, rather than merging every source into one map and
unmarshalling it once.

## Why v1.0 looks different from earlier versions

This release is a deliberate break from the pre-1.0 API, aligned with how
Go config libraries are actually built today:

- **A generics-based API.** `Configure[T](cfg *T, opts ...Option) (*T, error)`
  replaces `ConfigureSmart`/`ConfigureSetup`/`ConfigureEasy` and the
  `(interface{}, error)` + type-assert dance every caller used to write.
- **The core module has zero third-party dependencies.** Environment
  variables, `cf_default`, JSON, and INI are handled with the standard
  library and a small hand-written parser. YAML, TOML, HOCON, and Vault
  support live in **separate modules** you opt into individually — `go get`
  only pulls in what you actually use.
- **Hand-written INI and HOCON-subset and TOML-subset parsers**, replacing
  dependencies on a years-stale `go-ini` fork and an unstable, single-author
  HOCON library. YAML deliberately stays on the mature
  [`goccy/go-yaml`](https://github.com/goccy/go-yaml) rather than
  reimplementing that spec by hand — see [Known limitations](#known-limitations)
  for exactly why.
- **`log/slog`, not `logrus`.** Pass your own `*slog.Logger` via
  `WithLogger`; nothing is written anywhere unless you ask for it.
- **A real extension point.** `Source` is a two-method interface any package
  can implement (see how `gostructor/yaml`, `gostructor/vault`, etc. do it),
  registered via `WithSources` — not a stub that silently did nothing.

If you're upgrading from a pre-1.0 version, see
[Migrating from v0.x](#migrating-from-v0x).

## Table of contents

- [Install](#install)
- [Quick start](#quick-start)
- [Supported sources](#supported-sources)
- [Supported field types](#supported-field-types)
- [The `Configure` API](#the-configure-api)
- [Sources in detail](#sources-in-detail)
  - [Defaults](#defaults-cf_default)
  - [Environment variables](#environment-variables-cf_env)
  - [JSON and INI (core)](#json-and-ini-core)
  - [YAML, TOML, HOCON (optional modules)](#yaml-toml-hocon-optional-modules)
  - [HashiCorp Vault (optional module)](#hashicorp-vault-optional-module)
  - [Combining multiple sources with priority](#combining-multiple-sources-with-priority)
- [Hooks: validation and transformation](#hooks-validation-and-transformation)
- [Logging](#logging)
- [Writing your own Source](#writing-your-own-source)
- [Known limitations](#known-limitations)
- [Migrating from v0.x](#migrating-from-v0x)
- [Roadmap](#roadmap)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Install

```sh
go get github.com/goreflect/gostructor
```

That alone gets you `cf_env`, `cf_default`, `cf_json`, and `cf_ini` with no
dependencies beyond the Go standard library. Add whichever of these you
need:

```sh
go get github.com/goreflect/gostructor/yaml    # cf_yaml
go get github.com/goreflect/gostructor/toml    # cf_toml
go get github.com/goreflect/gostructor/hocon   # cf_hocon
go get github.com/goreflect/gostructor/vault   # cf_vault
```

Requires Go 1.24+.

## Quick start

```go
package main

import (
    "fmt"
    "log"

    "github.com/goreflect/gostructor"
)

type Config struct {
    Host string `cf_env:"APP_HOST" cf_default:"0.0.0.0"`
    Port int    `cf_env:"APP_PORT" cf_default:"8080"`
}

func main() {
    cfg, err := gostructor.Configure(&Config{})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%+v\n", cfg)
}
```

`Configure` reads the struct tags present on `Config`, tries each field's
sources in order (`cf_env` before `cf_default` here), and returns the same
pointer you passed in — so `cfg` is both the argument and the result. A
field with no gostructor tags at all is left untouched; a field that *does*
carry a tag but that no configured source can resolve is a hard error, so
misconfiguration fails loudly instead of shipping a zero value.

## Supported sources

| Tag | Source | Module | Requires |
|---|---|---|---|
| `cf_default` | Inline default value on the tag itself | core | — |
| `cf_env` | Environment variable | core | — |
| `cf_json` | JSON file | core | `GOSTRUCTOR_JSON=path/to/file.json` |
| `cf_ini` | INI file | core | `GOSTRUCTOR_INI=path/to/file.ini` |
| `cf_yaml` | YAML file | `gostructor/yaml` | `GOSTRUCTOR_YAML=path/to/file.yml` |
| `cf_toml` | TOML file | `gostructor/toml` | `GOSTRUCTOR_TOML=path/to/file.toml` |
| `cf_hocon` | HOCON file | `gostructor/hocon` | `GOSTRUCTOR_HOCON=path/to/file.hocon` |
| `cf_vault` | HashiCorp Vault secret | `gostructor/vault` | `VAULT_ADDR`, `VAULT_TOKEN` |
| `cf_priority` | Per-field source ordering when a field carries several of the tags above | core | — |

Non-core sources must be added explicitly via `WithSources` — the core
module has no way to know they exist (that's the whole point of the
zero-dependency core).

## Supported field types

- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `string`
- `bool`
- slices of any of the above, e.g. `[]int32`, `[]string`, `[]bool`
- `map[string|int]string|int|float32|float64|bool`, when the source is
  structured data (JSON, YAML, TOML, HOCON) and the tag addresses a whole
  nested object rather than one leaf value. `cf_env`, `cf_default`, `cf_ini`,
  and `cf_vault` encode values as a flat string and so can only populate
  slices, not maps.

## The `Configure` API

```go
func Configure[T any](target *T, opts ...Option) (*T, error)
```

With no options, `Configure` uses the core module's built-in sources — Env,
JSON, then Default, in that order — auto-selected per field by which tags
are actually present. Bring in other sources with `WithSources`, which
replaces the default list with an explicit, ordered one:

```go
import (
    "github.com/goreflect/gostructor"
    "github.com/goreflect/gostructor/yaml"
)

cfg, err := gostructor.Configure(&Config{}, gostructor.WithSources(
    gostructor.Env(),
    yaml.New(),
    gostructor.Default(),
))
```

Sources are tried in the order given; the first one that reports a value
for a field wins (see [priority](#combining-multiple-sources-with-priority)
for per-field overrides). `Env()`, `Default()`, `JSON()`, and `INI()` are
built into the core module; each also has a `*File(path string)` variant
(`JSONFile`, `INIFile`, `yaml.File`, ...) to read from an explicit path
instead of the source's environment variable.

## Sources in detail

### Defaults (`cf_default`)

```go
type Config struct {
    Retries int    `cf_default:"3"`
    Flags   []bool `cf_default:"true,false,true"`
}
```

The tag value is the literal default. Slices are comma-separated; there is
no map syntax for defaults (see [Supported field types](#supported-field-types)).

### Environment variables (`cf_env`)

```go
type Config struct {
    Port    int    `cf_env:"APP_PORT"`
    Signals []bool `cf_env:"MY_SIGNALS"` // e.g. MY_SIGNALS=true,false,true
}
```

### JSON and INI (core)

```go
type Config struct {
    Host string `cf_json:"server.host"`
    Tags []int  `cf_json:"server.tags"`
}
```

```go
os.Setenv(gostructor.JSONFileEnvVar, "config.json")
cfg, err := gostructor.Configure(&Config{})
```

```json
{"server": {"host": "0.0.0.0", "tags": [1, 2, 3]}}
```

A dotted tag value (`server.host`) descends into nested JSON objects without
requiring a matching nested Go struct; `cf_json:"server"` on a
`map[string]T` field would address the whole nested object instead.

INI uses `section#key` instead of dotted paths, since INI's own structure is
two levels (section, then key), matching how `cf_toml`/`cf_vault` address
their sources too:

```go
type Config struct {
    Password string `cf_ini:"database#password"`
}
```

```ini
[database]
password = secret
```

### YAML, TOML, HOCON (optional modules)

Each lives in its own module and mirrors JSON's addressing style (YAML) or
INI's (TOML/HOCON), and is registered explicitly via `WithSources`:

```go
import "github.com/goreflect/gostructor/yaml"

type Config struct {
    Host string   `cf_yaml:"server.host"`
    Tags []string `cf_yaml:"server.tags"`
}

os.Setenv(yaml.FileEnvVar, "config.yml")
cfg, err := gostructor.Configure(&Config{}, gostructor.WithSources(yaml.New()))
```

```yaml
server:
  host: 0.0.0.0
  tags:
    - primary
    - eu-west
```

`gostructor/toml` and `gostructor/hocon` ship **hand-written parsers for a
practical subset** of each format, not the full spec — see
[Known limitations](#known-limitations) for exactly what's out of scope
before you commit a config file that needs it.

### HashiCorp Vault (optional module)

Set `VAULT_ADDR` and `VAULT_TOKEN` (the same variables the `vault` CLI
itself uses), then tag fields as `path/to/secret#key`:

```go
import "github.com/goreflect/gostructor/vault"

type Config struct {
    APIKey    string  `cf_vault:"my-service/stage/creds#api-key"`
    RateLimit int16   `cf_vault:"my-service/stage/limits#rate"`
    Allowlist []int32 `cf_vault:"my-service/stage/net#allowlist"` // comma-separated secret value
}

cfg, err := gostructor.Configure(&Config{}, gostructor.WithSources(vault.New()))
```

Built on the official [`hashicorp/vault/api`](https://github.com/hashicorp/vault)
client, not a third-party wrapper.

### Combining multiple sources with priority

A field can carry several source tags at once; by default gostructor tries
whatever's in your source list, in that order. To override the order for
one specific field, add `cf_priority` and set `GOSTRUCTOR_PRIORITY` to the
name of the stage you want:

```go
type Config struct {
    Value string `cf_env:"MY_VALUE" cf_default:"fallback" cf_priority:"prod:cf_env,cf_default;dev:cf_default,cf_env"`
}
```

```go
os.Setenv(gostructor.PriorityEnvVar, "prod") // this field tries cf_env first, falls back to cf_default
```

## Hooks: validation and transformation

`WithHook` runs after a value is resolved but before it's set on the
struct — return an error to reject it, or a different value to transform
it:

```go
cfg, err := gostructor.Configure(&Config{}, gostructor.WithHook(
    func(field gostructor.FieldContext, value any) (any, error) {
        if field.Name == "Port" && value.(int) < 1024 {
            return nil, fmt.Errorf("port %v is a privileged port", value)
        }
        return value, nil
    },
))
```

## Logging

```go
logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
cfg, err := gostructor.Configure(&Config{}, gostructor.WithLogger(logger))
```

With no `WithLogger`, `Configure` logs nothing.

## Writing your own Source

```go
type Source interface {
    Tag() string
    Resolve(field FieldContext) (value any, found bool, err error)
}
```

`found=false` means "nothing to contribute for this field" (for example,
the field doesn't carry your tag), letting `Configure` fall through to the
next source instead of treating it as an error. Look at
`gostructor/yaml`'s `source.go` for a complete, short example — it's about
seventy lines including file loading and error handling.

## Known limitations

- **`gostructor/hocon` and `gostructor/toml` parse a practical subset, not
  the full spec.** Both are hand-written and intentionally bounded in
  scope:
  - HOCON: no `${}` substitutions, no `include`, no duration/size unit
    literals (`10m`, `5 MB`), no string concatenation across values.
  - TOML: no datetimes, no inline tables (`{ a = 1 }`), no arrays of
    tables (`[[table]]`), no multiline/triple-quoted strings.

  If your config file needs any of those, gostructor will return a parse
  error rather than silently misinterpreting it — but you'll want a
  different tool for that file. YAML deliberately did **not** get the same
  hand-written treatment: full YAML (block/flow styles, anchors and
  aliases, implicit typing) is large enough that reimplementing it
  correctly is a multi-year undertaking even for dedicated parser
  projects, so `gostructor/yaml` stays on `goccy/go-yaml`.
- **Error values are plain `errors.New`/`fmt.Errorf`, not sentinel
  errors.** You can `errors.Is`/`errors.As` through the `%w`-wrapped chain
  down to the underlying cause, but there's no `ErrNotFound`-style sentinel
  to match on yet.
- **Struct fields must be exported.** Unlike some earlier versions of this
  library, there's no `unsafe`-pointer trick to write into unexported
  fields — this matches `encoding/json`'s own convention and keeps the
  engine free of `unsafe`.

## Migrating from v0.x

The tag names (`cf_env`, `cf_default`, `cf_json`, ...) are unchanged. What's
different:

- `ConfigureSmart(cfg)` / `myStruct.(*Config)` → `cfg, err := gostructor.Configure(&Config{})`,
  no type assertion needed.
- `ConfigureSetup(cfg, prefix, []infra.FuncType{...})` →
  `gostructor.Configure(&Config{}, gostructor.WithSources(...))`.
- YAML, TOML, HOCON, and Vault now require importing their own module (see
  [Install](#install)) and passing their source explicitly via
  `WithSources` — they're no longer bundled into the core import.
- `cf_hocon`/`cf_yaml`/`cf_json` tag values are now always explicit dotted
  paths from the document root (`server.host`); the old implicit
  struct-name-based prefixing is gone.
- `ChangeLogLevel`/`ChangeLogFormatter` (global `logrus` config) →
  `gostructor.WithLogger(*slog.Logger)`, passed per call.

## Roadmap

`cf_server_file` and `cf_server_kv` (a remote config-server and a
key/value-store backend, Spring Cloud Config Server style) were on earlier
versions' roadmap and remain unimplemented. They're intentionally left out
of core v1.0 scope — the plan is to ship them the same way YAML/TOML/HOCON/
Vault work today: as separate, explicitly-opted-into modules implementing
`Source`, once there's a concrete design for the network/auth/retry
surface a remote source needs. Not yet scheduled:

- A key/value store backend module, with change callbacks
- A config-server-fetching module
- Sentinel/wrapped error types for `errors.Is`/`errors.As`
- Live-reloading a struct's values when its backing source changes

## Development

Each module (core, `yaml`, `toml`, `hocon`, `vault`) is verified
independently:

```sh
go build ./... && go vet ./... && go test ./... -cover        # from the repo root (core)
cd yaml  && go build ./... && go vet ./... && go test ./... -cover
cd toml  && go build ./... && go vet ./... && go test ./... -cover
cd hocon && go build ./... && go vet ./... && go test ./... -cover
cd vault && go build ./... && go vet ./... && go test ./... -cover
```

Each submodule's `go.mod` carries a local `replace` directive pointing at
`../` so it builds against your working copy of core instead of a published
release; that line is a no-op for anyone importing the module normally,
since Go only applies `replace` directives from the main module of a build.

## Contributing

Issues and pull requests are welcome. If you're picking up one of the
roadmap items above or fixing a bug, a short description of the approach in
the PR body is appreciated so reviewers don't have to reverse-engineer
intent from the diff.

## License

[MIT](LICENSE)
