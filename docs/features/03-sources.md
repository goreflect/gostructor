# Sources

← [Back to features](README.md)

A source is anything that can produce a value for a field. gostructor ships four
in the zero-dependency core and four in their own modules; you compose them, in
priority order, via `WithSources`.

| Source | Reads from | Module | Requires | Key from base name |
|---|---|---|---|---|
| `default` | Literal `gos:"default:..."` value | core | — | — (uses `gos`) |
| `env` | Environment variable | core | — | `SCREAMING_SNAKE` |
| `json` | JSON file | core | `GOSTRUCTOR_JSON=path` | name as written; `.` nests |
| `ini` | INI file | core | `GOSTRUCTOR_INI=path` | global key; `section#key` overrides |
| `yaml` | YAML file | `gostructor/yaml` | `GOSTRUCTOR_YAML=path` | name as written; `.` nests |
| `toml` | TOML file | `gostructor/toml` | `GOSTRUCTOR_TOML=path` | top-level key; `table#key` overrides |
| `hocon` | HOCON file | `gostructor/hocon` | `GOSTRUCTOR_HOCON=path` | name as written; `.` nests |
| `vault` | HashiCorp Vault secret | `gostructor/vault` | `VAULT_ADDR`, `VAULT_TOKEN` | **override only** (`vault:path#key`) |

Each file source also has a `*File(path)` variant (`JSONFile`, `INIFile`,
`yaml.File`, ...) to read from an explicit path instead of its environment
variable. Non-core sources must be added explicitly — the zero-dependency core
has no way to know they exist.

```go
cfg, err := gostructor.Configure(&Config{}, gostructor.WithSources(
	gostructor.Env(),
	gostructor.JSONFile("config.json"),
	gostructor.INIFile("config.ini"),
	gostructor.Default(),
))
```

## Try it — core file sources (JSON + INI)

The [`filesources`](../../examples/filesources) example writes a JSON and an INI
file to a temp dir, then resolves one struct from `Env, JSONFile, INIFile,
Default` — `Host` comes from JSON, `Port` from INI, and `MaxConns` from the `gos`
default because neither file has it.

```sh
go run ./examples/filesources
```

Output:

```text
Host     0.0.0.0    (from JSON file)
Port     9090       (from INI file)
MaxConns 256        (from gos default — absent in both files)
```

## Try it — an external module (YAML)

The [`multisource`](../../examples/multisource) example is its own Go module: it
pulls in `gostructor/yaml`, reads a real `config.yml`, and layers env + default +
a validation hook on top. Run it from its own directory (or via the path — the
module has a `replace` back to the repo):

```sh
go run ./examples/multisource
```

Output:

```text
&{Host:0.0.0.0 Port:8080}
```

### Adding a module source

```sh
go get github.com/goreflect/gostructor/yaml   # or /toml, /hocon, /vault
```

```go
import "github.com/goreflect/gostructor/yaml"

os.Setenv(yaml.FileEnvVar, "config.yml")
cfg, err := gostructor.Configure(&Config{}, gostructor.WithSources(yaml.New()))
```

> `gostructor/toml` and `gostructor/hocon` are hand-written parsers for a
> practical subset of each format (no `${}` substitutions, inline tables,
> arrays-of-tables, etc.) — see [Known limitations](../limitations.md).
> `gostructor/yaml` wraps `goccy/go-yaml` and has no such subset limit.
