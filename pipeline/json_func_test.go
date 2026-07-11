package pipeline

import (
	"reflect"
	"testing"

	"github.com/goreflect/gostructor/infra"
)

type ContextVl struct {
	Value  []int  `cf_json:"complextArray"`
	String string `cf_json:"string"`
}

func TestJSONConfig_GetComplexType(t *testing.T) {
	valueSimple := ContextVl{}

	fieldStruct1Type := reflect.ValueOf(valueSimple).Type().Field(0)
	fieldStruct1Value := reflect.ValueOf(valueSimple).Field(0)

	config := JSONConfig{
		fileName: "../test_configs/config1.json",
	}
	got := config.GetComplexType(&structContext{
		Value:       fieldStruct1Value,
		StructField: fieldStruct1Type,
		Prefix:      "",
	})

	if got.GetNotAValue() != nil {
		t.Errorf("JSONConfig.GetComplexType() unexpected error = %v", got.GetNotAValue().Error)
	}
	if !reflect.DeepEqual(got.Value.Interface(), []int{1, 2, 3}) {
		t.Errorf("JSONConfig.GetComplexType() = %v, want %v", got.Value.Interface(), []int{1, 2, 3})
	}
}

func TestJSONConfig_typeSafeLoadConfigFile(t *testing.T) {
	type fields struct {
		FileName   string
		parsedData map[string]interface{}
	}
	type args struct {
		context *structContext
	}
	valueSimple := ContextVl{}
	fieldStruct1Value := reflect.ValueOf(valueSimple).Field(0)

	lastWant := infra.NewGoStructorNoValue(fieldStruct1Value, nil)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
		want1  *infra.GoStructorValue
	}{
		{
			name: "check error while loading parsing node from file",
			fields: fields{
				FileName:   "",
				parsedData: nil,
			},
			args: args{
				context: &structContext{
					Value: fieldStruct1Value,
				},
			},
			want:  false,
			want1: &lastWant,
		},
		{
			name: "check can not loading config from file. File Not Exist",
			fields: fields{
				FileName:   "../test_configs/config_err1231.json",
				parsedData: nil,
			},
			args: args{
				context: &structContext{
					Value: fieldStruct1Value,
				},
			},
			want:  false,
			want1: &lastWant,
		},
		{
			name: "check success loading config",
			fields: fields{
				FileName:   "../test_configs/config_err.json",
				parsedData: nil,
			},
			args: args{
				context: &structContext{
					Value: fieldStruct1Value,
				},
			},
			want:  true,
			want1: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &JSONConfig{
				fileName:   tt.fields.FileName,
				parsedData: tt.fields.parsedData,
			}
			got, got1 := config.typeSafeLoadConfigFile(tt.args.context)
			if got != tt.want {
				t.Errorf("JSONConfig.typeSafeLoadConfigFile() got = %v, want %v", got, tt.want)
			}
			if got1 != nil && !reflect.DeepEqual(got1.Value, tt.want1.Value) {
				t.Errorf("JSONConfig.typeSafeLoadConfigFile() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestJSONConfig_GetBaseType(t *testing.T) {
	type fields struct {
		FileName   string
		parsedData map[string]interface{}
	}
	type args struct {
		context *structContext
	}
	valueSimple := ContextVl{}
	fieldStruct2Value := reflect.ValueOf(valueSimple).Field(1)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    string
	}{
		{
			name: "check type parsed. Error",
			fields: fields{
				FileName: "unknownFile",
			},
			args: args{
				context: &structContext{
					Value: fieldStruct2Value,
				},
			},
			wantErr: true,
		},
		{
			name: "check success base type read",
			fields: fields{
				FileName: "../test_configs/config1.json",
			},
			args: args{
				context: &structContext{
					Value:       fieldStruct2Value,
					StructField: reflect.ValueOf(ContextVl{}).Type().Field(1),
				},
			},
			wantErr: false,
			want:    "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := JSONConfig{
				fileName:   tt.fields.FileName,
				parsedData: tt.fields.parsedData,
			}
			got := config.GetBaseType(tt.args.context)
			if tt.wantErr {
				if got.GetNotAValue() == nil {
					t.Errorf("JSONConfig.GetBaseType() expected error, got value %v", got.Value)
				}
				return
			}
			if got.GetNotAValue() != nil {
				t.Errorf("JSONConfig.GetBaseType() unexpected error = %v", got.GetNotAValue().Error)
				return
			}
			if got.Value.String() != tt.want {
				t.Errorf("JSONConfig.GetBaseType() = %v, want %v", got.Value.String(), tt.want)
			}
		})
	}
}
