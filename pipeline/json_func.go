package pipeline

import "fmt"

type JsonConfig struct {
}

func (json JsonConfig) Configure(context *structContext) error {
	fmt.Println("Level: Debug. Json configurator source start.")
	return nil
}
