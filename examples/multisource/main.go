// Command multisource shows a field with several possible sources (env,
// then a YAML file), a cf_priority override, and a validation hook.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/yaml"
)

type Config struct {
	Host string `cf_env:"APP_HOST" cf_yaml:"server.host"`
	Port int    `cf_env:"APP_PORT" cf_yaml:"server.port" cf_priority:"prod:cf_env,cf_yaml;dev:cf_yaml,cf_env"`
}

func main() {
	os.Setenv(yaml.FileEnvVar, "config.yml")
	os.Setenv(gostructor.PriorityEnvVar, "dev")

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
