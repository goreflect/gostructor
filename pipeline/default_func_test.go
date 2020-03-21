package pipeline

import (
	"errors"
	"reflect"
	"testing"

	"github.com/goreflect/gostructor/infra"
)

func TestDefaultConfig_GetBaseType(t *testing.T) {
	strct := struct {
		field1 int8 `cf_default:"5"`
	}{}
	fieldType := reflect.ValueOf(strct).Type().Field(0)
	fieldValue := reflect.ValueOf(strct).Field(0)
	type args struct {
		context *structContext
	}
	tests := []struct {
		name   string
		config DefaultConfig
		args   args
		want   infra.GoStructorValue
	}{
		{
			name: "check while configuring base type int8",
			args: args{
				context: &structContext{
					Value:       fieldValue,
					StructField: fieldType,
				},
			},
			want: infra.NewGoStructorTrueValue(reflect.ValueOf(int8(5))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig{}
			if got := config.GetBaseType(tt.args.context); !reflect.DeepEqual(got.Value, tt.want.Value) {
				t.Errorf("DefaultConfig.GetBaseType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig_GetBaseTypeFaield(t *testing.T) {
	strct := struct {
		field1 int8 `cf_default:""`
	}{}
	fieldType := reflect.ValueOf(strct).Type().Field(0)
	fieldValue := reflect.ValueOf(strct).Field(0)
	type args struct {
		context *structContext
	}
	tests := []struct {
		name   string
		config DefaultConfig
		args   args
		want   infra.GoStructorValue
	}{
		{
			name: "check while configuring empty tag Value field",
			args: args{
				context: &structContext{
					Value:       fieldValue,
					StructField: fieldType,
				},
			},
			want: infra.NewGoStructorNoValue(fieldValue, errors.New("wrong format inside the tag")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig{}
			if got := config.GetBaseType(tt.args.context); !reflect.DeepEqual(got.Value, tt.want.Value) {
				t.Errorf("DefaultConfig.GetBaseType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig_GetComplexType(t *testing.T) {
	strct := struct {
		field []int8 `cf_default:"12,2,45,16"`
	}{}
	fieldType := reflect.ValueOf(strct).Type().Field(0)
	fieldValue := reflect.ValueOf(strct).Field(0)
	type args struct {
		context *structContext
	}
	tests := []struct {
		name   string
		config DefaultConfig
		args   args
		want   infra.GoStructorValue
	}{
		{
			name:   "get slice from default tag",
			config: DefaultConfig{},
			args: args{
				context: &structContext{
					Value:       fieldValue,
					StructField: fieldType,
				},
			},
			want: infra.NewGoStructorTrueValue(reflect.ValueOf([]int8{12, 2, 45, 16})),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig{}
			if got := config.GetComplexType(tt.args.context); !reflect.DeepEqual(got.Value.Interface(), tt.want.Value.Interface()) {
				t.Errorf("DefaultConfig.GetComplexType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig_GetComplexTypeNotImlemented(t *testing.T) {
	strct := struct {
		field map[string]string `cf_default:"12:12sda,51:5sda"`
	}{}
	fieldType := reflect.ValueOf(strct).Type().Field(0)
	fieldValue := reflect.ValueOf(strct).Field(0)
	type args struct {
		context *structContext
	}
	tests := []struct {
		name   string
		config DefaultConfig
		args   args
		want   infra.GoStructorValue
	}{
		{
			name:   "get map from default tag not implemented",
			config: DefaultConfig{},
			args: args{
				context: &structContext{
					Value:       fieldValue,
					StructField: fieldType,
				},
			},
			want: infra.NewGoStructorNoValue(fieldValue, errors.New("not implemented")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig{}
			if got := config.GetComplexType(tt.args.context); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultConfig.GetComplexType() = %v, want %v", got, tt.want)
			}
		})
	}
}
