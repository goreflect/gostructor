# Vault — secrets with live rotation

Fills secret fields from HashiCorp Vault and reloads them when a secret is
rotated. Vault (OSS) has no change-push, so the source **polls** the referenced
secrets; a rotation is picked up within the poll interval (2s here).

## Run it

```sh
docker compose up -d     # dev-mode Vault (root token "root"), seeded kv/orders/db
export VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=root
go run .
```

Rotate the secret in another terminal:

```sh
docker compose exec vault vault kv put kv/orders/db password=rotated max_conns=50
```

The program reloads on the next poll:

```
✅ maxConns=20 level=info  (password loaded, 10 chars)
✅ maxConns=50 level=info  (password loaded, 7 chars)
```

## Secret masking

`Password` is tagged `gos:"secret"`, so gostructor masks it in every place *it*
renders a value — the resolution trace, logs, and error messages. The example
prints only its length, never the value.

## KV v1 vs v2

The seeder mounts `kv/` as **KV version 1** (flat secrets). The core vault
source reads `secret.Data[key]` directly, which matches v1. For a KV v2 mount,
address the data path explicitly: `vault:secret/data/orders/db#password` and
read from the nested `data` map.

## Clean up

```sh
docker compose down
```
