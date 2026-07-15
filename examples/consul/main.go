// Command consul-example fills a struct from a Consul KV prefix and then live-
// reloads it whenever a key changes — driven by Consul blocking queries.
//
// Bring up Consul (seeded with the demo keys) and run:
//
//	docker compose up -d
//	go run .
//
// While it runs, change a key and watch it reload instantly:
//
//	docker compose exec consul consul kv put config/orders/host 10.0.0.9
//	docker compose exec consul consul kv put config/orders/tags "eu,fast,beta"
//
// Stop with Ctrl-C, then `docker compose down`.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/consul"
	"github.com/goreflect/gostructor/snapshot"
)

// Config is filled from keys under the Consul prefix config/orders.
type Config struct {
	Host  string   `cfg:"host"`
	Port  int      `cfg:"port"`
	Tags  []string `cfg:"tags" gos:"optional"`
	Level string   `cfg:"level" gos:"default:info"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// A durable snapshot so the app still starts if Consul is down. Try it:
	// `docker compose stop consul`, rerun — it serves the last-known-good.
	store, err := snapshot.NewDirStore("./snapshot")
	if err != nil {
		panic(err)
	}

	src, err := consul.New(consul.Options{
		Address:  envOr("CONSUL_HTTP_ADDR", "127.0.0.1:8500"),
		Prefix:   "config/orders",
		Snapshot: store,
		Logger:   logger,
	})
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Println("watching Consul prefix config/orders (Ctrl-C to quit)")
	err = gostructor.Watch(ctx, &Config{}, func(cfg *Config, err error) {
		if err != nil {
			fmt.Printf("⚠️  reload rejected: %v\n", err)
			return
		}
		fmt.Printf("✅ %s:%d  level=%s  tags=%v\n", cfg.Host, cfg.Port, cfg.Level, cfg.Tags)
	},
		gostructor.WithSources(src, gostructor.Default()),
		gostructor.WithLogger(logger),
	)
	if err != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, "watch error:", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
