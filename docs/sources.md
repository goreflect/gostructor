# Sources in detail

Every source shares the one `cfg` tag. The source name is what you target in a
per-source override (`cfg:"port,env:DB_PORT"`) and what appears in the resolution
trace.

| Source name | Reads from | Module | Requires | Key from base name |
|---|---|---|---|---|
| `default` | Literal `gos:"default:..."` value | core | — | — (uses gos) |
| `env` | Environment variable | core | — | `SCREAMING_SNAKE` |
| `json` | JSON file | core | `GOSTRUCTOR_JSON=path/to/file.json` | name as written; `.` nests |
| `ini` | INI file | core | `GOSTRUCTOR_INI=path/to/file.ini` | global key; `section#key` overrides |
| `yaml` | YAML file | `gostructor/yaml` | `GOSTRUCTOR_YAML=path/to/file.yml` | name as written; `.` nests |
| `toml` | TOML file | `gostructor/toml` | `GOSTRUCTOR_TOML=path/to/file.toml` | top-level key; `table#key` overrides |
| `hocon` | HOCON file | `gostructor/hocon` | `GOSTRUCTOR_HOCON=path/to/file.hocon` | name as written; `.` nests |
| `vault` | HashiCorp Vault secret | `gostructor/vault` | `VAULT_ADDR`, `VAULT_TOKEN` | **override only** (`vault:path#key`) |

Non-core sources must be added explicitly via `WithSources`; the core module
doesn't import them, which is what keeps it dependency-free. The default source
list is just `Env, Default`. File and secret sources are opt-in, so a bare `cfg`
base name never triggers an unexpected file load.

## Defaults

```go
type Config struct {
    Retries int    `gos:"default:3"`
    Flags   []bool `gos:"sep:|,default:true|false|true"`
}
```

`gos:"default:<v>"` is the literal default, resolved by the `default` source.
Slices split on the field separator (comma by default; a non-comma `sep` when
the value itself contains commas). There is no map syntax for defaults (see
[field-types.md](field-types.md)).

## Environment variables

```go
type Config struct {
    Port    int    `cfg:"port"`                   // reads PORT
    Legacy  int    `cfg:"legacy,env:OLD_PORT"`    // reads OLD_PORT (override)
    Signals []bool `cfg:"signals"`                // reads SIGNALS=true,false,true
}
```

The env source turns the base name into `SCREAMING_SNAKE_CASE` (`port` → `PORT`,
`maxConns` → `MAX_CONNS`). Pin a different variable with an `env:` override.

## JSON and INI (core)

```go
type Config struct {
    Host string `cfg:"host,json:server.host"`
    Tags []int  `cfg:"tags,json:server.tags"`
}
```

```go
cfg, err := gostructor.Configure(&Config{},
    gostructor.WithSources(gostructor.JSONFile("config.json")))
```

```json
{"server": {"host": "0.0.0.0", "tags": [1, 2, 3]}}
```

A top-level key uses the base name (`cfg:"host"` → the `host` key); a nested
value is addressed with a dotted `json:` override (`json:server.host`), no
matching nested Go struct required. A bare object key like `cfg:"server"` on a
`map[string]T` field addresses the whole nested object.

INI uses `section#key`, since INI's own structure is two levels (section, then
key). A base name reads the global, section-less part; a section value uses an
`ini:` override:

```go
type Config struct {
    Password string `cfg:"password,ini:database#password"`
}
```

```ini
[database]
password = secret
```

## YAML, TOML, HOCON (optional modules)

Each lives in its own module and mirrors JSON's addressing style (YAML) or
INI's (TOML/HOCON), targeted by its source name, and registered explicitly via
`WithSources`:

```go
import "github.com/goreflect/gostructor/yaml"

type Config struct {
    Host string   `cfg:"host,yaml:server.host"`
    Tags []string `cfg:"tags,yaml:server.tags"`
}

os.Setenv(yaml.FileEnvVar, "config.yml")
cfg, err := gostructor.Configure(&Config{}, gostructor.WithSources(yaml.New()))
```

```yaml
server:
  host: 0.0.0.0
  tags:
    - primary
    - eu-west
```

`gostructor/toml` and `gostructor/hocon` use hand-written parsers that cover a
practical subset of each format, not the full spec. See
[limitations.md](limitations.md) for what's out of scope before you commit a
config file that needs it.

## HashiCorp Vault (optional module)

Set `VAULT_ADDR` and `VAULT_TOKEN` (the same variables the `vault` CLI uses),
then give each field an explicit `vault:path/to/secret#key` override. Vault has
no name-based default, since a secret path can't be guessed from a field name, so
a field without a `vault:` override is skipped by the vault source:

```go
import "github.com/goreflect/gostructor/vault"

type Config struct {
    APIKey    string  `cfg:"apiKey,vault:my-service/stage/creds#api-key" gos:"secret"`
    RateLimit int16   `cfg:"rateLimit,vault:my-service/stage/limits#rate"`
    Allowlist []int32 `cfg:"allowlist,vault:my-service/stage/net#allowlist"` // comma-separated secret value
}

cfg, err := gostructor.Configure(&Config{}, gostructor.WithSources(vault.New()))
```

Built on the official [`hashicorp/vault/api`](https://github.com/hashicorp/vault)
client, not a third-party wrapper.

## Priority: the source order

A field can be resolvable by several sources at once. There's no priority tag and
no global selector: priority is the order of the sources you pass to
`WithSources`, and the first source that reports a value wins. To change which
source leads for a whole config, reorder the slice:

```go
// prod: an operator's env override leads; the default is the fallback.
gostructor.Configure(&cfg, gostructor.WithSources(gostructor.Env(), gostructor.Default()))

// dev: the baked-in default leads, so a stray env var is ignored.
gostructor.Configure(&cfg, gostructor.WithSources(gostructor.Default(), gostructor.Env()))
```

The same struct resolves differently purely from the list order — see
`examples/priority` for a runnable demo.

## Writing your own Source

```go
type Source interface {
    Name() string
    Resolve(field FieldContext) (value any, found bool, err error)
}
```

`Name` is the source's identity ("yaml") and its `cfg` override key. In
`Resolve`, ask the field for your key with
`field.SourceKey("yaml", gostructor.Identity)`. That returns the `yaml:` override
if present, otherwise your naming strategy applied to the base name, otherwise
`""` when the field doesn't apply to you. Return `found=false` when you have
nothing for a field, and `Configure` falls through to the next source instead of
treating it as an error. `gostructor/yaml`'s `source.go` is a short, complete
example, about seventy lines including file loading and error handling.
