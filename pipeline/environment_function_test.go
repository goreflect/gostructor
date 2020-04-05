package pipeline

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/tags"
)

func TestEnvironmentConfig_GetBaseType(t *testing.T) {
	myStruct := struct {
		field1 string `cf_env:"testBaseType"`
	}{}
	os.Setenv("testBaseType", "tururu")
	myStruct1 := reflect.Indirect(reflect.ValueOf(myStruct))
	field1 := myStruct1.Type().Field(0)
	t.Log("type of field: ", field1.Name+", "+field1.Type.Name())
	type args struct {
		context *structContext
	}
	tests := []struct {
		name   string
		config EnvironmentConfig
		args   args
		want   infra.GoStructorValue
	}{
		{
			name:   "get success base type string from environment",
			config: EnvironmentConfig{},
			args: args{
				context: &structContext{
					StructField: field1,
					Value:       myStruct1.Field(0),
				},
			},
			want: infra.NewGoStructorTrueValue(reflect.ValueOf("tururu")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := EnvironmentConfig{}
			got := config.GetBaseType(tt.args.context)
			os.Remove("testBaseType")

			if got.Value.String() != tt.want.Value.String() {
				t.Errorf("EnvironmentConfig.GetComplexType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvironmentConfig_GetBaseTypeFaield(t *testing.T) {
	myStruct := struct {
		field1 string `cf_env:""`
	}{}
	myStruct1 := reflect.Indirect(reflect.ValueOf(myStruct))
	field1 := myStruct1.Type().Field(0)
	t.Log("type of field: ", field1.Name+", "+field1.Type.Name())
	type args struct {
		context *structContext
	}
	tests := []struct {
		name   string
		config EnvironmentConfig
		args   args
		want   infra.GoStructorValue
	}{
		{
			name:   "get faield base type string from environment because not set name of this value",
			config: EnvironmentConfig{},
			args: args{
				context: &structContext{
					StructField: field1,
					Value:       myStruct1.Field(0),
				},
			},
			want: infra.NewGoStructorNoValue(reflect.ValueOf(myStruct1.Field(0)), errors.New("getBaseType can not get field by empty tag value of tag: "+tags.TagEnvironment)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := EnvironmentConfig{}
			got := config.GetBaseType(tt.args.context)
			if got.GetNotAValue().Error.Error() != tt.want.GetNotAValue().Error.Error() {
				t.Errorf("EnvironmentConfig.GetComplexType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvironmentConfig_GetComplexType(t *testing.T) {
	myStruct := struct {
		fieldhard []float32 `cf_env:"mySlice"`
	}{}
	os.Setenv("mySlice", "1.12,1.24,1.67")
	myStruct1 := reflect.Indirect(reflect.ValueOf(myStruct))
	field1 := myStruct1.Type().Field(0)
	t.Log("type of field: ", field1.Name+", "+field1.Type.Name())
	type args struct {
		context *structContext
	}
	tests := []struct {
		name   string
		config EnvironmentConfig
		args   args
		want   infra.GoStructorValue
	}{
		{
			name:   "get slice from environment into float32 slice structure field",
			config: EnvironmentConfig{},
			args: args{
				context: &structContext{
					StructField: field1,
					Value:       myStruct1.Field(0),
				},
			},
			want: infra.NewGoStructorTrueValue(reflect.ValueOf([]float32{1.12, 1.24, 1.67})),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := EnvironmentConfig{}
			got := config.GetComplexType(tt.args.context)
			if !reflect.DeepEqual(got.Value.Interface(), tt.want.Value.Interface()) {
				t.Errorf("EnvironmentConfig.GetComplexType() = %v, want %v", got, tt.want)
			}
		})
	}
}
