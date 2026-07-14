# gostructor examples

Runnable, zero-setup programs that show gostructor from a first `Configure`
call up to a full production configuration. Every example writes any files it
needs to a temp dir and cleans up after itself, so you can just run it:

```sh
go run ./examples/<name>
```

Start at the top and work down — each one introduces a little more.

| Example | What it shows | Modules needed |
|---|---|---|
| [`basic`](basic) | The minimum: `cf_env` with a `cf_default` fallback. | core only |
| [`priority`](priority) | The headline feature — the **same struct resolving differently** per environment via one `cf_priority` tag and `GOSTRUCTOR_PRIORITY`. | core only |
| [`filesources`](filesources) | The two core file sources (`cf_json` + `cf_ini`) combined with env and defaults in one priority chain. | core only |
| [`types`](types) | The breadth of supported field types from a JSON source: durations, slices/arrays, `time.Time`/`net.IP`, named types, pointers, maps, and slices/maps of structs. | core only |
| [`hooks`](hooks) | `WithHook` for validation (reject out-of-range values) and transformation (normalise strings), on the typed value. | core only |
| [`errors`](errors) | A tour of the typed error taxonomy — trigger each failure and classify it with `errors.Is`/`errors.As`. | core only |
| [`observability`](observability) | `ConfigureWithReport`: a per-field resolution trace, `Provenance()`, and `cf_secret` masking. | core only |
| [`webservice`](webservice) | **The big one.** A full microservice config (~30 fields, nested sub-structs) assembled from a JSON base + env overrides + defaults, with masked secrets and a full resolution trace. | core only |
| [`multisource`](multisource) | A field with several sources including a real YAML file, plus a `cf_priority` override and a validation hook. | `gostructor/yaml` |

## Suggested reading order

1. **`basic`** — get the shape of `Configure`.
2. **`priority`** — the one feature that sets gostructor apart from a plain
   unmarshaller: per-field source order, chosen at runtime.
3. **`filesources`** / **`types`** — real sources and the full type range.
4. **`hooks`** / **`errors`** — validation, transformation, and failure modes.
5. **`observability`** — see *why* each field got the value it did.
6. **`webservice`** — everything at once, at production scale.
7. **`multisource`** — bringing in an external source module (YAML).

Most examples live in the root module and need no extra setup.
`multisource` is its own module (it pulls in `gostructor/yaml`) with a
`replace` pointing back at the repo, so run it from its own directory or via
`go run ./examples/multisource`.
