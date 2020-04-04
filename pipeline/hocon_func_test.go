package pipeline

import (
	"errors"
	"reflect"
	"testing"

	gohocon "github.com/goreflect/go_hocon"
	"github.com/goreflect/gostructor/infra"
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

func TestHoconConfig_GetComplexType(t *testing.T) {
	type fields struct {
		fileName            string
		configureFileParsed *gohocon.Config
	}
	type args struct {
		context *structContext
	}
	value := reflect.ValueOf(int8(0))
	tests := []struct {
		name   string
		fields fields
		args   args
		want   infra.GoStructorValue
	}{
		{
			name: "error loading file",
			fields: fields{
				fileName: "test",
			},
			args: args{
				context: &structContext{
					Value: value,
				},
			},
			want: infra.NewGoStructorNoValue(value, errors.New("open test: no such file or directory")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HoconConfig{
				fileName:            tt.fields.fileName,
				configureFileParsed: tt.fields.configureFileParsed,
			}
			if got := config.GetComplexType(tt.args.context); got.GetNotAValue().Error.Error() != tt.want.GetNotAValue().Error.Error() {
				t.Errorf("HoconConfig.GetComplexType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHoconConfig_GetComplexTypeValueSlice(t *testing.T) {
	type fields struct {
		fileName            string
		configureFileParsed *gohocon.Config
	}
	type args struct {
		context *structContext
	}
	testStructure := struct {
		myField1   []string
		myMap      map[string]int
		myBaseType int
	}{}
	fieldTypeStructure1 := reflect.ValueOf(testStructure).Type().Field(0)
	fieldValueStructure1 := reflect.ValueOf(testStructure).Field(0)
	fieldTypeStructure2 := reflect.ValueOf(testStructure).Type().Field(1)
	fieldValueStructure2 := reflect.ValueOf(testStructure).Field(1)
	fieldTypeStructure3 := reflect.ValueOf(testStructure).Type().Field(2)
	fieldValueStructure3 := reflect.ValueOf(testStructure).Field(2)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   infra.GoStructorValue
	}{
		{
			name: "complete configuring slice by hocon",
			fields: fields{
				fileName: "../test_configs/testmap.hocon",
			},
			args: args{
				context: &structContext{
					Prefix:      "TestHocon.myField1",
					Value:       fieldValueStructure1,
					StructField: fieldTypeStructure1,
				},
			},
			want: infra.NewGoStructorTrueValue(reflect.ValueOf([]string{"test1", "test2"})),
		},
		{
			name: "complete configuring map by hocon",
			fields: fields{
				fileName: "../test_configs/testmap.hocon",
			},
			args: args{
				context: &structContext{
					Prefix:      "TestHocon.myMap",
					Value:       fieldValueStructure2,
					StructField: fieldTypeStructure2,
				},
			},
			want: infra.NewGoStructorTrueValue(reflect.ValueOf(map[string]int{
				"test1": 1,
				"test2": 2,
			})),
		},
		{
			name: "complete configuring int by hocon",
			fields: fields{
				fileName: "../test_configs/testmap.hocon",
			},
			args: args{
				context: &structContext{
					Prefix:      "TestHocon.myBaseType",
					Value:       fieldValueStructure3,
					StructField: fieldTypeStructure3,
				},
			},
			want: infra.NewGoStructorTrueValue(reflect.ValueOf(int(1))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HoconConfig{
				fileName:            tt.fields.fileName,
				configureFileParsed: tt.fields.configureFileParsed,
			}
			if got := config.GetComplexType(tt.args.context); !reflect.DeepEqual(got.Value.Interface(), tt.want.Value.Interface()) {
				t.Errorf("HoconConfig.GetComplexType() = %v, want %v", got.Value, tt.want.Value)
			}
		})
	}
}
