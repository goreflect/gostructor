package pipeline

import (
	"os"
	"reflect"
	"testing"

	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
	"github.com/stretchr/testify/assert"
)

func Test_getFunctionChain(t *testing.T) {
	type args struct {
		fileName       string
		pipelineChanes []infra.FuncType
	}
	tests := []struct {
		name string
		args args
		want *Pipeline
	}{
		{
			name: "error getting chain functinos by using not implemented source",
			args: args{
				fileName: "",
				pipelineChanes: []infra.FuncType{
					infra.FunctionNotExist,
				},
			},
			want: &Pipeline{chains: &Chain{stageFunction: nil, next: nil, notAValues: nil}, sourcesTypes: []int{0, 0, 0}},
		},
		{
			name: "check with implement source",
			args: args{
				fileName: "",
				pipelineChanes: []infra.FuncType{
					infra.FunctionSetupEnvironment,
					infra.FunctionSetupHocon,
				},
			},
			want: &Pipeline{chains: &Chain{
				stageFunction: &EnvironmentConfig{},
				next: &Chain{
					stageFunction: &HoconConfig{},
					next: &Chain{
						stageFunction: nil,
						next:          nil,
						notAValues:    nil,
					},
					notAValues: nil,
				},
				notAValues: nil,
			}, sourcesTypes: []int{1, 0, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFunctionChain(tt.args.pipelineChanes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFunctionChain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getChainByIdentifier(t *testing.T) {
	type args struct {
		idFunc infra.FuncType
	}
	tests := []struct {
		name    string
		args    args
		want    IConfigure
		want1   int
		wantErr bool
	}{
		{
			name: "check function setup default",
			args: args{
				idFunc: infra.FunctionSetupDefault,
			},
			want:    &DefaultConfig{},
			want1:   sourceFileNotUsed,
			wantErr: false,
		},
		{
			name: "check function setup environment",
			args: args{
				idFunc: infra.FunctionSetupEnvironment,
			},
			want:    &EnvironmentConfig{},
			want1:   sourceFileNotUsed,
			wantErr: false,
		},
		{
			name: "check function setup hocon",
			args: args{
				idFunc: infra.FunctionSetupHocon,
			},
			want:    &HoconConfig{},
			want1:   sourceFileInDisk,
			wantErr: false,
		},
		{
			name: "check function setup json",
			args: args{
				idFunc: infra.FunctionSetupJSON,
			},
			want:    &JSONConfig{},
			want1:   sourceFileInDisk,
			wantErr: false,
		},
		{
			name: "check function setup yaml",
			args: args{
				idFunc: infra.FunctionSetupYaml,
			},
			want:    &YamlConfig{},
			want1:   sourceFileInDisk,
			wantErr: false,
		},
		{
			name: "check function setup config server",
			args: args{
				idFunc: infra.FunctionSetupConfigServer,
			},
			want:    nil,
			want1:   sourceFielInServer,
			wantErr: true,
		},
		{
			name: "check function setup vault",
			args: args{
				idFunc: infra.FunctionSetupVault,
			},
			want:    &VaultConfig{},
			want1:   sourceFielInServer,
			wantErr: false,
		},
		{
			name: "check unknown setup function",
			args: args{
				idFunc: infra.FunctionNotExist,
			},
			want:    nil,
			want1:   sourceFileNotUsed,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getChainByIdentifier(tt.args.idFunc)
			if (err != nil) != tt.wantErr {
				t.Errorf("getChainByIdentifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getChainByIdentifier() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getChainByIdentifier() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_structContext_getFieldName(t *testing.T) {
	strctTest := testStructWithSimpleTypes{}
	fieldStructType1 := reflect.ValueOf(strctTest).Type().Field(0)
	fieldStructValue1 := reflect.ValueOf(strctTest).Field(0)
	type fields struct {
		Value       reflect.Value
		StructField reflect.StructField
		Prefix      string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
		want1  string
	}{
		{
			name: "check tag hocon in struct",
			fields: fields{
				Value:       fieldStructValue1,
				StructField: fieldStructType1,
			},
			want:  false,
			want1: "field1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := structContext{
				Value:       tt.fields.Value,
				StructField: tt.fields.StructField,
				Prefix:      tt.fields.Prefix,
			}
			got1 := context.getFieldName()
			if got1 != tt.want1 {
				t.Errorf("structContext.getFieldName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestPipeline_configuringValues(t *testing.T) {
	structureFoTest := struct {
		field1 uint8
		field2 []string `cf_env:"test"`
		field3 string
		field4 complex64
	}{
		field1: 7,
		field2: []string{"test1"},
		field3: "test2",
	}

	os.Setenv("test", "test,test2")
	fieldStructureType1 := reflect.ValueOf(structureFoTest).Type().Field(0)
	fieldStructureValue1 := reflect.ValueOf(structureFoTest).Field(0)
	fieldStructureType2 := reflect.ValueOf(structureFoTest).Type().Field(1)
	fieldStructureValue2 := reflect.ValueOf(structureFoTest).Field(1)
	fieldStructureType3 := reflect.ValueOf(structureFoTest).Type().Field(2)
	fieldStructureValue3 := reflect.ValueOf(structureFoTest).Field(2)
	fieldStructureType4 := reflect.ValueOf(structureFoTest).Type().Field(3)
	fieldStructureValue4 := reflect.ValueOf(structureFoTest).Field(3)
	chains := &Chain{
		stageFunction: EnvironmentConfig{},
	}
	type fields struct {
		chains       *Chain
		errors       []string
		sourcesTypes []int
	}
	type args struct {
		context *structContext
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "check error while configuring uint types",
			fields: fields{},
			args: args{
				context: &structContext{
					StructField: fieldStructureType1,
					Value:       fieldStructureValue1,
				},
			},
			wantErr: true,
		},
		{
			name: "check when getting complex type from chains",
			fields: fields{
				chains: chains,
			},
			args: args{
				context: &structContext{
					StructField: fieldStructureType2,
					Value:       fieldStructureValue2,
				},
			},
			wantErr: true,
		},
		{
			name: "check when getting base type from chains",
			fields: fields{
				chains: chains,
			},
			args: args{
				context: &structContext{
					StructField: fieldStructureType3,
					Value:       fieldStructureValue3,
				},
			},
			wantErr: true,
		},
		{
			name: "check when undefined type",
			fields: fields{
				chains: chains,
			},
			args: args{
				context: &structContext{
					StructField: fieldStructureType4,
					Value:       fieldStructureValue4,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := &Pipeline{
				chains:       tt.fields.chains,
				errors:       tt.fields.errors,
				sourcesTypes: tt.fields.sourcesTypes,
			}
			if err := pipeline.configuringValues(tt.args.context); (err != nil) != tt.wantErr {
				t.Errorf("Pipeline.configuringValues() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipeline_checkValuePrefix(t *testing.T) {
	type fields struct {
		chains       *Chain
		errors       []string
		sourcesTypes []int
	}
	type args struct {
		prefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "prefix is empty or prefix contain .",
			fields: fields{},
			args: args{
				prefix: ".",
			},
			wantErr: true,
		},
		{
			name:   "prefix not empty and not contain last sym as point",
			fields: fields{},
			args: args{
				prefix: "mysuper.prefix",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := &Pipeline{
				chains:       tt.fields.chains,
				errors:       tt.fields.errors,
				sourcesTypes: tt.fields.sourcesTypes,
			}
			if err := pipeline.checkValuePrefix(tt.args.prefix); (err != nil) != tt.wantErr {
				t.Errorf("Pipeline.checkValuePrefix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipeline_checkValueTypeIsPointer(t *testing.T) {
	type fields struct {
		chains       *Chain
		errors       []string
		sourcesTypes []int
	}
	type args struct {
		value reflect.Value
	}
	arg := int(4)
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "value is not ptr",
			fields: fields{},
			args: args{
				value: reflect.ValueOf(int8(4)),
			},
			wantErr: true,
		},
		{
			name:   "value is ptr",
			fields: fields{},
			args: args{
				value: reflect.ValueOf(&arg),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := &Pipeline{
				chains:       tt.fields.chains,
				errors:       tt.fields.errors,
				sourcesTypes: tt.fields.sourcesTypes,
			}
			if err := pipeline.checkValueTypeIsPointer(tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Pipeline.checkValueTypeIsPointer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipeline_setupValue(t *testing.T) {
	type fields struct {
		chains       *Chain
		errors       []string
		sourcesTypes []int
	}
	type args struct {
		context *structContext
		value   *infra.GoStructorValue
	}
	set := infra.NewGoStructorTrueValue(reflect.ValueOf(int8(4)))
	source := int8(0)
	testStruct := struct {
		Field1 *int8
	}{
		Field1: &source,
	}
	valueStructField1 := reflect.ValueOf(testStruct).Field(0)
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "check value can set into struct field",
			fields: fields{},
			args: args{
				context: &structContext{
					Value: valueStructField1,
				},
				value: &set,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := &Pipeline{
				chains:       tt.fields.chains,
				errors:       tt.fields.errors,
				sourcesTypes: tt.fields.sourcesTypes,
			}
			if err := pipeline.setupValue(tt.args.context, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Pipeline.setupValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipeline_getErrorAsOne(t *testing.T) {
	type fields struct {
		chains       *Chain
		errors       []string
		sourcesTypes []int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "return errors as one",
			fields: fields{
				errors: []string{"error1", "error2"},
			},
			wantErr: true,
		},
		{
			name: "return nil",
			fields: fields{
				errors: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := &Pipeline{
				chains:       tt.fields.chains,
				errors:       tt.fields.errors,
				sourcesTypes: tt.fields.sourcesTypes,
			}
			if err := pipeline.getErrorAsOne(); (err != nil) != tt.wantErr {
				t.Errorf("Pipeline.getErrorAsOne() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipeline_preparePrefix(t *testing.T) {
	type fields struct {
		chains       *Chain
		errors       []string
		sourcesTypes []int
	}
	type args struct {
		contextPrefix string
		value         reflect.StructField
	}

	testStructure := struct {
		Field1 string `cf_hocon:"context"`
	}{
		Field1: "",
	}
	fieldStruct := reflect.ValueOf(testStructure).Type().Field(0)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "get prefix with point",
			fields: fields{},
			args: args{
				contextPrefix: "",
				value:         fieldStruct,
			},
			want: "context",
		},
		{
			name:   "get prefix with point",
			fields: fields{},
			args: args{
				contextPrefix: "prefix",
				value:         fieldStruct,
			},
			want: "prefix.context",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := &Pipeline{
				chains:       tt.fields.chains,
				errors:       tt.fields.errors,
				sourcesTypes: tt.fields.sourcesTypes,
			}
			if got := pipeline.preparePrefix(tt.args.contextPrefix, tt.args.value); got != tt.want {
				t.Errorf("Pipeline.preparePrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigure(t *testing.T) {
	type args struct {
		structure       interface{}
		fileName        string
		pipelineChaines []infra.FuncType
		prefix          string
		smartConfigure  bool
	}

	myTestStruct := struct{ field1 string }{}

	tests := []struct {
		name       string
		args       args
		wantResult interface{}
		wantErr    bool
	}{
		{
			name: "smart configuring",
			args: args{
				structure:       &myTestStruct,
				fileName:        "",
				pipelineChaines: []infra.FuncType{infra.FunctionSetupDefault},
				prefix:          "",
				smartConfigure:  true,
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "smart configuring",
			args: args{
				structure:       &myTestStruct,
				fileName:        "test",
				pipelineChaines: []infra.FuncType{infra.FunctionSetupJSON},
				prefix:          "",
				smartConfigure:  true,
			},
			wantResult: nil,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := Configure(tt.args.structure, tt.args.pipelineChaines, tt.args.prefix, tt.args.smartConfigure)
			if (err != nil) != tt.wantErr {
				t.Errorf("Configure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("Configure() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestPipeline_recursiveParseFields(t *testing.T) {
	type fields struct {
		chains       *Chain
		errors       []string
		sourcesTypes []int
	}
	type args struct {
		context *structContext
	}
	structureTest := struct{ Field1 int8 }{}
	value := reflect.ValueOf(structureTest).Field(0)
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "should return error",
			fields: fields{},
			args: args{
				context: &structContext{
					Value: value,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := &Pipeline{
				chains:       tt.fields.chains,
				errors:       tt.fields.errors,
				sourcesTypes: tt.fields.sourcesTypes,
			}
			if err := pipeline.recursiveParseFields(tt.args.context); (err != nil) != tt.wantErr {
				t.Errorf("Pipeline.recursiveParseFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipeline_setNextChain(t *testing.T) {
	type fields struct {
		chains       *Chain
		errors       []string
		sourcesTypes []int
		curentChain  *Chain
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "correct changed current stage function",
			fields: fields{
				chains: &Chain{
					stageFunction: EnvironmentConfig{},
					next:          nil,
				},
				errors:       nil,
				sourcesTypes: nil,
				curentChain:  nil,
			},
			wantErr: false,
		},
		{
			name: "incorrect change current stage function",
			fields: fields{
				chains: &Chain{
					next: nil,
				},
				curentChain: &Chain{
					stageFunction: EnvironmentConfig{},
				},
			},
			wantErr: true,
		},
		{
			name: "correct change current stage function",
			fields: fields{
				curentChain: &Chain{
					stageFunction: EnvironmentConfig{},
					next: &Chain{
						stageFunction: &DefaultConfig{},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := &Pipeline{
				chains:       tt.fields.chains,
				errors:       tt.fields.errors,
				sourcesTypes: tt.fields.sourcesTypes,
				curentChain:  tt.fields.curentChain,
			}
			if err := pipeline.setNextChain(); (err != nil) != tt.wantErr {
				t.Errorf("Pipeline.setNextChain() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type TestStructPriority struct {
	field1 string `cf_default:"tururu" cf_env:"MY_TEST" cf_priority:"stage1:cf_env,cf_default;stage2:cf_default,cf_env"`
}

type TestStructTomlIni struct {
	field1 string   `cf_ini:"TEST#test" cf_toml:"postgres#password"`
	fiedl2 int32    `cf_ini:"TEST#test2" cf_toml:"postgres#test1"`
	field3 float32  `cf_ini:"TEST#test3" cf_toml:"postgres#test2"`
	field4 []string `cf_ini:"TEST#test4" cf_toml:"postgres#test3"`
	field5 int16    `cf_ini:"TEST#test5" cf_toml:"postgres#test4"`
}

func TestPipelineOrderConfiguring(t *testing.T) {
	type args struct {
		structure       interface{}
		fileName        string
		pipelineChaines []infra.FuncType
		prefix          string
		smartConfigure  bool
	}

	myTestStruct := TestStructPriority{}
	os.Setenv("MY_TEST", "TURURU")
	tests := []struct {
		name        string
		args        args
		wantResult  interface{}
		wantErr     bool
		setPriority string
	}{
		{
			name: "success configuring from cf_env",
			args: args{
				structure:       &myTestStruct,
				fileName:        "",
				pipelineChaines: []infra.FuncType{infra.FunctionSetupDefault},
				prefix:          "",
				smartConfigure:  true,
			},
			wantResult:  "TURURU",
			wantErr:     false,
			setPriority: "stage1",
		},
		{
			name: "success configuring from cf_default",
			args: args{
				structure:       &myTestStruct,
				fileName:        "test",
				pipelineChaines: []infra.FuncType{infra.FunctionSetupJSON},
				prefix:          "",
				smartConfigure:  true,
			},
			wantResult:  "tururu",
			wantErr:     false,
			setPriority: "stage2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GOSTRUCTOR_PRIORITY", tt.setPriority)
			gotResult, err := Configure(tt.args.structure, tt.args.pipelineChaines, tt.args.prefix, tt.args.smartConfigure)
			if (err != nil) != tt.wantErr {
				t.Errorf("Configure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotResult.(*TestStructPriority).field1 != tt.wantResult {
				t.Errorf("Not equaled result, ordering not working")
			}
		})
	}
}

func TestPipelineTomlConfiguring(t *testing.T) {
	type args struct {
		structure       interface{}
		fileName        string
		pipelineChaines []infra.FuncType
		prefix          string
		smartConfigure  bool
	}

	myTestStruct1 := TestStructTomlIni{}
	myTestStruct2 := TestStructTomlIni{}
	tests := []struct {
		name       string
		args       args
		wantResult interface{}
		wantErr    bool
		wantToml   bool
	}{
		{
			name: "success configuring from cf_toml",
			args: args{
				structure:       &myTestStruct1,
				fileName:        "",
				pipelineChaines: []infra.FuncType{infra.FunctionSetupDefault},
				prefix:          "",
				smartConfigure:  true,
			},
			wantResult: &TestStructTomlIni{
				field1: "mypassword",
				fiedl2: 1231,
				field3: 43.52,
				field4: []string{"myTest1", "myTest2"},
				field5: 123,
			},
			wantErr:  false,
			wantToml: true,
		},
		{
			name: "success configuring from cf_ini",
			args: args{
				structure:       &myTestStruct2,
				fileName:        "",
				pipelineChaines: []infra.FuncType{infra.FunctionSetupDefault},
				prefix:          "",
				smartConfigure:  true,
			},
			wantResult: &TestStructTomlIni{
				field1: "tururu",
				fiedl2: 614,
				field3: 86.27,
				field4: []string{"str1", "str2", "str3"},
				field5: 15,
			},
			wantErr:  false,
			wantToml: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantToml {
				os.Setenv(tags.TomlFile, "../test_configs/config.toml")
			} else {
				os.Setenv(tags.IniFile, "../test_configs/config.ini")
			}
			gotResult, err := Configure(tt.args.structure, tt.args.pipelineChaines, tt.args.prefix, tt.args.smartConfigure)
			if (err != nil) != tt.wantErr {
				t.Errorf("Configure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log("Got result : ", gotResult)
			assert.Equal(t, tt.wantResult, gotResult)
			os.Clearenv()
		})
	}
}
