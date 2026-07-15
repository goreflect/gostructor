# Hooks: validation & transformation

← [Back to features](README.md)

`WithHook` runs after a value is resolved but before it's set on the struct.
Return an error to **reject** the value, or a different value to **transform**
it. The hook sees the typed, converted value, not the raw string.

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

A hook that returns an error surfaces as a `*HookError` (see
[Error taxonomy](06-errors.md)), naming the field and your message.

## Try it

The [`hooks`](../../examples/hooks) example runs twice: first a valid config that
gets **normalised** (a string trimmed/lowercased) and **validated** (port in
range), then a config with a privileged port that the same hook **rejects**.

```sh
go run ./examples/hooks
```

Output:

```text
normalised + validated: Env="staging" Port=9090
rejected as expected: gostructor: field "Port": hook rejected value: port 80 is privileged
```

The first line shows a transformed value landing on the struct; the second shows
the same hook turning a bad value into a typed `*HookError` instead of letting it
through.
