# Known limitations & error handling

## Known limitations

- `gostructor/hocon` and `gostructor/toml` parse a practical subset, not
  the full spec:
  - HOCON: no `${}` substitutions, no `include`, no duration/size unit
    literals (`10m`, `5 MB`), no string concatenation across values.
  - TOML: no datetimes, no inline tables (`{ a = 1 }`), no arrays of
    tables (`[[table]]`), no multiline/triple-quoted strings.

  A config file needing any of those produces a parse error rather than a
  silently wrong value. `gostructor/yaml` doesn't have this limitation —
  it wraps `goccy/go-yaml` rather than a hand-written parser.
- Nested struct sections work either flattened (an untagged struct field is
  descended into and its own tagged fields resolved individually) or whole (a
  struct field that itself carries a `cfg`/`gos` tag is filled atomically from
  one object). gostructor does not recurse *through* a pointer-to-struct field:
  `*SubConfig` is only fillable as a whole tagged value, not flattened.
- Struct fields must be exported; there's no `unsafe`-pointer trick to
  write into unexported ones, matching `encoding/json`'s convention.

## Error handling

Every error `Configure` returns belongs to one of a small, closed set of
categories, so you can tell exactly what went wrong — and whose fault it is —
without string-matching. Match the sentinels with `errors.Is` and the struct
types with `errors.As`; each struct type unwraps to its underlying cause.

| Error | When | Whose problem |
|---|---|---|
| `ErrInvalidTarget` (sentinel) | `target` isn't a non-nil pointer to a struct | the calling code |
| `*NotResolvedError` (wraps `ErrFieldNotResolved`) | a configured field (not `optional`) produced no value from any source | a missing env var / file key / secret |
| `*SourceError` | a source failed: file missing or malformed, Vault unreachable, ... | the backing store |
| `*ConvertError` | a value was produced but doesn't fit the field's type (fractional float into `int`, overflow, bad duration) | the value in the config |
| `*HookError` | a `WithHook` callback rejected the value or returned the wrong type | your validation/transform |

Every field-scoped error (`NotResolvedError`, `SourceError`, `ConvertError`,
`HookError`) implements the `FieldError` interface, so you can recover which
field failed without switching on the concrete type:

```go
type FieldError interface {
    error
    FieldName() string
}
```

```go
cfg, err := gostructor.Configure(&Config{})
switch {
case err == nil:
    // ok

case errors.Is(err, gostructor.ErrFieldNotResolved):
    var nre *gostructor.NotResolvedError
    errors.As(err, &nre)
    log.Fatalf("no value for %s (tried %s)", nre.Field, strings.Join(nre.Sources, ", "))

case func() bool { var se *gostructor.SourceError; return errors.As(err, &se) }():
    var se *gostructor.SourceError
    errors.As(err, &se)
    log.Fatalf("source %s failed for %s: %v", se.Source, se.Field, se.Unwrap())

default:
    var ce *gostructor.ConvertError
    if errors.As(err, &ce) {
        // ce.Field, ce.Value, ce.Target describe exactly what didn't fit,
        // and errors.As can reach the underlying *strconv.NumError etc.
        log.Fatalf("field %s: bad value %#v for %s", ce.Field, ce.Value, ce.Target)
    }
    log.Fatal(err)
}
```
