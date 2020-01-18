package pipeline

import "fmt"

type YamlConfig struct {
}

func (yaml YamlConfig) Configure() error {
	fmt.Println("Level: Debug. Yaml configurator source run")
	return nil
}
