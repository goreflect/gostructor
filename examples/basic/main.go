// Command basic shows the minimum gostructor setup: env vars (named from each
// field via the cfg tag) with gos default fallbacks, using only the core
// module. With no env var set, the default wins; set one and it takes over.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/goreflect/gostructor"
)

type Config struct {
	Host  string `cfg:"host" gos:"default:0.0.0.0"`
	Port  int    `cfg:"port" gos:"default:8080"`
	Debug bool   `cfg:"debug" gos:"default:false"`
}

func main() {
	// The env source names the variable from the field's base name in
	// SCREAMING_SNAKE_CASE: "port" -> PORT. No per-field env name needed.
	os.Setenv("PORT", "9090")

	cfg, err := gostructor.Configure(&Config{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", cfg)
}
