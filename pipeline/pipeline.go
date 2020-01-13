package pipeline

type FuncType int

const (
	FunctionSetupHocon       = iota
	FunctionSetupJson        = iota
	FunctionSetupYaml        = iota
	FunctionSetupDefault     = iota
	FunctionSetupEnvironment = iota
	FunctionSetupValidation  = iota
)
