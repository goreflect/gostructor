// Command etcd-example fills a struct from an etcd key prefix and live-reloads
// it using the native etcd watch API.
//
//	docker compose up -d
//	go run .
//
// While it runs, change a key and watch it reload:
//
//	docker compose exec etcd etcdctl put config/orders/host 10.0.0.9
//	docker compose exec etcd etcdctl put config/orders/level debug
//
// Ctrl-C to quit, then `docker compose down`.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/etcd"
	"github.com/goreflect/gostructor/snapshot"
)

// Config is filled from keys under the etcd prefix config/orders.
type Config struct {
	Host  string `cfg:"host"`
	Port  int    `cfg:"port"`
	Level string `cfg:"level" gos:"default:info"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	store, err := snapshot.NewDirStore("./snapshot")
	if err != nil {
		panic(err)
	}

	src, err := etcd.New(etcd.Options{
		Endpoints: strings.Split(envOr("ETCD_ENDPOINTS", "http://127.0.0.1:2379"), ","),
		Prefix:    "config/orders",
		Snapshot:  store,
		Logger:    logger,
	})
	if err != nil {
		panic(err)
	}
	defer src.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Println("watching etcd prefix config/orders (Ctrl-C to quit)")
	err = gostructor.Watch(ctx, &Config{}, func(cfg *Config, err error) {
		if err != nil {
			fmt.Printf("⚠️  reload rejected: %v\n", err)
			return
		}
		fmt.Printf("✅ %s:%d  level=%s\n", cfg.Host, cfg.Port, cfg.Level)
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
