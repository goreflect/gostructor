// Command filesources shows the two core file sources, JSON and INI, used
// together on one struct with env and a gos default in the same priority
// chain. Priority is the WithSources order: the first source that resolves a
// field wins. No external modules required.
//
// Run it:
//
//	go run ./examples/filesources
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goreflect/gostructor"
)

type Config struct {
	// From the JSON file (nested key), overridable by env (HOST).
	Host string `cfg:"host,json:server.host" gos:"default:127.0.0.1"`
	// From the INI file (section#key).
	Port int `cfg:"port,ini:server#port" gos:"default:8080"`
	// Not present in either file → falls through to the default.
	MaxConns int `cfg:"maxConns,json:server.maxConns,ini:server#max_conns" gos:"default:256"`
}

const jsonFile = `{"server": {"host": "0.0.0.0"}}`

const iniFile = `[server]
port = 9090
`

func main() {
	dir := os.TempDir()
	jsonPath := filepath.Join(dir, "gostructor-fs.json")
	iniPath := filepath.Join(dir, "gostructor-fs.ini")
	_ = os.WriteFile(jsonPath, []byte(jsonFile), 0o644)
	_ = os.WriteFile(iniPath, []byte(iniFile), 0o644)
	defer os.Remove(jsonPath)
	defer os.Remove(iniPath)

	cfg, err := gostructor.Configure(&Config{},
		gostructor.WithSources(
			gostructor.Env(),
			gostructor.JSONFile(jsonPath),
			gostructor.INIFile(iniPath),
			gostructor.Default(),
		),
	)
	if err != nil {
		fmt.Println("configure failed:", err)
		os.Exit(1)
	}

	fmt.Printf("Host     %-10s (from JSON file)\n", cfg.Host)
	fmt.Printf("Port     %-10d (from INI file)\n", cfg.Port)
	fmt.Printf("MaxConns %-10d (from gos default — absent in both files)\n", cfg.MaxConns)
}
