package gostructor

import (
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/pipeline"
)

/*
ConfigureEasy - default pipeline setup for configure your structure
*/
func ConfigureEasy(
	structure interface{},
	fileName string) (interface{}, error) {
	return pipeline.Configure(structure, fileName, []infra.FuncType{
		infra.FunctionSetupEnvironment,
		infra.FunctionSetupHocon,
		infra.FunctionSetupDefault,
<<<<<<< HEAD
	}, pipeline.EmptyAdditionalPrefix, pipeline.DirtyConfiguring)
=======
	}, pipeline.EmptyAdditionalPrefix, pipeline.DurtyConfiguring)
>>>>>>> add logic for getting information about tags and fixture
}

/*
ConfigureSetup - pipeline with your settings stages for your structure
*/
func ConfigureSetup(
	structure interface{},
	fileName string,
	prefix string,
	functions []infra.FuncType) (interface{}, error) {
<<<<<<< HEAD
	return pipeline.Configure(structure, fileName, functions, prefix, pipeline.DirtyConfiguring)
=======
	return pipeline.Configure(structure, fileName, functions, prefix, pipeline.DurtyConfiguring)
>>>>>>> add logic for getting information about tags and fixture
}

/*
ConfigureSmart - configuring by analysing tags for add prefer strategy for configuring
*/
func ConfigureSmart(
	structure interface{},
	fileName string,
) (interface{}, error) {
	return pipeline.Configure(structure, fileName, nil, pipeline.EmptyAdditionalPrefix, pipeline.SmartConfiguring)
}
