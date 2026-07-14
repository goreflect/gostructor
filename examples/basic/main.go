// Command basic shows the minimum gostructor setup: env vars with
// cf_default fallbacks, using only the core module.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/goreflect/gostructor"
)

type Config struct {
	Host  string `cf_env:"APP_HOST" cf_default:"0.0.0.0"`
	Port  int    `cf_env:"APP_PORT" cf_default:"8080"`
	Debug bool   `cf_env:"APP_DEBUG" cf_default:"false"`
}

func main() {
	os.Setenv("APP_PORT", "9090")

	cfg, err := gostructor.Configure(&Config{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", cfg)
}
