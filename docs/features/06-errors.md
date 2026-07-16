# Typed errors

← [Back to features](README.md)

Every error `Configure` returns belongs to a small, closed set of categories, so
you can tell what went wrong, and whose fault it is, without string-matching.
Match sentinels with `errors.Is` and struct types with `errors.As`; each struct
unwraps to its underlying cause.

| Error | When | Whose problem |
|---|---|---|
| `ErrInvalidTarget` (sentinel) | `target` isn't a non-nil pointer to a struct | the calling code |
| `*NotResolvedError` (wraps `ErrFieldNotResolved`) | a configured, non-`optional` field produced no value | a missing env var / file key / secret |
| `*SourceError` | a source failed: file missing or malformed, Vault unreachable | the backing store |
| `*ConvertError` | a value was produced but doesn't fit the field's type | the value in the config |
| `*HookError` | a `WithHook` callback rejected the value or returned the wrong type | your validation/transform |

Every field-scoped error implements `FieldError` (an `error` with a
`FieldName() string`), so you can recover which field failed without switching on
the concrete type.

## Try it

The [`errors`](../../examples/errors) example deliberately triggers each failure
mode in turn and classifies it with `errors.Is`/`errors.As`.

```sh
go run ./examples/errors
```

Output:

```text
• unresolved field (no source produced a value)
  err: gostructor: field "APIKey": no configured source produced a value (tried env, default)
  → NotResolvedError on "APIKey", tried [env default]

• convert error (value doesn't fit the field type)
  err: gostructor: field "Port": cannot convert "not-a-number" into int: convert: cannot convert "not-a-number" (string) into int: strconv.ParseInt: parsing "not-a-number": invalid syntax
  → ConvertError on "Port" (bad value in config)

• source error (backing store failed)
  err: gostructor: field "Name": source json failed: gostructor: reading JSON file "/no/such/config.json": open /no/such/config.json: no such file or directory
  → SourceError on "Name" via json (backing store)

• hook error (validation rejected the value)
  err: gostructor: field "Port": hook rejected value: port 80 is privileged
  → HookError on "Port" (your validation)

• invalid target (a programming bug, not a config gap)
  err: gostructor: target must be a non-nil pointer to a struct: got nil
  → ErrInvalidTarget: the call itself is wrong; fix the code.
```

Each block shows the raw error message and the concrete type you'd recover with
`errors.As`, enough to route the failure to the right owner (operator, config
author, or the calling code).
