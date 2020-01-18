package pipeline

import "fmt"

type JsonConfig struct {
}

func (json JsonConfig) Configure() error {
	fmt.Println("Level: Debug. Json configurator source start.")
	return nil
}
