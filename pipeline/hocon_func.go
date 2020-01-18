package pipeline

import "fmt"

type HoconConfig struct {
	Next IConfigure
}

func (config HoconConfig) Configure() error {
	fmt.Println("Level: Debug. hocon values sources start")

	return nil
}
