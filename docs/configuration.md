# Configuration reference

The full grammar of the two tags, the `Configure` API, hooks, and logging.
For a task-first tour with runnable demos, see [`docs/features`](features/).

## The `Configure` API

```go
func Configure[T any](target *T, opts ...Option) (*T, error)
```

`Configure` reads the `cfg`/`gos` tags on `target`, tries each field's sources
in order, and returns the same pointer you passed in — so the argument is also
the result. A field with no gostructor tags is left untouched; a *configured*
field that no source can resolve is a hard error (unless it's `gos:"optional"`).

With no options, `Configure` uses a minimal default source list — `Env` then
`Default`. Bring in file and secret sources with `WithSources`, which replaces
the default list with an explicit, ordered one:

```go
import (
    "github.com/goreflect/gostructor"
    "github.com/goreflect/gostructor/yaml"
)

cfg, err := gostructor.Configure(&Config{}, gostructor.WithSources(
    gostructor.Env(),
    yaml.New(),
    gostructor.Default(),
))
```

Sources are tried in the order given; the first one that reports a value for a
field wins — that order **is** the priority (see [sources.md](sources.md)).
`Env()`, `Default()`, `JSON()`, and `INI()` are built into the core module;
each file source also has a `*File(path string)` variant (`JSONFile`,
`INIFile`, `yaml.File`, ...) to read from an explicit path instead of the
source's environment variable.

`ConfigureWithReport` fills the struct *and* returns a resolution trace — see
[observability.md](observability.md).

## The `cfg` and `gos` tags

```go
type Config struct {
    // env SERVER_PORT (override) or JSON server.port; default 8080.
    Port     int    `cfg:"port,env:SERVER_PORT,json:server.port" gos:"default:8080"`
    // env HOST / JSON "host" from the base name; no default.
    Host     string `cfg:"host"`
    // masked in the trace; required (errors if no source has it).
    Password string `cfg:"password,vault:secret/app#pw" gos:"secret"`
    // fine to be missing: stays the zero value.
    Debug    bool   `cfg:"debug" gos:"optional"`
}
```

**`cfg` — routing & naming.** `cfg:"base_name,source:override,..."`. The first
element is the base name each source turns into its own key via a naming
strategy (env upper-snakes it; file sources use it verbatim). A `source:override`
pins an exact key for one source — use it for nested file paths
(`json:server.host`), an INI/TOML section (`ini:db#password`), a legacy env
name (`env:DB_PORT_LEGACY`), or a Vault path (`vault:secret/app#pw`).

**`gos` — behavior & metadata.** A comma-list of `key:value` meta and bare flags:

| Token | Kind | Effect |
|---|---|---|
| `default:<v>` | meta | Literal fallback, resolved by the `default` source. |
| `sep:<char>` | meta | Separator for splitting a flat value into a slice (default `,`). |
| `secret` | flag | Mask the value in the trace, logs, and `*ConvertError`. |
| `optional` | flag | Do not error when no source resolves the field. |

> Comma is the `gos` list separator, so a meta value can't contain a comma.
> For a multi-valued slice default, choose a non-comma `sep` and use it in the
> default too: `gos:"sep:|,default:a|b|c"`.

## Hooks: validation and transformation

`WithHook` runs after a value is resolved but before it's set on the
struct — return an error to reject it, or a different value to transform it:

```go
cfg, err := gostructor.Configure(&Config{}, gostructor.WithHook(
    func(field gostructor.FieldContext, value any) (any, error) {
        if field.Name == "Port" && value.(int) < 1024 {
            return nil, fmt.Errorf("port %v is a privileged port", value)
        }
        return value, nil
    },
))
```

## Logging

```go
logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
cfg, err := gostructor.Configure(&Config{}, gostructor.WithLogger(logger))
```

With no `WithLogger`, `Configure` logs nothing. `WithTrace()` logs the
resolution report through your logger without changing call sites.
