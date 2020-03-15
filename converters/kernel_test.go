package converters

import (
	"reflect"
	"testing"
)

func TestConvertBetweenPrimitiveTypes(t *testing.T) {
	type args struct {
		source      reflect.Value
		destination reflect.Value
	}
	tests := []struct {
		name    string
		args    args
		want    reflect.Value
		wantErr bool
	}{
		{
			name: "test success convert into int",
			args: args{
				source: reflect.ValueOf("1234"),
				destination: reflect.Indirect(reflect.ValueOf(23)),
			},
			want: reflect.ValueOf(1234),
			wantErr: false,
		},
		{
			name: "failed convert into int from string",
			args: args{
				source: reflect.ValueOf("test"),
				destination: reflect.Indirect(reflect.ValueOf(0)),
			},
			want: reflect.ValueOf(-1),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertBetweenPrimitiveTypes(tt.args.source, tt.args.destination)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertBetweenPrimitiveTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(tt.want)
			if !reflect.DeepEqual(got.Int(), tt.want.Int()) {
				t.Errorf("ConvertBetweenPrimitiveTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}
