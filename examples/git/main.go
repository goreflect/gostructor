// Command git-example drives configuration from a git repository: it reads a
// JSON config file at a ref (branch/tag = version), follows the ref for new
// commits (drift), and switches version at runtime.
//
//	docker compose up -d
//	go run .
//
// The repo is seeded with two versions: branch `main` and tag `v1`. The program
// starts on `main`. Send SIGHUP to toggle between `main` and `v1` at runtime and
// watch the config reload from a different git ref:
//
//	kill -HUP $(pgrep -f 'git-example|go-build.*git')   # or note the PID it prints
//
// Drift: a new commit pushed to `main` is picked up automatically within the
// poll interval. Ctrl-C to quit, then `docker compose down`.
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
	"github.com/goreflect/gostructor/git"
	"github.com/goreflect/gostructor/snapshot"
)

// Config is filled from services/api/config.json in the git repo.
type Config struct {
	Host    string `cfg:"host,git:server.host"`
	Port    int    `cfg:"port,git:server.port"`
	Release string `cfg:"release,git:release"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	store, err := snapshot.NewDirStore("./snapshot")
	if err != nil {
		panic(err)
	}

	src, err := git.New(git.Options{
		Repo:     envOr("GIT_REPO", "git://127.0.0.1:9418/config.git"),
		Ref:      "main",
		Path:     "services/api/config.json",
		Poll:     5 * time.Second, // drift check interval
		Snapshot: store,
		Logger:   logger,
	})
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Toggle the served version between main and v1 on SIGHUP.
	var onMain atomic.Bool
	onMain.Store(true)
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	go func() {
		for range hup {
			next := "v1"
			if !onMain.Load() {
				next = "main"
			}
			fmt.Printf("↪  switching to ref %q\n", next)
			if err := src.SetVersion(next); err != nil {
				fmt.Printf("⚠️  switch failed: %v\n", err)
				continue
			}
			onMain.Store(!onMain.Load())
		}
	}()

	fmt.Printf("pid %d — reading git config at ref main; `kill -HUP %d` to toggle main/v1 (Ctrl-C to quit)\n", os.Getpid(), os.Getpid())
	err = gostructor.Watch(ctx, &Config{}, func(cfg *Config, err error) {
		if err != nil {
			fmt.Printf("⚠️  reload rejected: %v\n", err)
			return
		}
		fmt.Printf("✅ %s:%d  release=%s  (commit %s)\n", cfg.Host, cfg.Port, cfg.Release, short(src.Version()))
	},
		gostructor.WithSources(src, gostructor.Default()),
		gostructor.WithLogger(logger),
	)
	if err != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, "watch error:", err)
		os.Exit(1)
	}
}

func short(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
