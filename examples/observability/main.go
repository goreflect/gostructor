// Command observability shows gostructor's resolution trace: which source
// won each field, which ones lost and why, and how cf_secret fields stay
// masked in the report. Uses only the core module.
//
// Run it:
//
//	go run ./examples/observability
package main

import (
	"fmt"
	"os"

	"github.com/goreflect/gostructor"
)

type Config struct {
	// Port comes from the env var here, so cf_default is never tried.
	Port int `cf_env:"APP_PORT" cf_default:"8080"`
	// Host has no env var set below, so it falls through to cf_default.
	Host string `cf_env:"APP_HOST" cf_default:"0.0.0.0"`
	// APIKey is marked secret: its value is masked everywhere it prints.
	APIKey string `cf_env:"APP_API_KEY" cf_secret:""`
}

func main() {
	os.Setenv("APP_PORT", "9090")
	os.Setenv("APP_API_KEY", "super-secret-token")
	// APP_HOST intentionally left unset to show the fall-through to default.

	cfg, report, err := gostructor.ConfigureWithReport(&Config{},
		// Reveal only the last four characters of secrets in the trace.
		gostructor.WithMasker(func(_ gostructor.FieldContext, v any) string {
			s, _ := v.(string)
			if len(s) >= 4 {
				return "••••" + s[len(s)-4:]
			}
			return "••••"
		}),
	)
	if err != nil {
		fmt.Println("configure failed:", err)
		os.Exit(1)
	}

	fmt.Printf("resolved config: Port=%d Host=%s APIKey=%q\n\n", cfg.Port, cfg.Host, cfg.APIKey)

	fmt.Println("resolution trace:")
	fmt.Println(report.String())

	fmt.Println("\nprovenance (field -> winning source):")
	for field, source := range report.Provenance() {
		fmt.Printf("  %-8s %s\n", field, source)
	}
}
