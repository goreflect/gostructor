package gostructor

import (
	"github.com/goreflect/gostructor/pipeline"
)

/*
ConfigureEasy - default pipeline setup for configure your structure
*/
func ConfigureEasy(
	structure interface{},
	fileName string) (interface{}, error) {
	return pipeline.Configure(structure, fileName, []pipeline.FuncType{
		pipeline.FunctionSetupEnvironment,
		pipeline.FunctionSetupHocon,
		pipeline.FunctionSetupDefault,
	}, pipeline.EmptyAdditionalPrefix)
}

/*
ConfigureSetup - pipeline with your settings stages for your structure
*/
func ConfigureSetup(
	structure interface{},
	fileName string,
	prefix string,
	functions []pipeline.FuncType) (interface{}, error) {
	return pipeline.Configure(structure, fileName, functions, prefix)
}
