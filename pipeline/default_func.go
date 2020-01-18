package pipeline

import "fmt"

type DefaultConfig struct {
	next IConfigure
}

func (config DefaultConfig) Configure() error {
	fmt.Println("Level: Debug. default values sources start")
	return nil
}
