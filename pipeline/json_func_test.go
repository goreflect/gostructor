package pipeline

import (
	"reflect"
	"strings"
	"testing"

	"github.com/go-restit/lzjson"
	"github.com/goreflect/gostructor/infra"
)

type ContextVl struct {
	Value  []int  `cf_json:"complextArray"`
	String string `cf_json:"string"`
}

func TestJSONConfig_GetComplexType(t *testing.T) {
	reader := strings.NewReader(`
	{
		"string": "test",
		"complextArray": [1,2,3]
	}
	`)
	valueSimple := ContextVl{}

	fieldStruct1Type := reflect.ValueOf(valueSimple).Type().Field(0)
	fieldStruct1Value := reflect.ValueOf(valueSimple).Field(0)

	type fields struct {
		FileName            string
		configureFileParsed lzjson.Node
	}
	type args struct {
		context *structContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   infra.GoStructorValue
	}{
		{
			name: "getting complex type error",
			fields: fields{
				FileName:            "test",
				configureFileParsed: lzjson.Decode(reader),
			},
			args: args{
				context: &structContext{
					Value:       fieldStruct1Value,
					StructField: fieldStruct1Type,
					Prefix:      "",
				},
			},
			want: infra.NewGoStructorNoValue(fieldStruct1Value.Interface(), nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := JSONConfig{
				FileName:            tt.fields.FileName,
				configureFileParsed: tt.fields.configureFileParsed,
			}
			got := config.GetComplexType(tt.args.context)

			if !reflect.DeepEqual(got.GetNotAValue().ValueAddress, fieldStruct1Value.Interface()) {
				t.Errorf("JSONConfig.GetComplexType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONConfig_typeSafeLoadConfigFile(t *testing.T) {
	type fields struct {
		FileName            string
		configureFileParsed lzjson.Node
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
				FileName:            "",
				configureFileParsed: nil,
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
				FileName:            "../test_configs/config_err1231.json",
				configureFileParsed: nil,
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
				FileName:            "../test_configs/config_err.json",
				configureFileParsed: nil,
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
				FileName:            tt.fields.FileName,
				configureFileParsed: tt.fields.configureFileParsed,
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
		FileName            string
		configureFileParsed lzjson.Node
	}
	type args struct {
		context *structContext
	}
	reader := strings.NewReader(`
	{
		"string": "test",
		"complextArray": [1,2,3]
	}
	`)
	test := lzjson.Decode(reader)
	if test.Get("string").IsNull() {
		t.Error("can not reading inside packet with json: ")
	}
	t.Log(test.Get("string").Len())
	valueSimple := ContextVl{}
	fieldStruct2Value := reflect.ValueOf(valueSimple).Field(1)
	fieldStruct2Type := reflect.ValueOf(valueSimple).Type().Field(1)

	lastWant := infra.NewGoStructorNoValue(fieldStruct2Value, nil)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   infra.GoStructorValue
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
			want: lastWant,
		}, {
			name: "Completed parsed value",
			fields: fields{
				configureFileParsed: lzjson.Decode(reader),
			},
			args: args{
				context: &structContext{
					Value:       fieldStruct2Value,
					StructField: fieldStruct2Type,
					Prefix:      "",
				},
			},
			want: infra.NewGoStructorTrueValue(reflect.ValueOf("test")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := JSONConfig{
				FileName:            tt.fields.FileName,
				configureFileParsed: tt.fields.configureFileParsed,
			}
			got := config.GetBaseType(tt.args.context)
			if !reflect.DeepEqual(got.GetNotAValue().ValueAddress, tt.want.GetNotAValue().ValueAddress) {
				t.Errorf("JSONConfig.GetBaseType() = %v, want %v", got, tt.want)
			}
		})
	}
}
