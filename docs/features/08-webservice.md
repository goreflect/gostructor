# Full service config

← [Back to features](README.md)

The big one: everything at once, at production scale. The
[`webservice`](../../examples/webservice) example assembles a realistic
microservice config — ~30 fields across nested sub-structs (server, database,
redis, logging, auth) — from a JSON base + env overrides + `gos` defaults, with
masked secrets and a focused resolution trace.

It's the combination of every other feature page:

- **[Two tags](01-two-tags.md)** naming and shaping each field,
- **[Priority](02-priority.md)** — env overrides layered over a JSON base over defaults,
- **[Sources](03-sources.md)** — JSON file + env + default composed with `WithSources`,
- **[Field types](04-field-types.md)** — durations, bools, ints, nested structs,
- **[Observability & masking](07-observability.md)** — a trace with secrets masked.

## Try it

```sh
go run ./examples/webservice
```

Output:

```text
payments-api v1.8.3 starting in "production"
  listen:   0.0.0.0:443  (TLS true)
  database: app@prod-db.internal:5432/payments (pool 50, password set: 12 chars)
  redis:    redis.internal:6379 (pool 20)
  logging:  level=warn format=json sample=0.25
  auth:     issuer=https://auth.internal ttl=30m0s

── resolution trace (who won each field; secrets masked) ──
Configuring main.Config: 32 fields
[Primary Source] json (loaded 16 fields)
[Defaults] applied for 8 fields

Overrides & Secrets:
  Environment string  ⇐ env (override) = production
  Version     string  ⇐ env (override) = 1.8.3
  Port        int     ⇐ env (override) = 443
  Host        string  ⇐ env (override) = prod-db.internal
  Password    string  ⇐ env = •••-pw (secret)
  LogLevel    string  ⇐ env (override) = warn
  JWTSecret   string  ⇐ env = •••key (secret)
```

Sixteen fields come from the JSON base, eight from defaults, and a handful of
operator env vars override the rest — with `Password` and `JWTSecret` masked in
the trace but real on the struct. This is what a production `Configure` call
looks like end to end.
