package pipeline

import (
	"reflect"
	"testing"

	gohocon "github.com/goreflect/go_hocon"
)

func TestHoconConfig_getElementName(t *testing.T) {
	valSimple := testStructWithSimpleTypes{}
	fieldStruct1Type := reflect.ValueOf(valSimple).Type().Field(0)
	fieldStruct1Value := reflect.ValueOf(valSimple).Field(0)
	type fields struct {
		fileName            string
		configureFileParsed *gohocon.Config
	}
	type args struct {
		context *structContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "check how setup prefix name: ",
			fields: fields{
				fileName:            "testFile",
				configureFileParsed: &gohocon.Config{},
			},
			args: args{
				context: &structContext{
					Prefix: "test",
				},
			},
			want: "test",
		},
		{
			name: "currentTagHocon not insert in structure",
			fields: fields{
				fileName:            "",
				configureFileParsed: &gohocon.Config{},
			},
			args: args{
				context: &structContext{
					Prefix:      "test",
					StructField: fieldStruct1Type,
					Value:       fieldStruct1Value,
				},
			},
			want: "test.field1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HoconConfig{
				fileName:            tt.fields.fileName,
				configureFileParsed: tt.fields.configureFileParsed,
			}
			if got := config.getElementName(tt.args.context); got != tt.want {
				t.Errorf("HoconConfig.getElementName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHoconConfig_typeSafeLoadConfigFile(t *testing.T) {
	type fields struct {
		fileName            string
		configureFileParsed *gohocon.Config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "check loading configure source if not loaded success",
			fields: fields{
				fileName: "../test_configs/test1.hocon",
			},
			wantErr: false,
		},
		{
			name: "check loading configure source if not loaded failed",
			fields: fields{
				fileName: "../test_configs/test_not_exist.hocon",
			},
			wantErr: true,
		},
		{
			name: "check that configure source was loaded",
			fields: fields{
				configureFileParsed: &gohocon.Config{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &HoconConfig{
				fileName:            tt.fields.fileName,
				configureFileParsed: tt.fields.configureFileParsed,
			}
			if err := config.typeSafeLoadConfigFile(); (err != nil) != tt.wantErr {
				t.Errorf("HoconConfig.typeSafeLoadConfigFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
