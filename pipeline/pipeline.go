package pipeline

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/goreflect/gostructor/tags"
)

type (
	// FuncType - identifier type for one of our configure functions
	FuncType int

	// Pipeline -
	Pipeline struct {
		chains       *Chain
		errors       []string
		fileName     string
		sourcesTypes []int
	}

	Chain struct {
		stageFunction IConfigure
		// middleWares   IMiddleware realize in #7 issue
		notAValues []*NotAValue
		next       *Chain
	}

	structContext struct {
		Value       reflect.Value
		StructField reflect.StructField
		Prefix      string
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

	sourceFileInDisk   = 0
	sourceFielInServer = 1
	sourceFileNotUsed  = 2

	notSupportedTypeError = "not supported type "

	EmptyAdditionalPrefix = ""
)

// bool - skip preview prefix or not
func (context structContext) getFieldName() (bool, string) {
	for _, val := range []string{tags.TagHocon, tags.TagJson, tags.TagYaml, tags.TagDefault, tags.TagEnvironment, tags.TagConfigServer, tags.TagHashiCorpVault} {
		tag := context.StructField.Tag.Get(val)
		if tag == "" {
			continue
		}
		tagCustomer := strings.Split(tag, "=")
		if len(tagCustomer) == 0 {
			continue
		}
		for _, tagCustom := range []string{tags.TagCustomerNode, tags.TagCustomerPath} {
			if tagCustomer[0] == tagCustom {
				if tagCustomer[1] != "" {
					return true, tagCustomer[1]
				}
			}
		}
		return false, tag
	}
	return false, context.StructField.Name
}

// get pipeline of functions in chain notation
func getFunctionChain(fileName string, pipelineChanes []FuncType) *Pipeline {
	chain := &Chain{
		stageFunction: nil,
		next:          nil,
		notAValues:    nil,
	}
	sourcesTypes := make([]int, 3)

	currentChain := chain
	for _, pipelineChain := range pipelineChanes {
		stageFunction, sourceType, err := getChainByIdentifier(pipelineChain, fileName)
		if err != nil {
			fmt.Println("[Pipeline]: level: debuf. error while getting chain stage function. Error: " + err.Error())
			continue
		}
		sourcesTypes[sourceType]++
		currentChain.stageFunction = stageFunction
		currentChain.next = new(Chain)
		currentChain = currentChain.next
	}
	return &Pipeline{chains: chain, sourcesTypes: sourcesTypes}
}

// getChainByIdentifier - get function configure source and return
// Configure func struct contain
// hasSourcesFile - if in configuration pipeline have source configuring from file
// doesntHaveSourceFile - if source is another source
func getChainByIdentifier(
	// idFunc - identifier of function configuration source
	idFunc FuncType,
	fileName string) (IConfigure, int, error) {
	switch idFunc {
	case FunctionSetupDefault:
		return &DefaultConfig{}, sourceFileNotUsed, nil
	case FunctionSetupEnvironment:
		return &EnvironmentConfig{}, sourceFileNotUsed, nil
	case FunctionSetupHocon:
		return &HoconConfig{fileName: fileName}, sourceFileInDisk, nil
	case FunctionSetupJson:
		return nil, sourceFileInDisk, errors.New(notSupportedTypeError +
			"json configurator source. Not realized yet")
	case FunctionSetupYaml:
		return nil, sourceFileInDisk, errors.New(notSupportedTypeError +
			"yaml configurator source. Not realized yet")
	case FunctionSetupVault:
		return nil, sourceFielInServer, errors.New(notSupportedTypeError + "vault configurator source. Not realized yet")
	case FunctionSetupConfigServer:
		return nil, sourceFielInServer, errors.New(notSupportedTypeError + "configure server configurator source. Not realized yet")
	default:
		return nil, sourceFileNotUsed, errors.New(notSupportedTypeError +
			"you should search in lib available type configurator source or you are welcome to contribute.")
	}
}

// Configure - main configurer
func Configure(
	// Needed configure structure
	structure interface{},
	// filename for file configuring
	fileName string,
	// functions will be configure structure
	pipelineChaines []FuncType,
	// prefix by getting data from source placed in entry
	prefix string) (err error) {

	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
			err = e.(error)
		}
	}()

	pipeline := getFunctionChain(fileName, pipelineChaines)
	// currentChain := pipeline.chains
	if err := pipeline.setFile(fileName); err != nil {
		if pipeline.checkSourcesConfigure() {
			fmt.Println("[Pipeline]: level: Warning. can not be access to file or server. ", err.Error())
		}
	}

	if err := pipeline.recursiveParseFields(&structContext{
		Value:  reflect.ValueOf(structure),
		Prefix: prefix,
	}); err != nil {
		fmt.Println("[Pipeline]: level: error. error while configuring your structure. Errors: ", err.Error())
	}
	return nil
}

// if the source is a file or server, a warning will be added if the resource is not available to receive data from it
func (pipeline *Pipeline) checkSourcesConfigure() bool {
	for source, amount := range pipeline.sourcesTypes {
		switch source {
		case sourceFileInDisk, sourceFielInServer:
			if amount > 0 {
				return true
			}
		case sourceFileNotUsed:
			continue
		}
	}
	return false
}

