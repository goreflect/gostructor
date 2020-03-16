package pipeline

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
)

type (
	// Pipeline -
	Pipeline struct {
		chains       *Chain
		errors       []string
		sourcesTypes []int
	}

	/*Chain - this is structure contain information for executing function by ordering
	 */
	Chain struct {
		stageFunction IConfigure
		// middleWares   IMiddleware realize in #7 issue
		notAValues []*infra.NotAValue
		next       *Chain
	}

	/*structContext - context which using by library functions for configuring and preparing operations
	 */
	structContext struct {
		Value       reflect.Value
		StructField reflect.StructField
		Prefix      string
	}
)

const (
<<<<<<< HEAD
<<<<<<< HEAD
	/*SmartConfiguring - flag which inform library need for start analyzing all tags in derived structure for setting function types of configuring structure
	 */
	SmartConfiguring = true
	/*DirtyConfiguring - flag*/
	DirtyConfiguring = false

=======
>>>>>>> start writing
=======
	SmartConfiguring = true
	DurtyConfiguring = false

>>>>>>> add logic for getting information about tags and fixture
	sourceFileInDisk   = 0
	sourceFielInServer = 1
	sourceFileNotUsed  = 2

	notSupportedTypeError = "not supported type "

	/*EmptyAdditionalPrefix - prefix for setup before all values can be empty*/
	EmptyAdditionalPrefix = ""
)

// bool - skip preview prefix or not (if user setup node it is return true, if user setup path that's return false)
// TODO: add error interface while user setup not knowing type of tags. And maybe if user make a mistake (this case need for autocorrection or autofill)
func (context structContext) getFieldName() (bool, string) {
	for _, val := range []string{
		tags.TagHocon,
		tags.TagJson,
		tags.TagYaml,
		tags.TagDefault,
		tags.TagEnvironment,
		tags.TagConfigServer,
		tags.TagHashiCorpVault} {
		tag := context.StructField.Tag.Get(val)
		if tag == "" {
			continue
		}
		tagCustomer := strings.Split(tag, "=")
		if len(tagCustomer) == 0 {
			continue
		}
		// TODO: change this case by add for range case checking tag custom (if user setup node and path, it's should be return node.path way)
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
func getFunctionChain(fileName string, pipelineChanes []infra.FuncType) *Pipeline {
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
			fmt.Println("[Pipeline]: level: debug. error while getting chain stage function. Error: " + err.Error())
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
	idFunc infra.FuncType,
	fileName string) (IConfigure, int, error) {
	switch idFunc {
	case infra.FunctionSetupDefault:
		return &DefaultConfig{}, sourceFileNotUsed, nil
	case infra.FunctionSetupEnvironment:
		return &EnvironmentConfig{}, sourceFileNotUsed, nil
	case infra.FunctionSetupHocon:
		return &HoconConfig{fileName: fileName}, sourceFileInDisk, nil
	case infra.FunctionSetupJson:
		return &JSONConfig{FileName: fileName}, sourceFileInDisk, nil
	case infra.FunctionSetupYaml:
		return nil, sourceFileInDisk, errors.New(notSupportedTypeError +
			"yaml configurator source. Not implemented yet")
	case infra.FunctionSetupVault:
		return nil, sourceFielInServer, errors.New(notSupportedTypeError + "vault configurator source. Not implemented yet")
	case infra.FunctionSetupConfigServer:
		return nil, sourceFielInServer, errors.New(notSupportedTypeError + "configure server configurator source. Not implemented yet")
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
	pipelineChaines []infra.FuncType,
	// prefix by getting data from source placed in entry
	prefix string,
	// smartConfigure - analys structure by tags for find methods which should use for configuration
	smartConfigure bool) (result interface{}, err error) {

	defer func() {
		if e := recover(); e != nil {
			fmt.Println("[Pipeline]: ERROR: ", e)
			err = e.(error)
		}
	}()

	var pipeline *Pipeline

	if smartConfigure {
		analysedChains := tags.GetFunctionTypes(structure)
		pipeline = getFunctionChain(fileName, analysedChains)
	} else {
		pipeline = getFunctionChain(fileName, pipelineChaines)
	}

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
		fmt.Println("Level: error. Message: [Pipeline]: error while configuring your structure. Errors: ", err.Error())
	}
	return structure, nil
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
		default:
			return false
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
		if context.Prefix == "" {
			context.Prefix += valuePtr.Type().Name()
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

func (pipeline *Pipeline) getErrorAsOne() error {
	if len(pipeline.errors) > 0 {
		return errors.New("on stage recurisiveParseFields will have any of this errors: " + strings.Join(pipeline.errors, "\n"))
	}
	return nil
}

func (pipeline *Pipeline) configuringValues(context *structContext) error {
	valueIndirect := reflect.Indirect(context.Value)
	switch valueIndirect.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		valueGet := pipeline.chains.stageFunction.GetComplexType(context)
		fmt.Println("Loglevel: Debug Message: [Pipeline]:  value get from parsing slice: ", valueGet)
		if valueGet.CheckIsValue() {

			// add check for setuping valueGet in valueIndirect
			if valueIndirect.CanSet() {
				fmt.Println("Loglevel: Debug Message: [Pipeline]: setupe value in struct")
				valueIndirect.Set(valueGet.Value)
			} else {
				return errors.New("can not set " + valueIndirect.Kind().String() + " into struct field.")
			}
		} else {
			return errors.New("Loglevel: Debug Message:  value get not implementedable value: ")
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return errors.New("not supported types of unsigned integer")
	case reflect.String, reflect.Float32, reflect.Float64, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valueGet := pipeline.chains.stageFunction.GetBaseType(context)
		if valueGet.CheckIsValue() {
			fmt.Println("Loglevel: Debug Message: [Pipeline]: value get from parsing "+valueGet.Value.Kind().String()+": ", valueGet)
			if valueIndirect.CanSet() {
				fmt.Println("[Pipeline]: setupe value")
				valueIndirect.Set(valueGet.Value)
			} else {
				return errors.New("can not set " + valueIndirect.Kind().String() + " into struct field.")
			}
		} else {
			return errors.New("value get not implementedable value: ")
		}
	default:
		return errors.New("not supported type for hocon parsing")
	}
	return nil
}

func (pipeline *Pipeline) checkValueTypeIsPointer(value reflect.Value) error {
	if value.Kind() != reflect.Ptr {
		return errors.New("does not work without pointer types")
	}
	return nil
}

func (pipeline *Pipeline) checkValuePrefix(prefix string) error {
	if prefix == "" || prefix[len(prefix)-1] == '.' || len(prefix) <= 1 {
		return errors.New("can not have entry point")
	}
	return nil
}
