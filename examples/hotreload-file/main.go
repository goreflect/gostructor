// Command hotreload-file demonstrates gostructor.Watch with a file source that
// reloads on change — no external services, just a local JSON file you edit.
//
// Run it, then edit config.json in another terminal and save: the program
// re-fills a fresh struct, validates it, and prints the new value. A bad edit
// (invalid JSON, or port out of range) is reported and rejected — the last good
// config keeps serving. Stop with Ctrl-C.
//
//	go run .
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/watch"
)

// Config is the service configuration filled from config.json.
type Config struct {
	Host    string `cfg:"host,file:server.host" gos:"default:0.0.0.0"`
	Port    int    `cfg:"port,file:server.port" gos:"default:8080"`
	Message string `cfg:"message,file:message" gos:"default:hello"`
}

const configFile = "config.json"

func main() {
	// Seed a config file on first run so the example is self-contained.
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		seed := `{
  "server": { "host": "127.0.0.1", "port": 8080 },
  "message": "edit me and save"
}
`
		if err := os.WriteFile(configFile, []byte(seed), 0o644); err != nil {
			panic(err)
		}
		fmt.Printf("wrote %s — edit it while this runs\n", configFile)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	src, err := watch.JSONFile(configFile)
	if err != nil {
		panic(err)
	}

	var current atomic.Pointer[Config]

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Println("watching", configFile, "— edit and save to see a live reload (Ctrl-C to quit)")

	err = gostructor.Watch(ctx, &Config{}, func(cfg *Config, err error) {
		if err != nil {
			fmt.Printf("⚠️  reload rejected, keeping last good config: %v\n", err)
			return
		}
		current.Store(cfg)
		fmt.Printf("✅ config now: %s:%d  message=%q\n", cfg.Host, cfg.Port, cfg.Message)
	},
		gostructor.WithSources(src, gostructor.Default()),
		gostructor.WithLogger(logger),
		gostructor.WithDebounce(200*time.Millisecond), // collapse an editor's multi-write save
		gostructor.WithValidate(func(c *Config) error {
			if c.Port < 1 || c.Port > 65535 {
				return fmt.Errorf("port %d out of range 1..65535", c.Port)
			}
			return nil
		}),
	)
	if err != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, "watch error:", err)
		os.Exit(1)
	}
	// A real service's request handlers would read current.Load() to serve on
	// the latest good config; here we just report the final one.
	if cfg := current.Load(); cfg != nil {
		fmt.Printf("\nfinal config: %s:%d\n", cfg.Host, cfg.Port)
	}
	fmt.Println("bye")
}
