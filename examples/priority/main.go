// Command priority shows how gostructor expresses priority: the SAME struct
// resolves DIFFERENTLY purely from the ORDER of sources passed to WithSources.
// There is no priority tag and no global selector - whoever is listed first
// and reports a value wins. In "prod" the operator's env override leads; in
// "dev" the baked-in default leads and any stray env var is ignored.
//
// Run it:
//
//	go run ./examples/priority
package main

import (
	"fmt"
	"os"

	"github.com/goreflect/gostructor"
)

type Config struct {
	// Env name derives from the base ("logLevel" -> LOG_LEVEL); the default
	// lives on the gos tag. Which one wins is decided by source order below.
	LogLevel  string `cfg:"logLevel" gos:"default:info"`
	TraceRate string `cfg:"traceRate" gos:"default:0.01"`
}

func main() {
	// An operator sets these in the real environment; in dev we want them
	// deliberately ignored in favour of the local defaults.
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("TRACE_RATE", "1.0")

	orders := map[string][]gostructor.Source{
		// prod: env override leads, default is the fallback.
		"prod": {gostructor.Env(), gostructor.Default()},
		// dev: default leads, so the stray env vars are ignored.
		"dev": {gostructor.Default(), gostructor.Env()},
	}

	for _, env := range []string{"prod", "dev"} {
		cfg, err := gostructor.Configure(&Config{},
			gostructor.WithSources(orders[env]...))
		if err != nil {
			fmt.Println("configure failed:", err)
			os.Exit(1)
		}
		fmt.Printf("[%-4s] LogLevel=%-5s TraceRate=%s\n", env, cfg.LogLevel, cfg.TraceRate)
	}

	fmt.Println("\nSame struct, same env vars — only the WithSources order changed:")
	fmt.Println("  prod trusts the operator's env override; dev pins the local default.")
}
