// Command filesources shows the two core file sources - JSON and INI - used
// together on one struct, with env and cf_default in the same priority chain.
// Each field declares several sources; the first that resolves wins. No
// external modules required.
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
	// From the JSON file (nested key), overridable by env.
	Host string `cf_env:"APP_HOST" cf_json:"server.host" cf_default:"127.0.0.1"`
	// From the INI file (section#key).
	Port int `cf_ini:"server#port" cf_default:"8080"`
	// Not present in either file → falls through to the default.
	MaxConns int `cf_json:"server.maxConns" cf_ini:"server#max_conns" cf_default:"256"`
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
	fmt.Printf("MaxConns %-10d (from cf_default — absent in both files)\n", cfg.MaxConns)
}
