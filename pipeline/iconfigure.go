package pipeline

// IConfigure - configurer interface for chain pipeline configuration
type IConfigure interface {
	Configure() error
}
