# gostructor

[![CI](https://github.com/goreflect/gostructor/actions/workflows/ci.yaml/badge.svg)](https://github.com/goreflect/gostructor/actions/workflows/ci.yaml)
[![Coverage](https://img.shields.io/badge/coverage-89%25-brightgreen)](#development)
[![Go Report Card](https://goreportcard.com/badge/github.com/goreflect/gostructor)](https://goreportcard.com/report/github.com/goreflect/gostructor)
[![Go Reference](https://pkg.go.dev/badge/github.com/goreflect/gostructor.svg)](https://pkg.go.dev/github.com/goreflect/gostructor)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

<img src="assets/card.png" alt="gostructor — fill any Go struct from any source; surprises are errors" width="820"/>

**gostructor** fills the fields of a Go struct from any mix of configuration
sources (environment variables, files, HashiCorp Vault, or plain defaults),
driven by two small struct tags, in the priority order you set with the source
list.

```go
type Config struct {
    Host  string `cfg:"host" gos:"default:0.0.0.0"`
    Port  int    `cfg:"port" gos:"default:8080"`
    Debug bool   `cfg:"debug" gos:"default:false"`
}

cfg, err := gostructor.Configure(&Config{})
```

The two tags split cleanly:

- **`cfg`** — routing and naming: `cfg:"base_name,source:override,..."`. The base
  name is what each source looks up (the env source reads `HOST` from `host`),
  and a `source:override` pins a specific key for one source.
- **`gos`** — behaviour: `gos:"default:8080,secret,optional,sep:;"` is a literal
  default, a masked-secret flag, an unresolved-is-ok flag, and a slice separator.

A field can be resolvable by several sources at once; priority is the order of
the source list, not a merge of every source into one map. A configured field
that no source can resolve is an error unless you mark it `gos:"optional"`, so a
missing value fails loudly instead of shipping a silent zero.

> **v1.0** replaces the old `cf_*` tag family with just `cfg` + `gos`, makes
> priority the source order, and splits YAML/TOML/HOCON/Vault into their own
> modules so the core stays dependency-free. Upgrading from v0.x? See
> [docs/migration.md](docs/migration.md).

## Install

```sh
go get github.com/goreflect/gostructor
```

That alone gets you the env, default, JSON, and INI sources with no
dependencies beyond the Go standard library. Add whichever of these you need:

```sh
go get github.com/goreflect/gostructor/yaml    # yaml source
go get github.com/goreflect/gostructor/toml    # toml source
go get github.com/goreflect/gostructor/hocon   # hocon source
go get github.com/goreflect/gostructor/vault   # vault source
```

Live-config and remote sources (each its own module, all optional):

```sh
go get github.com/goreflect/gostructor/watch        # fsnotify file source (hot reload)
go get github.com/goreflect/gostructor/git          # git repo as config, ref = version
go get github.com/goreflect/gostructor/consul       # Consul KV
go get github.com/goreflect/gostructor/etcd         # etcd v3
go get github.com/goreflect/gostructor/springcloud  # Spring Cloud Config Server
go get github.com/goreflect/gostructor/snapshot     # durable last-known-good store
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
    Host string `cfg:"host" gos:"default:0.0.0.0"` // env HOST, else default
    Port int    `cfg:"port" gos:"default:8080"`    // env PORT, else default
}

func main() {
    cfg, err := gostructor.Configure(&Config{})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%+v\n", cfg)
}
```

`Configure` reads the `cfg`/`gos` tags, tries each field's sources in list order
(the default list is `Env, Default`, so an env var beats the default), and
returns the same pointer you passed in. A field with no gostructor tags is left
untouched. A configured field that no source can resolve is an error unless it's
`gos:"optional"`, so misconfiguration fails loudly instead of shipping a zero
value.

## Features & live demos

Each feature has a short page under [`docs/features`](docs/features/) that pairs
an explanation with a command you can run and the output it produces. Copy the
command, run it against your checkout, and you'll see the same thing.

| Feature | What you get | Run it |
|---|---|---|
| [Two-tag configuration](docs/features/01-two-tags.md) | Fill a struct from `cfg` (routing) + `gos` (behavior). | `go run ./examples/basic` |
| [Priority = source order](docs/features/02-priority.md) | The same struct resolving differently per environment, from the `WithSources` order alone. | `go run ./examples/priority` |
| [Sources](docs/features/03-sources.md) | env, default, JSON, INI (core) + YAML/TOML/HOCON/Vault (modules). | `go run ./examples/filesources` |
| [Field types](docs/features/04-field-types.md) | Durations, slices, `time.Time`/`net.IP`, named types, pointers, maps, structs, with strict conversion. | `go run ./examples/types` |
| [Hooks](docs/features/05-hooks.md) | Validate and transform each resolved value before it lands. | `go run ./examples/hooks` |
| [Typed errors](docs/features/06-errors.md) | A closed set of typed errors, classified with `errors.Is`/`errors.As`. | `go run ./examples/errors` |
| [Observability & masking](docs/features/07-observability.md) | A focused resolution trace, provenance map, masked secrets. | `go run ./examples/observability` |
| [Full service config](docs/features/08-webservice.md) | ~30 fields, nested sub-structs, JSON + env + defaults at once. | `go run ./examples/webservice` |
| [Live reload](docs/live-reload.md) | `Watch` re-fills the struct on change; transactional last-known-good; git/consul/etcd/springcloud/vault adapters. | `cd examples/hotreload-file && go run .` |

## Supported sources

The source name is what you target in a per-source override
(`cfg:"port,env:DB_PORT"`) and what appears in the resolution trace. Non-core
sources are opt-in via `WithSources`, so a bare `cfg` name never triggers an
unexpected file load.

| Source | Reads from | Module | Requires |
|---|---|---|---|
| `default` | Literal `gos:"default:..."` value | core | — |
| `env` | Environment variable | core | — |
| `json` | JSON file | core | `GOSTRUCTOR_JSON=...` |
| `ini` | INI file | core | `GOSTRUCTOR_INI=...` |
| `keyvalue` | Key/value file (`.env`/`.properties`) | core | `GOSTRUCTOR_KEYVALUE=...` |
| `yaml` | YAML file | `gostructor/yaml` | `GOSTRUCTOR_YAML=...` |
| `toml` | TOML file | `gostructor/toml` | `GOSTRUCTOR_TOML=...` |
| `hocon` | HOCON file | `gostructor/hocon` | `GOSTRUCTOR_HOCON=...` |
| `vault` | HashiCorp Vault secret | `gostructor/vault` | `VAULT_ADDR`, `VAULT_TOKEN` |
| `map` | In-memory map (`gostructor.Map`) | core | — |

**Live / remote sources.** Each implements `Watchable`, so `gostructor.Watch`
re-fills the struct when the backing data changes (see
[docs/live-reload.md](docs/live-reload.md)):

| Source | Reads from | Module | Live via |
|---|---|---|---|
| `file` | Config file on disk (any format) | `gostructor/watch` | fsnotify |
| `git` | File in a git repo, any format (ref = version) | `gostructor/git` | poll for drift + `SetVersion` |
| `consul` | Consul KV prefix | `gostructor/consul` | blocking queries |
| `etcd` | etcd v3 key prefix | `gostructor/etcd` | native watch API |
| `springcloud` | Spring Cloud Config Server | `gostructor/springcloud` | poll |
| `vault` | HashiCorp Vault secret | `gostructor/vault` | poll (secret rotation) |

Each has a docker-compose example under [`examples/`](examples/). Full addressing
rules, per-source key derivation, and priority are in
[docs/sources.md](docs/sources.md).

The `file` and `git` sources are format-agnostic: they default to JSON but read
any format the library supports via a pluggable decoder. Pass
`gostructor.DecodeINI` / `gostructor.DecodeKeyValue` (zero-dep) or `yaml.Decode`
/ `toml.Decode` / `hocon.Decode` as the source's `Decoder`.

## Documentation

The [`docs/`](docs/) directory holds the deeper reference:

- **[configuration.md](docs/configuration.md)** — the `Configure` API,
  `WithSources`, the full `cfg`/`gos` tag grammar, hooks, and logging.
- **[sources.md](docs/sources.md)** — every source in detail, priority, and
  writing your own `Source`.
- **[field-types.md](docs/field-types.md)** — supported field types and strict
  conversion rules.
- **[codegen.md](docs/codegen.md)** — `gostructor-gen`, the reflection-free
  `Fill` fast path, and the engine modes.
- **[observability.md](docs/observability.md)** — the resolution trace
  (`ConfigureWithReport`) and secret masking.
- **[limitations.md](docs/limitations.md)** — known limitations and the typed
  errors.
- **[migration.md](docs/migration.md)** — v1.0 changes and migrating from v0.x.

## Roadmap

Shipped: the resolution trace and secret masking, hot reload (`Watch`), git as a
config source, the Consul/etcd/Spring Cloud/Vault adapters, and code generation
(`gostructor-gen`, a reflection-free `Fill`). Still ahead: self-documenting
config, custom time layouts, config-file template generation, richer env
handling, CLI flags from the struct, an `fx` integration, and migration tooling.
See [ROADMAP.md](ROADMAP.md) for the full list.

## Development

```sh
make build   # go build ./... in every module (core, yaml, toml, hocon, vault, examples/multisource)
make vet
make test    # go test ./... -race -cover in every module
make all     # build + vet + test
```

Test coverage (statements, library packages, excluding the runnable
`examples/`): **~89%** aggregate. Per module: core `86%`, yaml `79%`, toml
`85%`, hocon `84%`, vault `87%`, with the internal `convert`/`ini`/`structplan`
parsers each above `89%`. Reproduce with:

```sh
go test -coverpkg=$(go list ./... | grep -v /examples/ | paste -sd, -) \
    $(go list ./... | grep -v /examples/) -cover
```

Each submodule's `go.mod` carries a local `replace` directive pointing at `../`
so it builds against your working copy of core; that line is a no-op for anyone
importing the module normally, since Go only applies `replace` directives from
the main module of a build.

Runnable usage examples live in [`examples/`](examples/).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
