package converters

import (
	"errors"
	"reflect"
	"testing"

	"github.com/goreflect/gostructor/infra"
)

func Test_convertToFloat32(t *testing.T) {
	type args struct {
		source      reflect.Value
		destination reflect.Value
	}
	tests := []struct {
		name string
		args args
		want infra.GoStructorValue
	}{
		{
			name: "converted from float32 into float32",
			args: args{
				source:      reflect.ValueOf(float32(12.3)),
				destination: reflect.ValueOf(float32(0.0)),
			},
			want: infra.NewGoStructorTrueValue(reflect.ValueOf(float32(12.3))),
		},
		{
			name: "converted from int into float32",
			args: args{
				source:      reflect.ValueOf(int(12)),
				destination: reflect.ValueOf(float32(0.0)),
			},
			want: infra.NewGoStructorTrueValue(reflect.ValueOf(float32(12))),
		},
		{
			name: "converted from string into float32",
			args: args{
				source:      reflect.ValueOf("12.3"),
				destination: reflect.ValueOf(float32(0.0)),
			},
			want: infra.NewGoStructorTrueValue(reflect.ValueOf(float32(12.3))),
		},
		// {
		// 	name: "converted from float64 into float32",
		// 	args: args{
		// 		source:      reflect.ValueOf(float64(12.3)),
		// 		destination: reflect.ValueOf(float32(0.0)),
		// 	},
		// 	want: infra.NewGoStructorTrueValue(reflect.ValueOf(float32(12.3))),
		// },
		// {
		// 	name: "converted from string into float32 failed",
		// 	args: args{
		// 		source:      reflect.ValueOf("12.3asd"),
		// 		destination: reflect.ValueOf(float32(0.0)),
		// 	},
		// 	want: infra.NewGoStructorNoValue(reflect.ValueOf("12.3asd"), errors.New("")),
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToFloat32(tt.args.source, tt.args.destination); !reflect.DeepEqual(got.Value.Interface(), tt.want.Value.Interface()) {
				t.Errorf("convertToFloat32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToFloat32Failed(t *testing.T) {
	type args struct {
		source      reflect.Value
		destination reflect.Value
	}
	tests := []struct {
		name string
		args args
		want infra.GoStructorValue
	}{
		{
			name: "converted from float64 into float32 failed",
			args: args{
				source:      reflect.ValueOf(float64(12.3)),
				destination: reflect.ValueOf(float32(0.0)),
			},
			want: infra.NewGoStructorNoValue(reflect.ValueOf(float64(12.3)), errors.New("can not convert from float64 into float32")),
		},
		{
			name: "converted from string into float32 failed",
			args: args{
				source:      reflect.ValueOf("12.3asd"),
				destination: reflect.ValueOf(float32(0.0)),
			},
			want: infra.NewGoStructorNoValue(reflect.ValueOf("12.3asd"), errors.New("strconv.ParseFloat: parsing \"12.3asd\": invalid syntax")),
		},
		{
			name: "convert from sruct into float32 failed",
			args: args{
				source:      reflect.ValueOf(struct{ field1 string }{field1: "test"}),
				destination: reflect.ValueOf(float32(0.0)),
			},
			want: infra.NewGoStructorNoValue(reflect.ValueOf(struct{ field1 string }{field1: "test"}), errors.New("can not be converted from this type: struct beacuse this type not supported")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToFloat32(tt.args.source, tt.args.destination); !reflect.DeepEqual(got.GetNotAValue().Error.Error(), tt.want.GetNotAValue().Error.Error()) {
				t.Errorf("convertToFloat32() = %v, want %v", got.GetNotAValue().Error.Error(), tt.want.GetNotAValue().Error.Error())
			}
		})
	}
}

func Test_convertToFloat64(t *testing.T) {
	type args struct {
		source      reflect.Value
		destination reflect.Value
	}
	tests := []struct {
		name string
		args args
		want infra.GoStructorValue
	}{
		{
			name: "convert from int into float64",
			args: args{
				source:      reflect.ValueOf(int(12)),
				destination: reflect.ValueOf(float64(0.0)),
			},
			want: infra.NewGoStructorTrueValue(reflect.ValueOf(float64(12))),
		},
		{
			name: "convert from string into float64",
			args: args{
				source:      reflect.ValueOf("12.3"),
				destination: reflect.ValueOf(float64(0.0)),
			},
			want: infra.NewGoStructorTrueValue(reflect.ValueOf(float64(12.3))),
		},
		// TODO: upgrade this by change mantice
		// {
		// 	name: "convert from float32 into float64",
		// 	args: args{
		// 		source:      reflect.ValueOf(float32(12.3)),
		// 		destination: reflect.ValueOf(float64(0.0)),
		// 	},
		// 	want: infra.NewGoStructorTrueValue(reflect.ValueOf(float64(12.3))),
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToFloat64(tt.args.source, tt.args.destination); !reflect.DeepEqual(got.Value.Interface(), tt.want.Value.Interface()) {
				t.Errorf("convertToFloat64() = %v, want %v", got.Value.Interface(), tt.want.Value.Interface())
			}
		})
	}
}
