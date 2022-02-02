package tags

import (
	"reflect"
	"testing"

	"github.com/goreflect/gostructor/infra"
)

func TestGetFunctionTypes(t *testing.T) {
	type args struct {
		sourceStruct interface{}
	}
	tests := []struct {
		name string
		args args
		want []infra.FuncType
	}{
		{
			name: "first test",
			args: args{
				sourceStruct: struct {
					field1 string `cf_hocon:"test1" cf_default:"test2"`
					field2 struct {
						field  string `cf_vault:"test4"`
						field3 int    `cf_json:"kuil"`
					}
				}{},
			},
			want: []infra.FuncType{
				infra.FunctionSetupHocon,
				infra.FunctionSetupJSON,
				infra.FunctionSetupDefault,
				infra.FunctionSetupVault,
			},
		},
		{
			name: "check repeated tags",
			args: args{
				sourceStruct: struct {
					field1 string `cf_hocon:"test1" cf_default:"test2"`
					field2 int    `cf_hocon:"test3" cf_default:"test4"`
					field3 struct {
						field string `cf_vault:"token=1234,value=password" cf_yaml:"for_test" cf_env:"test" cf_server_file:"test"`
					}
				}{},
			},
			want: []infra.FuncType{
				infra.FunctionSetupEnvironment,
				infra.FunctionSetupHocon,
				infra.FunctionSetupYaml,
				infra.FunctionSetupDefault,
				infra.FunctionSetupVault,
				infra.FunctionSetupConfigServer,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFunctionTypes(tt.args.sourceStruct); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFunctionTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getFuncTypeByTag(t *testing.T) {
	type args struct {
		tagName string
	}
	tests := []struct {
		name string
		args args
		want infra.FuncType
	}{
		{
			name: "check if tagname not exist",
			args: args{
				tagName: "cf_unknown",
			},
			want: infra.FunctionNotExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFuncTypeByTag(tt.args.tagName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFuncTypeByTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_combineFields(t *testing.T) {
	type args struct {
		summCurrent []int
		newSumm     []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "check summCurrent more that newSumm",
			args: args{
				summCurrent: []int{1, 1, 1},
				newSumm:     []int{1, 1},
			},
			want: []int{2, 2, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := combineFields(tt.args.summCurrent, tt.args.newSumm); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("combineFields() = %v, want %v", got, tt.want)
			}
		})
	}
}
