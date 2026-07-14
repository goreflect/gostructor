# gostructor features

A tour of what gostructor does, one capability per page. Every page pairs a
short explanation with a **runnable command and its real output** — so you can
copy the command, run it against your own checkout, and see the same thing.

Each feature maps to a self-contained program under
[`examples/`](../../examples/). The examples write any files they need to a temp
dir and clean up after themselves, so there's no setup:

```sh
go run ./examples/<name>
```

## Feature pages

| # | Feature | What you get | Example | Modules |
|---|---|---|---|---|
| 1 | [Two-tag configuration](01-two-tags.md) | Fill a struct from `cfg` (routing) + `gos` (behavior) tags. | `basic` | core |
| 2 | [Priority = source order](02-priority.md) | The same struct resolving differently per environment, from the `WithSources` order alone. | `priority` | core |
| 3 | [Sources](03-sources.md) | env, default, JSON, INI (core) + YAML, TOML, HOCON, Vault (modules). | `filesources`, `multisource` | core / `yaml` |
| 4 | [Field types](04-field-types.md) | Durations, slices/arrays, `time.Time`/`net.IP`, named types, pointers, maps, structs — with strict conversion. | `types` | core |
| 5 | [Hooks](05-hooks.md) | Validate and transform each resolved value before it lands. | `hooks` | core |
| 6 | [Error taxonomy](06-errors.md) | A small closed set of typed errors you classify with `errors.Is`/`errors.As`. | `errors` | core |
| 7 | [Observability & masking](07-observability.md) | A focused resolution trace, provenance map, and masked secrets. | `observability` | core |
| 8 | [Full service config](08-webservice.md) | ~30 fields, nested sub-structs, JSON + env + defaults, all at once. | `webservice` | core |

## Suggested reading order

1. **Two-tag configuration** — the shape of `Configure` and the two tags.
2. **Priority** — the one feature that sets gostructor apart from a plain unmarshaller.
3. **Sources** / **Field types** — real backends and the full type range.
4. **Hooks** / **Error taxonomy** — validation, transformation, failure modes.
5. **Observability** — see *why* each field got the value it did.
6. **Full service config** — everything at once, at production scale.

For the full API reference, the `cfg`/`gos` tag grammar, and migration notes,
see the [root README](../../README.md).
