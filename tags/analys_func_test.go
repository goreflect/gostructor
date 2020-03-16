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
				infra.FunctionSetupJson,
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
						field string `cf_vault:"token=1234,value=password"`
					}
				}{},
			},
			want: []infra.FuncType{
				infra.FunctionSetupHocon,
				infra.FunctionSetupDefault,
				infra.FunctionSetupVault,
				infra.FunctionSetupJson,
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
