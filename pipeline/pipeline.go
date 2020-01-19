package pipeline

import (
	"errors"
	"fmt"
)

type (
	// FuncType - identifier type for one of our configure functions
	FuncType int

	// Pipeline -
	Pipeline struct {
		chains *Chain
	}

	Chain struct {
		stageFunction IConfigure
		notAValues    []*NotAValue
		next          *Chain
	}
)

const (
	// FunctionSetupEnvironment - identifier function configuration your structure
	FunctionSetupEnvironment  = iota
	FunctionSetupHocon        = iota
	FunctionSetupJson         = iota
	FunctionSetupYaml         = iota
	FunctionSetupDefault      = iota
	FunctionSetupVault        = iota
	FunctionSetupConfigServer = iota

	notSupportedTypeError = "not supported type "
)

// get pipeline of functions in chain notation
func getFunctionChain(pipelineChanes []FuncType) *Pipeline {
	chain := &Chain{
		stageFunction: nil,
		next:          nil,
		notAValues:    nil,
	}
	currentChain := chain
	for _, pipelineChain := range pipelineChanes {
		stageFunction, err := getChainByIdentifier(pipelineChain)
		if err != nil {
			fmt.Println("error while getting chain stage function. Error: " + err.Error())
			continue
		}
		currentChain.stageFunction = stageFunction
		currentChain.next = new(Chain)
		currentChain = currentChain.next
	}
	return &Pipeline{chains: chain}
}

func getChainByIdentifier(
	// idFunc - identifier of function configuration source
	idFunc FuncType) (IConfigure, error) {
	switch idFunc {
	case FunctionSetupDefault:
		return &DefaultConfig{}, nil
	case FunctionSetupEnvironment:
		return &EnvironmentConfig{}, nil
	case FunctionSetupHocon:
		return &HoconConfig{}, nil
	case FunctionSetupJson:
		return nil, errors.New(notSupportedTypeError +
			"json configurator source. Not realized yet")
	case FunctionSetupYaml:
		return nil, errors.New(notSupportedTypeError +
			"yaml configurator source. Not realized yet")
	case FunctionSetupVault:
		return nil, errors.New(notSupportedTypeError + "vault configurator source. Not realized yet")
	case FunctionSetupConfigServer:
		return nil, errors.New(notSupportedTypeError + "configure server configurator source. Not realized yet")
	default:
		return nil, errors.New(notSupportedTypeError +
			"you should search in lib available type configurator source")
	}
}

// Configure - main configurer
func Configure(
	// Needed configure structure
	structure interface{},
	// filename for file configuring
	fileName string,
	pipelineChaines []FuncType) (err error) {

	defer reestablish()
	pipeline := getFunctionChain(pipelineChaines)
	currentChain := pipeline.chains

	for {
		fmt.Println("Level: Debug. current stage function.")
		if err := currentChain.stageFunction.Configure(); err != nil {
			fmt.Println("Level: Warning. can not chain function source setuping value. Error: ", err.Error())
		}
		if currentChain.next != nil {
			currentChain = currentChain.next
		} else {
			break
		}
	}
	return nil
}

func reestablish() error {
	if err := recover(); err != nil {
		return err.(error)
	}
	return nil
}
