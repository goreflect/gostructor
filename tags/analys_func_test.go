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
				infra.FunctionSetupDefault,
				infra.FunctionSetupHocon,
				infra.FunctionSetupJson,
				infra.FunctionSetupVault,
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
