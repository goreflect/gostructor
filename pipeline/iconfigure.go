package pipeline

type IConfigure interface {
	Configure() (bool, error)
}
