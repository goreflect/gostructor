# Git as source of truth

Drives configuration from a git repo: a JSON file at a **ref** (branch/tag =
version), followed for new commits (drift) and switchable at runtime.

## Run it

```sh
docker compose up -d     # git server seeded with branch `main` + tag `v1`
go run .
```

The program starts on `main`:

```
pid 41234 — reading git config at ref main; `kill -HUP 41234` to toggle main/v1 (Ctrl-C to quit)
✅ main.internal:9000  release=2025.11  (commit 3f1a2b7)
```

### Switch version at runtime

Send `SIGHUP` to toggle between `main` and the `v1` tag — the config reloads
from a different git ref, without a restart:

```sh
kill -HUP <pid>     # use the pid printed above
```

```
↪  switching to ref "v1"
✅ v1.internal:8001  release=2025.10  (commit a90c1e4)
```

### Drift

The source polls the ref every 5s; a new commit on `main` is picked up
automatically. To try it, push a change into the container's repo:

```sh
docker compose exec gitserver sh -c '
  cd /tmp/work &&
  sed -i "s/9000/9100/" services/api/config.json &&
  git commit -aqm "bump port" &&
  git push -q /srv/config.git main'
```

Within ~5s the running program reloads to the new commit.

## Last-known-good

Good fetches are written to `./snapshot`. Stop the git server and rerun — the
app still starts, serving the last snapshot:

```sh
docker compose stop gitserver
go run .
```

## How it works

- **Pure Go** via `go-git` — no `git` binary needed on the host.
- **`git.New`** clones in memory, reads the file at the ref, and resolves fields
  with the same nested-key addressing as the JSON source.
- **`SetVersion(ref)`** re-points to another branch/tag/commit and reloads.
- **`Snapshot`** (the `gostructor/snapshot` "файлопомойка") keeps the app up
  through a git outage.

### Any format

The config file needn't be JSON. Pass a decoder for the format you keep in git:

```go
git.New(git.Options{Repo: ..., Path: "config.yaml", Decoder: yaml.Decode})
git.New(git.Options{Repo: ..., Path: "app.env",     Decoder: gostructor.DecodeKeyValue})
```

`gostructor.DecodeINI` / `gostructor.DecodeKeyValue` are zero-dep; `yaml.Decode`,
`toml.Decode`, `hocon.Decode` come from those modules.

## Clean up

```sh
docker compose down
rm -rf snapshot
```
