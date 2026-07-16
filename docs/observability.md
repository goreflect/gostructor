# Observability: the resolution trace

When a value isn't what you expected ("why is `Port` 8080 and not what's in my
file?"), you want to see how it was resolved. `ConfigureWithReport` fills the
struct and also returns a per-field account: every source tried, in order, which
one won, and the raw and converted values.

`report.String()` renders a focused trace, not every field. It shows the primary
source, the count of defaults, and the fields that were overridden away from the
primary config or that are secret:

```go
cfg, report, err := gostructor.ConfigureWithReport(&Config{},
    gostructor.WithSources(gostructor.Env(), gostructor.JSONFile("config.json")))
if err != nil {
    log.Fatal(err)
}
fmt.Println(report.String())
// Configuring main.Config: 4 fields
// [Primary Source] json (loaded 2 fields)
//
// Overrides & Secrets:
//   Port   int     ⇐ env (override) = 9090
//   APIKey string  ⇐ env = ••••oken (secret)

report.Provenance()    // map[string]string: field -> winning source name
report.PrimarySource() // "json": the non-default source that won the most fields
```

`report.Fields` is the machine view (`[]FieldResolution`, each with its
`Attempts` and an `IsSecret` flag); `report.String()` renders the summary above.
Plain `Configure` builds no report, so there's no overhead when you don't ask for
one. `WithTrace()` logs the same report through your `WithLogger` without
changing call sites.

## Masking secrets

Flag a sensitive field with `gos:"secret"` and its value is masked everywhere it
would otherwise print (the report, trace logs, and `*ConvertError` messages),
while the real value still lands on your struct:

```go
type Config struct {
    APIKey string `cfg:"apiKey,env:APP_API_KEY" gos:"secret"`
}

// Default fully redacts ("••••••"). Override to reveal, e.g., the last 4 chars:
gostructor.WithMasker(func(_ gostructor.FieldContext, v any) string {
    s, _ := v.(string)
    if len(s) >= 4 {
        return "••••" + s[len(s)-4:]
    }
    return "••••"
})
```

See `examples/observability` for a runnable end-to-end demo.

## Inspecting a running service: the debug dump

`ConfigureWithReport` answers "how was this resolved?" at the call site. The debug
dump answers the same question for an already-running process — you want to see
how a deployed service actually resolved its config, without redeploying. Opt in
with `WithDebugDump`, choosing where the dump goes:

```go
gostructor.Configure(&cfg,
    gostructor.WithSources(gostructor.Env(), gostructor.JSONFile("config.json")),
    // A loopback TCP port: `nc 127.0.0.1 6555` or the gostructor-dump client
    // prints the live config. No HTTP involved.
    gostructor.WithDebugDump(gostructor.DumpTCP("127.0.0.1:6555")),
    // ...or a file to cat, handy under `kubectl exec`:
    // gostructor.WithDebugDump(gostructor.DumpFile("/tmp/gostructor.dump")),
)
```

Both sinks write the full per-field state after every successful fill — every
field, its type, the source that won it (or its outcome if none did), and its
value, **with `gos:"secret"` fields masked** exactly as in the report:

```
gostructor config dump — main.Config (4 fields)

FIELD   TYPE    SOURCE        VALUE
Host    string  json          db
Name    string  (unresolved)  —
Port    int     env           9090
Secret  string  vault         ••••••
```

Under `Watch`, the dump refreshes on every live reload, so it always shows the
config the service is currently serving; a failed reload leaves the
last-known-good dump in place. The `DumpTCP` listener is closed when `Watch`'s
context is cancelled; a plain `Configure` leaves it serving for the life of the
process.

The dump never affects resolution and never fails `Configure`: a busy port or an
unwritable file is logged (via `WithLogger`) and skipped. `DumpTCP` binds
loopback by default, so a pod's dump port is reachable via `kubectl exec` /
`kubectl port-forward` but not exposed on the pod network — pass an explicit host
(`DumpTCP("0.0.0.0:6555")`) to widen it.

### Turning it on without a recompile

The dump is off unless asked for. When you pass no `WithDebugDump`, the
`GOSTRUCTOR_DEBUG_DUMP` environment variable is consulted instead, so you can
enable it on a running deployment by setting an env var and restarting:

| `GOSTRUCTOR_DEBUG_DUMP` | Effect                                        |
| ----------------------- | --------------------------------------------- |
| unset / `off` / `0`     | disabled (the default)                        |
| `on` / `tcp`            | `DumpTCP` on `127.0.0.1:6555`                  |
| `host:port`             | `DumpTCP` on that address                     |
| `file:/path/to/file`    | `DumpFile` at that path                        |

An explicit `WithDebugDump` in code always wins over the environment variable.

### The gostructor-dump client

`cmd/gostructor-dump` is a tiny client for the `DumpTCP` endpoint: it connects,
prints the dump, and exits — like `nc`, with a sane default address.

```
go run github.com/goreflect/gostructor/cmd/gostructor-dump           # 127.0.0.1:6555
go run github.com/goreflect/gostructor/cmd/gostructor-dump host:6555
```
