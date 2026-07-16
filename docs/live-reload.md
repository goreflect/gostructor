# Live reload / hot configuration

`Configure` fills a struct once. For a long-lived service that should pick up
changed configuration without a restart, gostructor adds a small, opt-in layer
on top of the same resolution path:

- a `Watchable` interface a source may implement,
- a `Watch[T]` driver that re-fills the struct transactionally, and
- `WithDebounce` / `WithValidate` options that gate reloads.

The same `cfg`/`gos` tags, source order, hooks, trace, and secret masking apply
to a live reload just as they do to the first fill.

## The `Watchable` interface

A source signals that its backing data can change by implementing:

```go
type Watchable interface {
    Watch(ctx context.Context, onChange func()) error
}
```

`Watch` blocks until `ctx` is cancelled, calling `onChange` whenever the data
changes. A source that caches a snapshot for `Resolve` must refresh that snapshot
before calling `onChange`, so the re-fill reads the new data. The core ships
`PollWatch`, a helper that implements `Watch` by polling a cheap fingerprint (a
commit SHA, a KV index, a secret version) on an interval. The poll-based adapters
use it.

## The `Watch` driver

```go
func Watch[T any](
    ctx context.Context,
    target *T,
    onReload func(*T, error),
    opts ...Option,
) error
```

`Watch` fills `target` once, then re-fills a fresh copy of `T` whenever any
configured `Watchable` source reports a change, delivering each result to
`onReload`. It blocks until `ctx` is cancelled.

Reloads are transactional, so the app always runs on a complete config:

1. Resolve into a fresh `*T`. The live struct is never mutated in place, so a
   half-applied fill is never observable.
2. Run `WithValidate` (whole-struct) against the fresh copy.
3. Only on success publish it via `onReload(fresh, nil)`. On failure the previous
   config keeps serving and the error goes to `onReload(nil, err)` (and to the
   log if `WithLogger` is set). A bad change is logged, not fatal.

`onReload` is called once with the initial fill (so you can publish it with the
same atomic-swap code you use for reloads) and again for every reload attempt.
If the initial fill fails, `Watch` returns that error immediately without calling
`onReload`, so startup misconfiguration stays fatal like it is with `Configure`.

### Typical usage

```go
var current atomic.Pointer[Config]

go func() {
    err := gostructor.Watch(ctx, &Config{}, func(cfg *Config, err error) {
        if err != nil {
            log.Printf("config reload rejected, keeping last good: %v", err)
            return
        }
        current.Store(cfg) // atomic swap; request handlers read current.Load()
    },
        gostructor.WithSources(src, gostructor.Default()),
        gostructor.WithDebounce(200*time.Millisecond),
        gostructor.WithValidate(func(c *Config) error {
            if c.Port < 1 || c.Port > 65535 {
                return fmt.Errorf("port %d out of range", c.Port)
            }
            return nil
        }),
    )
    if err != nil && ctx.Err() == nil {
        log.Fatal(err)
    }
}()
```

## Options

- `WithDebounce(d)` — coalesce a burst of change signals into one reload. A
  single logical change often arrives as several low-level events (an editor
  writes a file in multiple syscalls; a config server pushes a batch of key
  events). The debounce window collapses a burst that's quiet for `d` into one
  re-fill. Zero disables it.
- `WithValidate(func(*T) error)` — a whole-struct check run after a reload fills
  a fresh copy, before it's published. Returning an error rejects that reload.
  Complements per-field `WithHook` by seeing the whole struct at once.

Both affect `Watch` only; a plain `Configure` ignores them.

## Sources that support live reload

| Module | Source | Live via |
|---|---|---|
| `gostructor/watch` | `file` (JSON on disk) | fsnotify (directory watch, survives atomic writes) |
| `gostructor/git` | `git` (file in a repo) | poll for a new commit on the ref; `SetVersion` to switch |
| `gostructor/consul` | `consul` (KV prefix) | Consul blocking queries |
| `gostructor/etcd` | `etcd` (key prefix) | etcd native watch API |
| `gostructor/springcloud` | `springcloud` | poll `GET /{app}/{profile}` |
| `gostructor/vault` | `vault` | poll referenced secrets (rotation) |

Each has a runnable, docker-compose-backed example under
[`examples/`](../examples/):
[`hotreload-file`](../examples/hotreload-file/) (no Docker),
[`git`](../examples/git/), [`consul`](../examples/consul/),
[`etcd`](../examples/etcd/), [`springcloud`](../examples/springcloud/),
[`vault`](../examples/vault/).

## File formats (git & file sources)

The `git` and `file` sources are format-agnostic. They default to JSON but read
any format via a pluggable `Decoder` (`func([]byte) (map[string]any, error)`):

| Format | Decoder | Dependency |
|---|---|---|
| JSON | `gostructor.DecodeJSON` (default) | none (core) |
| INI | `gostructor.DecodeINI` | none (core) |
| key/value (`.env`) | `gostructor.DecodeKeyValue` | none (core) |
| YAML | `yaml.Decode` | `gostructor/yaml` |
| TOML | `toml.Decode` | `gostructor/toml` |
| HOCON | `hocon.Decode` | `gostructor/hocon` |

```go
src, _ := git.New(git.Options{Repo: ..., Path: "config.yaml", Decoder: yaml.Decode})
src, _ := watch.New(watch.Options{Path: "app.env", Decoder: gostructor.DecodeKeyValue})
```

Flat formats (key/value, INI) resolve dotted keys by an exact top-level match,
so `cfg:"host,file:server.host"` finds either a literal `server.host` key or the
nested `server` → `host` a structured format produces.

## Durable last-known-good (`gostructor/snapshot`)

Remote sources (git, config servers) accept an optional snapshot `Store`. Every
good fetch is written to it, and if the upstream is unreachable at startup the
last snapshot is served instead of failing, so a transient outage doesn't take
the service down.

```go
store, _ := snapshot.NewDirStore("/var/lib/app/config")
src, _ := git.New(git.Options{Repo: "...", Ref: "main", Path: "config.json", Snapshot: store})
```

`Store` is a two-method interface (`Save`, `Load`); `DirStore` is the default
on-disk implementation, and you can plug in your own (object store, database).

## Writing a `Watchable` adapter

Any `Source` becomes live by adding a `Watch` method. For a backend you poll,
`PollWatch` does the loop for you:

```go
func (s *mySource) Watch(ctx context.Context, onChange func()) error {
    return gostructor.PollWatch(ctx, 30*time.Second,
        func(ctx context.Context) (string, error) {
            return s.currentVersion(ctx) // a cheap fingerprint
        },
        func() {
            s.refresh()  // update the cached snapshot first…
            onChange()   // …then signal the reload
        },
        s.logger,
    )
}
```

For a backend with a push/stream API (etcd watch, Consul blocking queries),
implement `Watch` directly: block on the stream, refresh the cached snapshot on
each event, then call `onChange`.
