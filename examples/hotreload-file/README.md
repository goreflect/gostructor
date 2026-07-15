# Hot reload from a file (no Docker)

The simplest live-config demo: `gostructor.Watch` + a `gostructor/watch`
fsnotify file source. Edit a JSON file, the struct re-fills.

```sh
go run .
```

Then, in another terminal, edit `config.json` and save. You'll see:

```
✅ config now: 127.0.0.1:8080  message="edit me and save"
✅ config now: 127.0.0.1:9090  message="changed!"
```

Try a **bad** edit — set `"port": 70000`, or write invalid JSON. The reload is
rejected and the previous good config keeps serving:

```
⚠️  reload rejected, keeping last good config: port 70000 out of range 1..65535
```

## What it shows

- **`gostructor.Watch`** — fills once, then re-fills a *fresh* struct on every
  change through the same resolution + hook path as the initial fill.
- **`WithDebounce(200ms)`** — an editor writes a file in several syscalls; the
  debounce collapses that burst into a single reload.
- **`WithValidate`** — a whole-struct gate; a failing reload is delivered as an
  error and the last-known-good stays in place (transactional reload).
- **`gostructor/watch.JSONFile`** — watches the file's directory (so it survives
  atomic rename-over-write saves) and re-reads on change.
