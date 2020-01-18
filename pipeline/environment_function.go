package pipeline

import "fmt"

type EnvironmentConfig struct {
}

func (config EnvironmentConfig) Configure() error {
	fmt.Println("Level: Debug. environment values sources start")

	return nil
}
