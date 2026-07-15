# Observability: the resolution trace

A config layer is a black box exactly when you most need to trust it ("why is
`Port` 8080 and not what's in my file?"). `ConfigureWithReport` fills the struct
*and* returns a structured, per-field account of resolution: every source tried,
in the order tried, which one won, and the raw and converted values.

`report.String()` renders a **focused** trace — not every field, only what's
actionable: the primary source, the count of defaults, and the fields that were
overridden away from the primary config or that are secret:

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
`Attempts` and an `IsSecret` flag); `report.String()` renders the focused
summary above. The plain `Configure` builds no report, so there's zero overhead
when you don't ask for one. `WithTrace()` logs the same report through your
`WithLogger` without changing call sites.

## Masking secrets

Flag a sensitive field with `gos:"secret"` and its value is masked everywhere
it would otherwise print — the report, trace logs, and `*ConvertError`
messages — while the real value still lands on your struct:

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
