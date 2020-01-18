package pipeline

import "fmt"

type HoconConfig struct {
}

func (config HoconConfig) Configure() error {
	fmt.Println("Level: Debug. hocon values sources start")

	return nil
}
