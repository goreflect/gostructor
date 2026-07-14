// Command observability shows gostructor's focused resolution trace: a summary
// line, the primary source, the defaults count, and an Overrides & Secrets
// section listing only the fields that were overridden away from the primary
// config or that are secret (masked). Uses only the core module.
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
	// Host and MaxConns come from the JSON file (the primary source).
	Host     string `cfg:"host,json:server.host"`
	MaxConns int    `cfg:"maxConns,json:server.maxConns"`
	// Port comes from the env var, overriding the JSON value: an override.
	Port int `cfg:"port,json:server.port"`
	// APIKey is secret: its value is masked everywhere it prints.
	APIKey string `cfg:"apiKey,env:APP_API_KEY" gos:"secret"`
}

const configJSON = `{"server": {"host": "0.0.0.0", "maxConns": 256, "port": 8080}}`

func main() {
	path := os.TempDir() + "/gostructor-observability.json"
	if err := os.WriteFile(path, []byte(configJSON), 0o644); err != nil {
		fmt.Println("write config:", err)
		os.Exit(1)
	}
	defer os.Remove(path)

	// PORT overrides the JSON server.port; APP_API_KEY supplies the secret.
	os.Setenv("PORT", "9090")
	os.Setenv("APP_API_KEY", "super-secret-token")

	cfg, report, err := gostructor.ConfigureWithReport(&Config{},
		// Env wins over JSON where both have a value (so PORT overrides
		// server.port); JSON is the primary file source behind it.
		gostructor.WithSources(gostructor.Env(), gostructor.JSONFile(path)),
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
