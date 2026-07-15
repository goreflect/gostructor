# etcd — live config with the watch API

Fills a struct from an etcd key prefix and hot-reloads it using etcd's native
**watch** stream (event-driven, no polling).

## Run it

```sh
docker compose up -d     # etcd, seeded with config/orders/*
go run .
```

Change a key in another terminal:

```sh
docker compose exec etcd etcdctl put config/orders/host 10.0.0.9
docker compose exec etcd etcdctl put config/orders/level debug
```

The program reloads on each event:

```
✅ 10.0.0.5:5432  level=info
✅ 10.0.0.9:5432  level=info
✅ 10.0.0.9:5432  level=debug
```

## Last-known-good

Good reads are written to `./snapshot`; stop etcd and rerun to serve the last
snapshot instead of failing:

```sh
docker compose stop etcd
go run .
```

## Clean up

```sh
docker compose down
rm -rf snapshot
```
