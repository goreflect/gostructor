// Command hooks shows WithHook for two jobs: transforming a resolved value
// (normalise a string) and validating one (reject an out-of-range port),
// both operating on the already-converted, field-typed value.
//
// Run it:
//
//	go run ./examples/hooks
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/goreflect/gostructor"
)

type Config struct {
	Env  string `cfg:"env" gos:"default:Production"`
	Port int    `cfg:"port" gos:"default:8080"`
}

func main() {
	os.Setenv("ENV", "  Staging  ") // messy input to be normalised
	os.Setenv("PORT", "9090")

	cfg, err := gostructor.Configure(&Config{},
		// Transform: trim + lowercase every string field.
		gostructor.WithHook(func(f gostructor.FieldContext, v any) (any, error) {
			if s, ok := v.(string); ok {
				return strings.ToLower(strings.TrimSpace(s)), nil
			}
			return v, nil
		}),
		// Validate: a hook sees the typed value (an int here, not "9090"),
		// so range checks are natural. Return an error to reject it.
		gostructor.WithHook(func(f gostructor.FieldContext, v any) (any, error) {
			if f.Name == "Port" {
				if p := v.(int); p < 1024 || p > 65535 {
					return nil, fmt.Errorf("port %d out of range 1024-65535", p)
				}
			}
			return v, nil
		}),
	)
	if err != nil {
		fmt.Println("configure failed:", err)
		os.Exit(1)
	}
	fmt.Printf("normalised + validated: Env=%q Port=%d\n", cfg.Env, cfg.Port)

	// Now show the validation hook rejecting a bad value.
	os.Setenv("PORT", "80") // privileged, below 1024
	if _, err := gostructor.Configure(&Config{},
		gostructor.WithHook(func(f gostructor.FieldContext, v any) (any, error) {
			if f.Name == "Port" && v.(int) < 1024 {
				return nil, fmt.Errorf("port %d is privileged", v)
			}
			return v, nil
		}),
	); err != nil {
		fmt.Println("rejected as expected:", err)
	}
}
