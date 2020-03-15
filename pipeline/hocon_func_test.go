package pipeline

import (
	"testing"

	gohocon "github.com/goreflect/go_hocon"
)

// preparing test structure for unit test below. It will be uncommented in the next issues while implementing infrastructures for writing unit tests
// func prepareTestFile() (*gohocon.Config, error) {
// 	result := `testMap = {
// 		Field1 = ""
// 	}
// 	`
// 	config, err := gohocon.ParseString(result)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return config, nil
// }

func TestHoconConfig_getElementName(t *testing.T) {
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
				context: &structContext{},
			},
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
