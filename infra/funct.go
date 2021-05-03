package infra

// FuncType - identifier type for one of our configure functions
type FuncType int

const (
	// FunctionNotExist - not founded function for configuring
	FunctionNotExist = -1
	// FunctionSetupEnvironment - identifier function configuration your structure
	FunctionSetupEnvironment = iota
	// FunctionSetupHocon - function configuring from hocon
	FunctionSetupHocon = iota
	// FunctionSetupJSON - function configuring from json
	FunctionSetupJSON = iota
	// FunctionSetupYaml - function configuring from yaml
	FunctionSetupYaml = iota
	FunctionSetupIni  = iota
	FunctionSetupToml = iota
	// FunctionSetupDefault - function configuring from default values
	FunctionSetupDefault = iota
	// FunctionSetupVault - function configuring from vault secured backend
	FunctionSetupVault = iota
	// FunctionSetupConfigServer - function configuring by any types (json, yaml, hocon, toml, txt...) from configuring server with settings
	FunctionSetupConfigServer = iota

	FunctionKeyValueServer = iota
)
