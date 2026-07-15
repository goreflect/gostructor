# Priority = source order

← [Back to features](README.md)

This is the feature that sets gostructor apart from a merge-everything-into-one-map
unmarshaller. A field can be resolvable by several sources at once. There is **no
priority tag and no global selector** — priority is simply the order of the sources
you pass to `WithSources`. The first source that reports a value wins.

To flip which source leads for a whole config, reorder the slice:

```go
// prod: an operator's env override leads; the default is the fallback.
gostructor.Configure(&cfg, gostructor.WithSources(gostructor.Env(), gostructor.Default()))

// dev: the baked-in default leads, so a stray env var is ignored.
gostructor.Configure(&cfg, gostructor.WithSources(gostructor.Default(), gostructor.Env()))
```

## Try it

The [`priority`](../../examples/priority) example sets `LOG_LEVEL=debug` and
`TRACE_RATE=1.0` in the environment, then resolves the **same struct** twice —
once with `Env, Default` (prod) and once with `Default, Env` (dev). Only the
source order differs.

```sh
go run ./examples/priority
```

Output:

```text
[prod] LogLevel=debug TraceRate=1.0
[dev ] LogLevel=info  TraceRate=0.01

Same struct, same env vars — only the WithSources order changed:
  prod trusts the operator's env override; dev pins the local default.
```

Same struct, same env vars — `prod` trusts the operator's env override, `dev`
pins the local default and ignores the stray env vars. The priority decision
lives at the call site, not in the tags.