func (pipeline *Pipeline) setFile(fileName string) error {
	if fileName == "" {
		return errors.New("file with empty name can not be open")
	}

	if err := checkFileAccessibility(fileName); err != nil {
		return errors.New("can not open file. Error: " + err.Error())
	}
	return nil
}

func (pipeline *Pipeline) recursiveParseFields(context *structContext) error {
	if err := pipeline.checkValueTypeIsPointer(context.Value); err != nil {
		return err
	}
	valuePtr := reflect.Indirect(context.Value)
	switch valuePtr.Kind() {
	case reflect.Struct:
		newPrefix := valuePtr.Type().Name()
		if context.Prefix == "" {
			context.Prefix += newPrefix
		} else {
			context.Prefix += "." + newPrefix
		}
		for i := 0; i < valuePtr.NumField(); i++ {
			prefix := context.Prefix
			if err := pipeline.checkValuePrefix(context.Prefix); err == nil {
				prefix += "."
			}
			ok, getPrefix := (structContext{StructField: valuePtr.Type().Field(i)}.getFieldName())
			if ok {
				prefix = getPrefix
			} else {
				prefix += getPrefix
			}
			if err := pipeline.recursiveParseFields(&structContext{
				Value:       valuePtr.Field(i).Addr(),
				StructField: valuePtr.Type().Field(i),
				Prefix:      prefix,
			}); err != nil {
				pipeline.addNewErrorWhileParsing(err.Error())
			}
		}
	default:
		if err := pipeline.configuringValues(context); err != nil {
			pipeline.addNewErrorWhileParsing(err.Error())
		}
	}
	return pipeline.getErrorAsOne()
}

func (pipeline *Pipeline) addNewErrorWhileParsing(err string) {
	pipeline.errors = append(pipeline.errors, err)
}

func (pipeline *Pipeline) clearErrors() {
	pipeline.errors = []string{}
}

func (pipeline *Pipeline) getErrorAsOne() error {
	if len(pipeline.errors) > 0 {
		return errors.New("on stage recurisiveParseFields will have any of this errors: " + strings.Join(pipeline.errors, "\n"))
	} else {
		return nil
	}
}

// //environment values sourcesing configuratino your structure field by functions pipeline
// func (pipeline *Pipeline) startConfiguringByFunctions(context *structContext) error {
// 	currentChain := pipeline.chains
// 	for {
// 		if err := currentChain.stageFunction.Configure(context); err != nil {
// 			if currentChain.next != nil {
// 				currentChain = currentChain.next
// 			} else {
// 				return errors.New("can not configure field: " + context.Prefix)
// 			}
// 		} else {
// 			return nil
// 		}
// 	}
// }

func (pipeline *Pipeline) configuringValues(context *structContext) error {
	// if config.configureFileParsed == nil {
	// 	config.configureFileParsed = gohocon.LoadConfig(config.fileName)
	// }
	valueIndirect := reflect.Indirect(context.Value)
	switch valueIndirect.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		// valueIndirect := reflect.Indirect(context.Value)
		valueGet := pipeline.chains.stageFunction.GetComplexType(context)
		if valueGet.CheckIsValue() {
			fmt.Println("[Pipeline]: Level: debug. value get from parsing slice: ", valueGet)
			if valueIndirect.CanSet() {
				valueIndirect.Set(valueGet.Value)
			} else {
				return errors.New("can not set " + valueIndirect.Kind().String() + " into struct field.")
			}
		} else {
			return errors.New("Level: debug. value get not setupable value: ")
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return errors.New("not supported types of unsigned integer")
	case reflect.String, reflect.Float32, reflect.Float64, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valueGet := pipeline.chains.stageFunction.GetBaseType(context)
		if valueGet.CheckIsValue() {
			fmt.Println("[Pipeline]: Level: debug. value get from parsing "+valueGet.Value.Kind().String()+": ", valueGet)
			if valueIndirect.CanSet() {
				valueIndirect.Set(valueGet.Value)
			} else {
				return errors.New("can not set " + valueIndirect.Kind().String() + " into struct field.")
			}
		} else {
			return errors.New("Level: debug. value get not setupable value: ")
		}
	default:
		return errors.New("not supported type for hocon parsing")
	}
	return nil
}

func (pipeline *Pipeline) checkValueTypeIsPointer(value reflect.Value) error {
	if value.Kind() != reflect.Ptr {
		return errors.New("not working without pointer types")
	}
	return nil
}

func (pipeline *Pipeline) checkValuePrefix(prefix string) error {
	if prefix == "" || prefix[len(prefix)-1] == '.' || len(prefix) <= 1 {
		return errors.New("can not entry point")
	}
	return nil
}

func (context *structContext) conversion(send reflect.Value, recive reflect.Type) (convertResult *reflect.Value, err error) {
	defer func() {
		if errPanic := recover(); errPanic != nil {
			err = errPanic.(error)
			convertResult = nil
		}
	}()
	result := send.Convert(recive)
	return &result, nil
}
