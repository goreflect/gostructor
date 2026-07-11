package infra

// FuncType - identifier type for one of our configure functions
type FuncType int

// Values are assigned explicitly (not via iota) because they double as slice
// indices (see tags.AmountTags and tags.GetFunctionTypes) - an implicit,
// position-dependent numbering would silently renumber every constant below
// it whenever a new one is inserted.
const (
	// FunctionNotExist - not founded function for configuring
	FunctionNotExist FuncType = -1
	// FunctionSetupEnvironment - identifier function configuration your structure
	FunctionSetupEnvironment FuncType = 1
	// FunctionSetupHocon - function configuring from hocon
	FunctionSetupHocon FuncType = 2
	// FunctionSetupJSON - function configuring from json
	FunctionSetupJSON FuncType = 3
	// FunctionSetupYaml - function configuring from yaml
	FunctionSetupYaml FuncType = 4
	// FunctionSetupIni - function configuring from ini
	FunctionSetupIni FuncType = 5
	// FunctionSetupToml - function configuring from toml
	FunctionSetupToml FuncType = 6
	// FunctionSetupDefault - function configuring from default values
	FunctionSetupDefault FuncType = 7
	// FunctionSetupVault - function configuring from vault secured backend
	FunctionSetupVault FuncType = 8
	// FunctionSetupConfigServer - function configuring by any types (json, yaml, hocon, toml, txt...) from configuring server with settings
	FunctionSetupConfigServer FuncType = 9
	// FunctionKeyValueServer - function configuring from a key/value store backend
	FunctionKeyValueServer FuncType = 10
)
