# gostructor documentation

The [root README](../README.md) is the quick tour. This directory holds the
deeper reference — read a page when you actually need it.

## Learn by example

The [`features/`](features/) pages are the guided tour — one capability per
page, each with a **runnable command and its real output**. Start here if
you're new. Every page maps to a self-contained program under
[`examples/`](../examples/) you can `go run` directly.

| # | Feature | Run it |
|---|---|---|
| 1 | [Two-tag configuration](features/01-two-tags.md) | `go run ./examples/basic` |
| 2 | [Priority = source order](features/02-priority.md) | `go run ./examples/priority` |
| 3 | [Sources](features/03-sources.md) | `go run ./examples/filesources` |
| 4 | [Field types](features/04-field-types.md) | `go run ./examples/types` |
| 5 | [Hooks](features/05-hooks.md) | `go run ./examples/hooks` |
| 6 | [Error taxonomy](features/06-errors.md) | `go run ./examples/errors` |
| 7 | [Observability & masking](features/07-observability.md) | `go run ./examples/observability` |
| 8 | [Full service config](features/08-webservice.md) | `go run ./examples/webservice` |

## Reference

- [configuration.md](configuration.md) — the `Configure` API, `WithSources`,
  the full `cfg`/`gos` tag grammar, hooks, and logging.
- [live-reload.md](live-reload.md) — `Watch`, the `Watchable` interface,
  `WithDebounce`/`WithValidate`, the durable snapshot store, and the live
  sources (file, git, Consul, etcd, Spring Cloud Config, Vault).
- [sources.md](sources.md) — every source in detail (env, default, JSON, INI,
  YAML, TOML, HOCON, Vault), priority, and writing your own `Source`.
- [field-types.md](field-types.md) — every supported field type and the strict
  conversion rules.
- [codegen.md](codegen.md) — `gostructor-gen`, the reflection-free `Fill` fast
  path, the three engine modes, and how equivalence is guaranteed.
- [observability.md](observability.md) — the resolution trace
  (`ConfigureWithReport`) and secret masking.
- [limitations.md](limitations.md) — known limitations and the error taxonomy.
- [migration.md](migration.md) — what changed in v1.0 and how to migrate from
  v0.x.
