package infra

// FuncType - identifier type for one of our configure functions
type FuncType int

const (
	FunctionNotExist = -1
	// FunctionSetupEnvironment - identifier function configuration your structure
	FunctionSetupEnvironment  = iota
	FunctionSetupHocon        = iota
	FunctionSetupJson         = iota
	FunctionSetupYaml         = iota
	FunctionSetupDefault      = iota
	FunctionSetupVault        = iota
	FunctionSetupConfigServer = iota
)
