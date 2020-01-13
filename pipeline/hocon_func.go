package pipeline

type HoconConfig struct {
	Next IConfigure
}

func (hocon *HoconConfig) Configure() (bool, error) {
	return false, nil
}
