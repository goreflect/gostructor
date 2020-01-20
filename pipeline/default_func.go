package pipeline

import "fmt"

type DefaultConfig struct {
}

func (config DefaultConfig) Configure(context *structContext) error {
	fmt.Println("Level: Debug. default values sources start")
	return nil
}
