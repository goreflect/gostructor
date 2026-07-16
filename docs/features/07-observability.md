# Observability & masking

← [Back to features](README.md)

When a value isn't what you expected ("why is `Port` 8080 and not what's in my
file?"), `ConfigureWithReport` fills the struct and also returns a per-field
account of resolution: every source tried, in order, which one won, and the raw
and converted values.

`report.String()` renders a focused trace, not every field: the primary source,
the count of defaults, and the fields overridden away from the primary config or
marked secret.

```go
cfg, report, err := gostructor.ConfigureWithReport(&Config{},
	gostructor.WithSources(gostructor.Env(), gostructor.JSONFile("config.json")))

fmt.Println(report.String())
report.Provenance()    // map[string]string: field -> winning source name
report.PrimarySource() // the non-default source that won the most fields
```

`report.Fields` is the machine view (`[]FieldResolution`, each with its
`Attempts` and an `IsSecret` flag). The plain `Configure` builds no report, so
there's zero overhead when you don't ask for one. `WithTrace()` logs the same
report through your `WithLogger` without changing call sites.

## Masking secrets

Flag a sensitive field with `gos:"secret"` and its value is masked everywhere it
would otherwise print — the report, trace logs, and `*ConvertError` messages —
while the real value still lands on your struct. `WithMasker` customises the
mask; the example below reveals only the last four characters.

## Try it

The [`observability`](../../examples/observability) example resolves a struct
from `Env, JSONFile` where `PORT` overrides the JSON `server.port` and
`APP_API_KEY` supplies a secret, then prints the trace and provenance map.

```sh
go run ./examples/observability
```

Output:

```text
resolved config: Port=9090 Host=0.0.0.0 APIKey="super-secret-token"

resolution trace:
Configuring main.Config: 4 fields
[Primary Source] json (loaded 2 fields)

Overrides & Secrets:
  Port   int     ⇐ env (override) = 9090
  APIKey string  ⇐ env = ••••oken (secret)

provenance (field -> winning source):
  APIKey   env
  Host     json
  MaxConns json
  Port     env
```

The trace names JSON as the primary source, calls out `Port` as an env
**override**, and masks `APIKey` to `••••oken` — while the resolved struct still
holds the real `super-secret-token`. The provenance map gives the winning source
per field for programmatic use.
