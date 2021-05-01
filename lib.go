package gostructor

import (
	"os"

	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/pipeline"
	logrus "github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.ErrorLevel)
}

/*ChangeLogLevel - changing current loggin level*/
func ChangeLogLevel(logLevel logrus.Level) {
	logrus.SetLevel(logLevel)
}

/*ChangeLogFormatter - changing current formatter*/
func ChangeLogFormatter(formatter logrus.Formatter) {
	logrus.SetFormatter(formatter)
}

/*
ConfigureEasy - default pipeline setup for configure your structure
*/
func ConfigureEasy(
	structure interface{}) (interface{}, error) {
	return pipeline.Configure(structure, []infra.FuncType{
		infra.FunctionSetupEnvironment,
		infra.FunctionSetupHocon,
		infra.FunctionSetupDefault,
	}, pipeline.EmptyAdditionalPrefix, pipeline.DirtyConfiguring)
}

/*
ConfigureSetup - pipeline with your settings stages for your structure
*/
func ConfigureSetup(
	structure interface{},
	prefix string,
	functions []infra.FuncType) (interface{}, error) {
	return pipeline.Configure(structure, functions, prefix, pipeline.DirtyConfiguring)
}

/*
ConfigureSmart - configuring by analysing tags for add prefer strategy for configuring
*/
func ConfigureSmart(
	structure interface{},
) (interface{}, error) {
	return pipeline.Configure(structure, nil, pipeline.EmptyAdditionalPrefix, pipeline.SmartConfiguring)
}
