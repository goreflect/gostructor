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
| [`basic`](basic) | The minimum: a `cfg` base name (env) with a `gos:"default:..."` fallback. | core only |
| [`priority`](priority) | The headline feature — the **same struct resolving differently** per environment, purely from the `WithSources` order. | core only |
| [`filesources`](filesources) | The two core file sources (json + ini) combined with env and defaults in one source-order chain. | core only |
| [`types`](types) | The breadth of supported field types from a JSON source: durations, slices/arrays, `time.Time`/`net.IP`, named types, pointers, maps, and slices/maps of structs. | core only |
| [`hooks`](hooks) | `WithHook` for validation (reject out-of-range values) and transformation (normalise strings), on the typed value. | core only |
| [`errors`](errors) | A tour of the typed error taxonomy — trigger each failure and classify it with `errors.Is`/`errors.As`. | core only |
| [`observability`](observability) | `ConfigureWithReport`: the focused resolution trace (primary source, overrides & secrets), `Provenance()`, and `gos:"secret"` masking. | core only |
| [`webservice`](webservice) | **The big one.** A full microservice config (~30 fields, nested sub-structs) assembled from a JSON base + env overrides + defaults, with masked secrets and a focused trace. | core only |
| [`multisource`](multisource) | A field with several sources including a real YAML file, plus a validation hook. Priority is the source order. | `gostructor/yaml` |

### Live config & remote sources

These show `gostructor.Watch` — the struct re-fills when the backing source
changes. All but `hotreload-file` bring their backend up with **docker-compose**
and seed it, so each is fully reproducible; see the example's own README for the
exact commands.

| Example | What it shows | Backend |
|---|---|---|
| [`hotreload-file`](hotreload-file) | `Watch` + fsnotify file source: edit a JSON file, the struct reloads. Transactional last-known-good via `WithValidate`. | none (local file) |
| [`git`](git) | Config from a git repo, ref = version; drift polling and runtime `SetVersion`; snapshot fallback. | `git daemon` (compose) |
| [`consul`](consul) | Consul KV prefix, live via blocking queries. | `hashicorp/consul` (compose) |
| [`etcd`](etcd) | etcd key prefix, live via the native watch API. | `etcd` (compose) |
| [`springcloud`](springcloud) | Spring Cloud Config Server over HTTP, live via polling. | `spring-cloud-config-server` (compose) |
| [`vault`](vault) | Vault secrets with live rotation (polling); secret masking. | `hashicorp/vault` (compose) |

## Suggested reading order

1. **`basic`** — get the shape of `Configure` and the two tags.
2. **`priority`** — the one feature that sets gostructor apart from a plain
   unmarshaller: source-order priority, chosen at the call site.
3. **`filesources`** / **`types`** — real sources and the full type range.
4. **`hooks`** / **`errors`** — validation, transformation, and failure modes.
5. **`observability`** — see *why* each field got the value it did.
6. **`webservice`** — everything at once, at production scale.
7. **`multisource`** — bringing in an external source module (YAML).

Most core examples live in the root module and need no extra setup.
`multisource` and every live-config example are their own modules (they pull in
an external source module) with `replace` directives pointing back at the repo,
so run each from its own directory: `cd examples/<name> && go run .`.
