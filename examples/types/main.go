// Command types shows the breadth of field types gostructor fills from a
// single structured source (a JSON file here): durations, slices and arrays,
// TextUnmarshaler types (time.Time, net.IP), named scalar types, pointers,
// maps, and slices/maps of structs - all with strict, lossless conversion.
//
// Run it:
//
//	go run ./examples/types
package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/goreflect/gostructor"
)

// LogLevel is a named scalar type: it stays LogLevel, not plain string.
type LogLevel string

// Backend is filled as a struct element by matching JSON object keys to
// exported field names (case-insensitively) - no per-field tags needed.
type Backend struct {
	URL    string
	Weight int
}

type Config struct {
	Timeout   time.Duration      `cf_json:"timeout"`   // "1h30m" duration string
	Retries   []int              `cf_json:"retries"`   // slice
	Coords    [2]float64         `cf_json:"coords"`    // fixed-size array
	StartedAt time.Time          `cf_json:"startedAt"` // TextUnmarshaler
	BindIP    net.IP             `cf_json:"bindIp"`    // TextUnmarshaler
	Level     LogLevel           `cf_json:"level"`     // named scalar type
	MaxConns  *int               `cf_json:"maxConns"`  // pointer, allocated + set
	Backends  []Backend          `cf_json:"backends"`  // []Struct
	Shards    map[string]Backend `cf_json:"shards"`    // map[string]Struct
	Labels    map[string]string  `cf_json:"labels"`    // map[string]string
}

const configJSON = `{
  "timeout":   "1h30m",
  "retries":   [1, 2, 3],
  "coords":    [51.5074, -0.1278],
  "startedAt": "2024-01-02T15:04:05Z",
  "bindIp":    "10.0.0.42",
  "level":     "debug",
  "maxConns":  100,
  "backends":  [{"url": "eu-1", "weight": 5}, {"url": "eu-2", "weight": 3}],
  "shards":    {"payments": {"url": "shard-a", "weight": 1}},
  "labels":    {"team": "platform", "tier": "gold"}
}`

func main() {
	path := filepath.Join(os.TempDir(), "gostructor-types.json")
	if err := os.WriteFile(path, []byte(configJSON), 0o644); err != nil {
		fmt.Println("write config:", err)
		os.Exit(1)
	}
	defer os.Remove(path)

	cfg, err := gostructor.Configure(&Config{},
		gostructor.WithSources(gostructor.JSONFile(path)))
	if err != nil {
		fmt.Println("configure failed:", err)
		os.Exit(1)
	}

	fmt.Printf("Timeout   %v (%T)\n", cfg.Timeout, cfg.Timeout)
	fmt.Printf("Retries   %v\n", cfg.Retries)
	fmt.Printf("Coords    %v\n", cfg.Coords)
	fmt.Printf("StartedAt %s\n", cfg.StartedAt.Format(time.RFC1123))
	fmt.Printf("BindIP    %s\n", cfg.BindIP)
	fmt.Printf("Level     %q (%T)\n", cfg.Level, cfg.Level)
	fmt.Printf("MaxConns  %d (via *int)\n", *cfg.MaxConns)
	fmt.Printf("Backends  %+v\n", cfg.Backends)
	fmt.Printf("Shards    %+v\n", cfg.Shards)
	fmt.Printf("Labels    %v\n", cfg.Labels)
}
