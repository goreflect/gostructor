package tools

import (
	"reflect"
	"testing"
)

func TestConvertStringIntoArray(t *testing.T) {
	type args struct {
		value         string
		configuration ConfigureConverts
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "not setup separator",
			args: args{
				value:         "test1,test2",
				configuration: ConfigureConverts{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "completed split",
			args: args{
				value: "test1,test2",
				configuration: ConfigureConverts{
					Separator: COMMA,
				},
			},
			want:    []string{"test1", "test2"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertStringIntoArray(tt.args.value, tt.args.configuration)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertStringIntoArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertStringIntoArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
