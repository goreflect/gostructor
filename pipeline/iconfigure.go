package pipeline

// IConfigure - configurer interface for chain pipeline configuration
type IConfigure interface {
	GetComplexType(*structContext) GoStructorValue
	GetBaseType(*structContext) GoStructorValue
}
