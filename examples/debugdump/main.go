// Command debugdump demonstrates the debug dump: a running service that watches
// a local JSON file and exposes its live, effective configuration over a
// loopback TCP port — no HTTP, no redeploy needed to inspect it.
//
// Run it (this directory is its own module):
//
//	cd examples/debugdump && go run .
//
// Then, in another terminal, read the live config the service resolved — either
// with the client (from the repo root) or with nc:
//
//	go run ./cmd/gostructor-dump          # from repo root; connects to 127.0.0.1:6555
//	nc 127.0.0.1 6555
//
// Now edit config.json (change the port, break the JSON, set an out-of-range
// port) and save: the service reloads, and the next dump reflects the new state
// — while a rejected reload leaves the last-known-good dump serving. The secret
// field is always masked in the dump. Stop with Ctrl-C.
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
	// Token is secret: masked everywhere it prints, including the debug dump.
	Token string `cfg:"token,file:token" gos:"secret,optional"`
}

const (
	configFile = "config.json"
	dumpAddr   = "127.0.0.1:6555"
)

func main() {
	// Seed a config file on first run so the example is self-contained.
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		seed := `{
  "server": { "host": "127.0.0.1", "port": 8080 },
  "message": "edit me and save",
  "token": "super-secret-token"
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

	fmt.Printf("watching %s — inspect the live config with:\n", configFile)
	fmt.Printf("  go run ./cmd/gostructor-dump   (or: nc %s)\n", dumpAddr)
	fmt.Println("edit config.json and save to see the dump refresh (Ctrl-C to quit)")

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
		// Expose the live, effective config on a loopback port. The listener is
		// closed automatically when ctx is cancelled (Ctrl-C).
		gostructor.WithDebugDump(gostructor.DumpTCP(dumpAddr)),
	)
	if err != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, "watch error:", err)
		os.Exit(1)
	}
	if cfg := current.Load(); cfg != nil {
		fmt.Printf("\nfinal config: %s:%d\n", cfg.Host, cfg.Port)
	}
	fmt.Println("bye")
}
