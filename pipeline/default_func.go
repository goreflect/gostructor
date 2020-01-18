package pipeline

import "fmt"

type DefaultConfig struct {
}

func (config DefaultConfig) Configure() error {
	fmt.Println("Level: Debug. default values sources start")
	return nil
}
