// Command priority shows gostructor's headline feature: the SAME struct
// resolving DIFFERENTLY depending on a per-field source order chosen at
// runtime. In prod the environment wins; in dev the baked-in default wins -
// all from one cf_priority tag and the GOSTRUCTOR_PRIORITY selector, no code
// branches.
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
	// prod: try cf_env first (an operator override), fall back to default.
	// dev:  ignore any stray env var and pin the safe local default first.
	LogLevel string `cf_env:"LOG_LEVEL" cf_default:"info" cf_priority:"prod:cf_env,cf_default;dev:cf_default,cf_env"`
	// Same idea for the sampling rate the two environments want differently.
	TraceRate string `cf_env:"TRACE_RATE" cf_default:"0.01" cf_priority:"prod:cf_env,cf_default;dev:cf_default,cf_env"`
}

func main() {
	// An operator sets these in the real environment; in dev we want them
	// deliberately ignored in favour of the local defaults.
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("TRACE_RATE", "1.0")

	for _, env := range []string{"prod", "dev"} {
		os.Setenv(gostructor.PriorityEnvVar, env) // GOSTRUCTOR_PRIORITY=prod|dev

		cfg, err := gostructor.Configure(&Config{})
		if err != nil {
			fmt.Println("configure failed:", err)
			os.Exit(1)
		}
		fmt.Printf("[%-4s] LogLevel=%-5s TraceRate=%s\n", env, cfg.LogLevel, cfg.TraceRate)
	}

	fmt.Println("\nSame struct, same env vars — only GOSTRUCTOR_PRIORITY changed:")
	fmt.Println("  prod trusts the operator's env override; dev pins the local default.")
}
