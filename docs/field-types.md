# Supported field types

- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `string`
- `bool`
- `time.Duration`, filled from a Go duration string like `"1h30m"` (numeric
  sources are read as nanoseconds, matching `encoding/json`).
- any named scalar type (`type Level int`, `type Env string`, ...) — the
  value is converted to the field's underlying kind and keeps the named type.
- any type implementing `encoding.TextUnmarshaler` (`time.Time`, `net.IP`,
  your own enums), filled by handing it the string form.
- pointers to any of the above (`*int`, `*time.Time`, ...): the value is
  allocated and set. Multi-level pointers (`**T`) are not supported.
- slices and fixed-size arrays of any supported element type, including
  nested ones — `[]int32`, `[3]string`, `[][]int`, `[]time.Duration`.
  For an array the source must have exactly the array's length.
- `map[K]V` and `[]Struct` / `map[string]Struct`, when the source is
  structured data (JSON, YAML, TOML, HOCON) and the tag addresses a whole
  nested object rather than one leaf value. Struct elements are filled by
  matching each exported field to an object key by name, case-insensitively.
  The `env`, `default`, `ini`, and `vault` sources encode values as a flat
  string and so can only populate slices, not maps or struct elements.

Numeric conversions are **exact**: a fractional float into an integer field
(`3.9` → `int`), a value that overflows the field's width (`300` → `int8`), a
negative number into an unsigned field, and non-finite floats are all hard
errors — never silent truncation or wraparound. A conversion failure is
reported as a `*ConvertError` (see [limitations.md](limitations.md#error-handling)).
