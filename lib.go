package gostructor

import (
	"github.com/goreflect/gostructor/pipeline"
)

// ConfigureEasy - default pipeline setup for configure your structure
func ConfigureEasy(
	structure interface{},
	fileName string) error {
	return pipeline.Configure(structure, fileName, []pipeline.FuncType{
		pipeline.FunctionSetupEnvironment,
		pipeline.FunctionSetupHocon,
		pipeline.FunctionSetupDefault,
	})
}

// ConfigureSetup - pipeline with your settings stages for your structure
func ConfigureSetup(
	structure interface{},
	fileName string,
	functions []pipeline.FuncType) error {
	return pipeline.Configure(structure, fileName, functions)
}
