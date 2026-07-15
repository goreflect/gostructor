# Field types

← [Back to features](README.md)

gostructor fills a broad range of Go types, and does so with **strict, lossless
conversion** — never silent truncation or wraparound.

Supported: all sized ints/uints, floats, `string`, `bool`, `time.Duration`
(from `"1h30m"`), any named scalar type (`type Level int`), any
`encoding.TextUnmarshaler` (`time.Time`, `net.IP`, your own enums), pointers to
any of the above, slices and fixed-size arrays (including nested), and
`map[K]V` / `[]Struct` / `map[string]Struct` when the source is structured data
(JSON/YAML/TOML/HOCON).

Numeric conversions are exact: a fractional float into an integer (`3.9` → `int`),
an overflow (`300` → `int8`), a negative into an unsigned field, and non-finite
floats are all hard `*ConvertError`s — see [Error taxonomy](06-errors.md).

## Try it

The [`types`](../../examples/types) example fills one struct covering the whole
range from a single JSON source: a duration, a slice, a float array, a
`time.Time`, a `net.IP`, a named `LogLevel` type, a `*int`, a `[]Struct`, a
`map[string]Struct`, and a `map[string]string`.

```sh
go run ./examples/types
```

Output:

```text
Timeout   1h30m0s (time.Duration)
Retries   [1 2 3]
Coords    [51.5074 -0.1278]
StartedAt Tue, 02 Jan 2024 15:04:05 UTC
BindIP    10.0.0.42
Level     "debug" (main.LogLevel)
MaxConns  100 (via *int)
Backends  [{URL:eu-1 Weight:5} {URL:eu-2 Weight:3}]
Shards    map[payments:{URL:shard-a Weight:1}]
Labels    map[team:platform tier:gold]
```

Each line is a different type category filled from JSON: durations parse from Go
duration strings, `time.Time`/`net.IP` via `TextUnmarshaler`, named types keep
their Go type, pointers are allocated and set, and structs/maps are matched
field-by-field by name.
