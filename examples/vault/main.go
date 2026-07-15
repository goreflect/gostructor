// Command vault-example fills secret fields from HashiCorp Vault and live-
// reloads them when a secret is rotated. Vault has no change-push, so the vault
// source polls the referenced secrets; a rotation triggers a reload.
//
//	docker compose up -d
//	go run .
//
// While it runs, rotate the secret and watch it reload (the poll picks it up
// within a few seconds):
//
//	docker compose exec vault vault kv put kv/orders/db password=rotated-pass max_conns=50
//
// Ctrl-C to quit, then `docker compose down`.
//
// The secret is masked in all gostructor output (gos:"secret").
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/vault"
)

// Config pulls DB credentials from a Vault KV v1 secret at kv/orders/db. Each
// field names its secret path and key via a vault: override.
type Config struct {
	Password string `cfg:"password,vault:kv/orders/db#password" gos:"secret"`
	MaxConns int    `cfg:"maxConns,vault:kv/orders/db#max_conns"`
	Level    string `cfg:"level" gos:"default:info"`
}

func main() {
	// The vault client reads VAULT_ADDR and VAULT_TOKEN from the environment,
	// the same variables the `vault` CLI uses. docker-compose runs a dev server
	// with a known root token; export these before running:
	//   export VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=root
	if os.Getenv("VAULT_ADDR") == "" {
		os.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	}
	if os.Getenv("VAULT_TOKEN") == "" {
		os.Setenv("VAULT_TOKEN", "root")
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	src := vault.New(
		vault.WithPollInterval(2*time.Second),
		vault.WithLogger(logger),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Println("watching Vault secret kv/orders/db (Ctrl-C to quit)")
	err := gostructor.Watch(ctx, &Config{}, func(cfg *Config, err error) {
		if err != nil {
			fmt.Printf("⚠️  reload rejected: %v\n", err)
			return
		}
		// cfg.Password is a real value here; we just don't print it. gostructor
		// masks it anywhere *it* would render it (reports, traces, errors).
		fmt.Printf("✅ maxConns=%d level=%s  (password loaded, %d chars)\n",
			cfg.MaxConns, cfg.Level, len(cfg.Password))
	},
		gostructor.WithSources(src, gostructor.Default()),
		gostructor.WithLogger(logger),
	)
	if err != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, "watch error:", err)
		os.Exit(1)
	}
}
