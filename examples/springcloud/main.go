// Command springcloud-example fills a struct from a Spring Cloud Config Server
// and live-reloads it by polling for changes (Spring Cloud Config has no
// change-push).
//
//	docker compose up -d
//	go run .
//
// The server serves ./config/orders-production.yml. Edit that file and save —
// the config server picks it up, and the program reloads on the next poll:
//
//	sed -i '' 's/8443/8500/' config/orders-production.yml   # macOS
//	sed -i    's/8443/8500/' config/orders-production.yml   # linux
//
// Ctrl-C to quit, then `docker compose down`.
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
	"github.com/goreflect/gostructor/snapshot"
	"github.com/goreflect/gostructor/springcloud"
)

// Config is filled from the merged property sources returned by
// GET /orders/production. Keys are the flat, dotted Spring property names.
type Config struct {
	Host  string `cfg:"server.host,springcloud:server.host"`
	Port  int    `cfg:"server.port,springcloud:server.port"`
	Level string `cfg:"log.level,springcloud:log.level" gos:"default:info"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	store, err := snapshot.NewDirStore("./snapshot")
	if err != nil {
		panic(err)
	}

	src, err := springcloud.New(springcloud.Options{
		Address:     envOr("CONFIG_SERVER", "http://127.0.0.1:8888"),
		Application: "orders",
		Profile:     "production",
		Poll:        3 * time.Second,
		Snapshot:    store,
		Logger:      logger,
	})
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Println("watching Spring Cloud Config for orders/production (Ctrl-C to quit)")
	err = gostructor.Watch(ctx, &Config{}, func(cfg *Config, err error) {
		if err != nil {
			fmt.Printf("⚠️  reload rejected: %v\n", err)
			return
		}
		fmt.Printf("✅ %s:%d  level=%s  (version %s)\n", cfg.Host, cfg.Port, cfg.Level, short(src.Version()))
	},
		gostructor.WithSources(src, gostructor.Default()),
		gostructor.WithLogger(logger),
	)
	if err != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, "watch error:", err)
		os.Exit(1)
	}
}

func short(v string) string {
	if len(v) > 12 {
		return v[:12]
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
