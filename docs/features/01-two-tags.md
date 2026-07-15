# Two-tag configuration

← [Back to features](README.md)

gostructor fills a struct's fields from two small, cleanly separated tags:

- **`cfg`** — *routing & naming*: `cfg:"base_name,source:override,..."`. The base
  name is what each source looks up; each source derives its own key from it (the
  env source reads `PORT` from `port`). A `source:override` pins an exact key for
  one source.
- **`gos`** — *behavior*: `gos:"default:8080,secret,optional,sep:;"` — a literal
  default, a masked-secret flag, an unresolved-is-ok flag, and a slice separator.

With no options, `Configure` uses a minimal default source list — `Env` then
`Default` — so an environment variable beats the baked-in default. A *configured*
field that no source can resolve is a hard error (unless it's `gos:"optional"`):
**surprises are errors**.

```go
type Config struct {
	Host  string `cfg:"host" gos:"default:0.0.0.0"`
	Port  int    `cfg:"port" gos:"default:8080"`
	Debug bool   `cfg:"debug" gos:"default:false"`
}

cfg, err := gostructor.Configure(&Config{})
```

## Try it

The [`basic`](../../examples/basic) example sets `PORT=9090` in the process, then
calls `Configure` with no options. The env source names its variable from the
base name (`port` → `PORT`), so the env value wins over the `gos` default; `host`
and `debug` fall back to their defaults.

```sh
go run ./examples/basic
```

Output:

```text
&{Host:0.0.0.0 Port:9090 Debug:false}
```

`Port` is `9090` (from the env var), while `Host` and `Debug` keep their
`gos:"default:..."` values because nothing else supplied them.
