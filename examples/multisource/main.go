// Command multisource shows a field with several possible sources (env, then a
// YAML file) and a validation hook. Priority is the WithSources order: Env is
// listed first, so an env var wins over the YAML file when both have a value.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/yaml"
)

type Config struct {
	Host string `cfg:"host,yaml:server.host"`
	Port int    `cfg:"port,yaml:server.port"`
}

func main() {
	os.Setenv(yaml.FileEnvVar, "config.yml")

	cfg, err := gostructor.Configure(&Config{},
		gostructor.WithSources(gostructor.Env(), yaml.New()),
		gostructor.WithHook(func(field gostructor.FieldContext, value any) (any, error) {
			if field.Name == "Port" && value.(int) < 1024 {
				return nil, fmt.Errorf("port %v is a privileged port", value)
			}
			return value, nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", cfg)
}
