# Spring Cloud Config Server

Fills a struct from a Spring Cloud Config Server via `GET /{app}/{profile}` and
hot-reloads by polling (Spring Cloud Config has no change-push). This is the
worked answer to *"can I point gostructor at a config server?"* — yes, the same
`Source` shape as Vault, over plain HTTP.

## Run it

```sh
docker compose up -d     # config server, native backend over ./config
go run .
```

```
✅ api.internal:8443  level=warn  (version 8d3f9c2a...)
```

Edit the served file and save; the program reloads on the next poll (3s):

```sh
# macOS
sed -i '' 's/8443/8500/' config/orders-production.yml
# linux
sed -i    's/8443/8500/' config/orders-production.yml
```

```
✅ api.internal:8500  level=warn  (version 1b77e0f4...)
```

## Property addressing

Spring returns flat, dotted property keys. A field targets one by its base name
or a `springcloud:` override: `cfg:"server.port,springcloud:server.port"`.
Multiple `propertySources` are merged with Spring's precedence (earlier wins).

## Last-known-good

Good fetches are saved to `./snapshot`; stop the server and rerun to serve the
last snapshot.

## Notes

- The **native** backend reports no `version`, so the source fingerprints the
  response body to detect change. A **git-backed** server reports the backing
  commit as `version`, which the source uses directly.
- Point at a git-backed server instead by setting `CONFIG_SERVER` and giving the
  source a `Label` (the git branch/tag).

## Clean up

```sh
docker compose down
rm -rf snapshot
```
