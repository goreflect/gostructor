# gostructor roadmap

Where the library is headed up to 2.0. These are rough ideas grouped by theme,
not dated commitments — the shortlist of what seems worth building, and what we
won't. Priorities move around.

## Where we are (v1.0)

The stable baseline everything below builds on:

- **Two-tag fill** from any mix of sources. `cfg` routes and names a field,
  `gos` carries behaviour (`default`, `secret`, `optional`, `sep`). Priority is
  the order you pass sources to `WithSources` — the first one that resolves wins.
- **Strict conversion.** Overflow, fractional-float-to-int, negative-to-unsigned
  and non-finite floats are errors, not silent truncation. Named types,
  `time.Duration`, `encoding.TextUnmarshaler`, pointers, arrays, slices,
  `[]Struct` and `map[string]Struct` are all supported. `time.Time` parses as
  RFC3339, or a custom layout via `gos:"layout:2006-01-02"`.
- **Typed errors** (`ErrInvalidTarget`, `NotResolvedError`, `SourceError`,
  `ConvertError`, `HookError`) that unwrap to the real cause.
- **Zero-dependency core.** YAML/TOML/HOCON/Vault each live in their own module.
- **`Source` interface** (`Name` + `Resolve`) as the extension point.
- **Resolution report** (`ConfigureWithReport`) and **secret masking**
  (`WithMasker`).
- **Hot reload**: `Watch`, debounce, validated last-known-good reload — in the
  core, with file watching in `gostructor/watch`.
- **Git as a config source** (`gostructor/git`) with snapshots
  (`gostructor/snapshot`).
- **Config-server adapters**: Consul, etcd, Spring Cloud Config, Vault.
- **Code generation** (`gostructor-gen`): a reflection-free `Fill` method,
  opt-in via `WithEngine`, kept honest by a generated test that diffs it against
  the reflective path.

## Principles

- **Zero-dependency core.** Anything needing a third-party dep ships as its own
  module.
- **No global state.** Everything is a per-call `Option`.
- **Priority is source order.** Features compose with that, not around it.
- **Strict and predictable.** No fuzzy key matching, no silent coercion.

---

## Open ideas

### Observability

- Turn tracing on from the environment or a CLI flag, not just from code.

### Ergonomics

- Self-documenting config: a description tag plus a helper that prints each
  field, its sources and its default.
- Generate a commented sample config file from the struct, so a missing config
  gives the operator a template to edit instead of an error.
- Fuller env handling (prefixes, `*_FILE` indirection, `${VAR}` expansion,
  separators) so reaching for a second env-parsing library is never necessary.
- Derive CLI flags from the struct — one flag per field, composing with the
  other sources through the normal priority model rather than as a parallel
  system.

### Codegen extras

Things that come cheap once the generator already parses the struct:

- A directive to pick key casing per source once, resolved at generate time.
- A `required` check with its own error type.
- Auto-wiring a `Validate()` method so it can't be forgotten.
- Emitting a canonical sample config a test can diff against, to catch drift.

### Integrations

- An `uber-go/fx` provider so DI apps wire config in a line, with optional live
  reload.
- Embedded local sources (SQLite, LevelDB) for edge and P2P nodes that have no
  central config server.

### Adoption

- A migration CLI that detects the existing config library (viper, cleanenv,
  envconfig, caarlos0/env, koanf, …) and rewrites its tags and call sites.
- Compatibility shims for viper and cleanenv: keep the old call sites, swap the
  engine underneath.

## Out of scope

- No global singleton or package-level state.
- No writing config back to disk.
- No case-insensitive or fuzzy key matching.
