# Consul KV — live config with blocking queries

Fills a struct from a Consul KV prefix and hot-reloads it when any key changes.
The watch uses Consul **blocking queries**, so a change is picked up the instant
it lands — no polling.

## Run it

```sh
docker compose up -d     # dev-mode Consul, seeded with config/orders/*
go run .
```

You'll see the seeded config, then change a key in another terminal:

```sh
docker compose exec consul consul kv put config/orders/host 10.0.0.9
docker compose exec consul consul kv put config/orders/tags "eu,fast,beta"
```

The program reloads immediately:

```
✅ 10.0.0.5:5432  level=info  tags=[eu primary]
✅ 10.0.0.9:5432  level=info  tags=[eu primary]
✅ 10.0.0.9:5432  level=info  tags=[eu fast beta]
```

## Last-known-good

Each good read is written to `./snapshot`. Stop Consul and rerun — the app still
starts, serving the last snapshot instead of failing:

```sh
docker compose stop consul
go run .    # logs a warning, serves ./snapshot
```

## Clean up

```sh
docker compose down
rm -rf snapshot
```

## Key addressing

A field's key is its base name joined onto the prefix: `cfg:"host"` →
`config/orders/host`. Pin a different key with a per-source override:
`cfg:"host,consul:database/primary_host"`.
