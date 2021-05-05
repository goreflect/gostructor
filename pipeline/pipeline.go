package pipeline

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"unsafe"

	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
	"github.com/goreflect/gostructor/tags/priority"
	"github.com/sirupsen/logrus"
)

type (
	// Pipeline -
	Pipeline struct {
		chains       *Chain
		errors       []string
		sourcesTypes []int
		curentChain  *Chain
		priority     string
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
	/*SmartConfiguring - flag which inform library need for start analyzing all tags in derived structure for setting function types of configuring structure
	 */
	SmartConfiguring = true
	/*DirtyConfiguring - flag*/
	DirtyConfiguring = false

	sourceFileInDisk   = 0
	sourceFielInServer = 1
	sourceFileNotUsed  = 2

	notSupportedTypeError = "not supported type "
	// EmptyAdditionalPrefix - empty of node prefix name
	EmptyAdditionalPrefix = ""
)

// bool - skip preview prefix or not (if user setup node it is return true, if user setup path that's return false)
// TODO: add error interface while user setup not knowing type of tags. And maybe if user make a mistake (this case need for autocorrection or autofill)
// Deprecated.
func (context structContext) getFieldName() string {
	for _, val := range []string{
		tags.TagHocon,
		tags.TagJSON,
		tags.TagYaml,
		tags.TagDefault,
		tags.TagEnvironment,
		tags.TagServerFile,
		tags.TagHashiCorpVault,
		tags.TagServerKeyValue,
		tags.TagIni,
		tags.TagToml} {
		tag := context.StructField.Tag.Get(val)
		if tag == "" {
			continue
		}
		return tag
	}
	return context.StructField.Name
}

// get pipeline of functions in chain notation
func getFunctionChain(pipelineChanes []infra.FuncType) *Pipeline {
	chain := &Chain{
		stageFunction: nil,
		next:          nil,
		notAValues:    nil,
	}
	sourcesTypes := make([]int, 3)

	currentChain := chain
	for _, pipelineChain := range pipelineChanes {
		stageFunction, sourceType, err := getChainByIdentifier(pipelineChain)
		if err != nil {
			logrus.Error("error while getting chain stage function. Error: " + err.Error())
			continue
		}
		sourcesTypes[sourceType]++
		currentChain.stageFunction = stageFunction
		currentChain.next = new(Chain)
		currentChain = currentChain.next
	}

	return &Pipeline{chains: chain, sourcesTypes: sourcesTypes, priority: extractPriorityFromEnv()}
}

func getChainsByPriority(context *structContext, pipelineChanes []infra.FuncType) *Chain {
	chain := &Chain{
		stageFunction: nil,
		next:          nil,
		notAValues:    nil,
	}
	sourcesTypes := make([]int, 3)

	currentChain := chain
	for _, pipelineChain := range pipelineChanes {
		stageFunction, sourceType, err := getChainByIdentifier(pipelineChain)
		if err != nil {
			logrus.Error("error while getting chain stage function. Error: " + err.Error())
			continue
		}
		sourcesTypes[sourceType]++
		currentChain.stageFunction = stageFunction
		currentChain.next = new(Chain)
		currentChain = currentChain.next
	}

	return chain
}

func extractPriorityFromEnv() string {
	return os.Getenv(tags.Priority)
}

func getFuncTypesByTagsName(tagsResult []string) []infra.FuncType {
	result := []infra.FuncType{}
	for _, value := range tagsResult {
		switch value {
		case tags.TagHocon:
			result = append(result, infra.FunctionSetupHocon)
		case tags.TagJSON:
			result = append(result, infra.FunctionSetupJSON)
		case tags.TagEnvironment:
			result = append(result, infra.FunctionSetupEnvironment)
		case tags.TagYaml:
			result = append(result, infra.FunctionSetupYaml)
		case tags.TagIni:
			result = append(result, infra.FunctionSetupIni)
		case tags.TagToml:
			result = append(result, infra.FunctionSetupToml)
		case tags.TagDefault:
			result = append(result, infra.FunctionSetupDefault)
		case tags.TagHashiCorpVault:
			result = append(result, infra.FunctionSetupVault)
		case tags.TagServerFile:
			result = append(result, infra.FunctionSetupConfigServer)
		case tags.TagServerKeyValue:
			result = append(result, infra.FunctionKeyValueServer)
		default:
			continue
		}
	}
	return result
}

// getChainByIdentifier - get function configure source and return
// Configure func struct contain
// hasSourcesFile - if in configuration pipeline have source configuring from file
// doesntHaveSourceFile - if source is another source
func getChainByIdentifier(
	// idFunc - identifier of function configuration source
	idFunc infra.FuncType) (IConfigure, int, error) {
	switch idFunc {
	case infra.FunctionSetupDefault:
		return &DefaultConfig{}, sourceFileNotUsed, nil
	case infra.FunctionSetupEnvironment:
		return &EnvironmentConfig{}, sourceFileNotUsed, nil
	case infra.FunctionSetupHocon:
		return &HoconConfig{}, sourceFileInDisk, nil
	case infra.FunctionSetupJSON:
		return &JSONConfig{}, sourceFileInDisk, nil
	case infra.FunctionSetupYaml:
		return &YamlConfig{}, sourceFileInDisk, nil
	case infra.FunctionSetupVault:
		return &VaultConfig{}, sourceFielInServer, nil
	case infra.FunctionSetupIni:
		return &IniConfig{}, sourceFileInDisk, nil
	case infra.FunctionSetupToml:
		return &TomlConfig{}, sourceFileInDisk, nil
	case infra.FunctionSetupConfigServer:
		return nil, sourceFielInServer, errors.New(notSupportedTypeError + "configure server configurator source. Not implemented yet")
	case infra.FunctionKeyValueServer:
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
	// functions will be configure structure
	pipelineChains []infra.FuncType,
	// prefix by getting data from source placed in entry
	prefix string,
	// smartConfigure - analys structure by tags for find methods which should use for configuration
	smartConfigure bool) (result interface{}, err error) {

	defer func() {
		if e := recover(); e != nil {
			logrus.Error(e)
			err = e.(error)
		}
	}()

	var pipeline *Pipeline

	if smartConfigure {
		analysedChains := tags.GetFunctionTypes(structure)
		pipeline = getFunctionChain(analysedChains)
	} else {
		pipeline = getFunctionChain(pipelineChains)
	}

	if err := pipeline.recursiveParseFields(&structContext{
		Value:  reflect.ValueOf(structure),
		Prefix: prefix,
	}); err != nil {
		logrus.Error("error while configuring your structure. Errors: ", err.Error())
		return nil, err
	}
	return structure, nil
}

func (pipeline *Pipeline) recursiveParseFields(context *structContext) error {
	if err := pipeline.checkValueTypeIsPointer(context.Value); err != nil {
		return err
	}
	valuePtr := reflect.Indirect(context.Value)
	switch valuePtr.Kind() {
	case reflect.Struct:
		return pipeline.prepareInlineStructFields(valuePtr, context)
	default:
		pipeline.extractOrderChainOrUseSettingUpChains(context)
		for {
			if err := pipeline.configuringValues(context); err != nil {
				if errSettingChain := pipeline.setNextChain(); errSettingChain != nil {
					return pipeline.getErrorAsOne()
				}
				continue
			}
			break
		}
		return nil
	}
}

func (pipeline *Pipeline) extractOrderChainOrUseSettingUpChains(context *structContext) {
	tag := context.StructField.Tag.Get(tags.TagPriority)
	if tag != "" {
		analyzedParser := priority.NewParser(strings.NewReader(tag))
		analyzedOrder, err := analyzedParser.Parse()
		if err == nil {
			resultDirty := priority.GetPriorityChains(analyzedOrder, pipeline.priority)
			pipeline.curentChain = getChainsByPriority(context, getFuncTypesByTagsName(resultDirty))
			return
		}
		logrus.Error("Can not worked with priority order: ", err)
	}
	pipeline.curentChain = pipeline.chains
}

func (pipeline *Pipeline) prepareInlineStructFields(value reflect.Value, context *structContext) error {
	if context.Prefix == "" {
		context.Prefix += value.Type().Name()
	}
	for i := 0; i < value.NumField(); i++ {
		if err := pipeline.recursiveParseFields(&structContext{
			Value:       value.Field(i).Addr(),
			StructField: value.Type().Field(i),
			Prefix:      pipeline.preparePrefix(context.Prefix, value.Type().Field(i)),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (pipeline *Pipeline) preparePrefix(contextPrefix string, value reflect.StructField) string {
	prefix := contextPrefix
	if err := pipeline.checkValuePrefix(contextPrefix); err == nil {
		prefix += "."
	}
	getPrefix := (structContext{StructField: value}.getFieldName())
	return prefix + getPrefix
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
		if pipeline.curentChain != nil {
			valueGet := pipeline.curentChain.stageFunction.GetComplexType(context)
			logrus.Debug("value get from parsing slice: ", valueGet)
			return pipeline.setupValue(context, &valueGet)
		}
		return errors.New("can not be configuring complex type. Can not configured current field")
	case
		reflect.String,
		reflect.Float32,
		reflect.Float64,
		reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		if pipeline.curentChain != nil {
			valueGet := pipeline.curentChain.stageFunction.GetBaseType(context)
			return pipeline.setupValue(context, &valueGet)
		}
		return errors.New("can not be configuring base type. Can not configured current field")
	default:
		return errors.New("not supported type for parsing")
	}
}

func (pipeline *Pipeline) setupValue(context *structContext, value *infra.GoStructorValue) error {
	valueIndirect := reflect.Indirect(context.Value)
	if value.CheckIsValue() {
		// add check for setuping valueGet in valueIndirect
		if valueIndirect.CanSet() {
			logrus.Debug("setup value in struct")
			valueIndirect.Set(value.Value)
			return nil
		}
		// change data by address
		logrus.Info(context.Value.Type())
		logrus.Info(valueIndirect.Type())
		unsafeValue := reflect.NewAt(valueIndirect.Type(), unsafe.Pointer(valueIndirect.UnsafeAddr())).Elem()
		unsafeValue.Set(value.Value)
		return nil

	}

	return errors.New("Loglevel: Debug Message:  value get not implementedable value: ")
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

func (pipeline *Pipeline) setNextChain() error {
	if pipeline.curentChain == nil {
		pipeline.curentChain = pipeline.chains
		return nil
	}
	if pipeline.curentChain.next == nil || pipeline.curentChain.next.stageFunction == nil {
		return errors.New("can not change chain function")
	}
	pipeline.curentChain = pipeline.curentChain.next
	return nil
}
