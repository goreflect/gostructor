package converters

import (
	"errors"
	"reflect"
	"testing"

	"github.com/goreflect/gostructor/pipeline"
)

func TestConvertBetweenPrimitiveTypes(t *testing.T) {
	type args struct {
		source      reflect.Value
		destination reflect.Value
	}
	tests := []struct {
		name string
		args args
		want pipeline.GoStructorValue
	}{
		{
			name: "not supported type return GoStructorNoValue",
			args: args{
				source:      reflect.ValueOf(1),
				destination: reflect.ValueOf(struct{ name string }{name: "test"}),
			},
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(struct{ name string }{name: "test"}),
				errors.New("can not converted to this type"+reflect.Struct.String()+" beacuse this type not supported"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertBetweenPrimitiveTypes(tt.args.source, tt.args.destination)
			if got.GetNotAValue().Error != nil {
				t.Log("completed. Error: ", got.GetNotAValue().Error.Error())
			} else {
				t.Error("success converted not supported")
			}
		})
	}
}

func TestConvertBetweenPrimitiveTypesToIntFromIntSuccess(t *testing.T) {
	type args struct {
		source      reflect.Value
		destination reflect.Value
	}
	tests := []struct {
		name string
		args args
		want pipeline.GoStructorValue
	}{
		{
			name: "success convert from int to int",
			args: args{
				source:      reflect.ValueOf(1),
				destination: reflect.ValueOf(0),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(1)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertBetweenPrimitiveTypes(tt.args.source, tt.args.destination)
			if got.Value.Interface().(int) == 1 {
				t.Log("completed")
			} else {
				t.Error("not setuped into destination")
			}
		})
	}
}

func TestConvertBetweenPrimitiveTypesToIntFromIntFailed(t *testing.T) {
	type args struct {
		source      reflect.Value
		destination reflect.Value
	}
	tests := []struct {
		name string
		args args
		want pipeline.GoStructorValue
	}{
		{
			name: "failed convert from int to int",
			args: args{
				source:      reflect.ValueOf(0000.1),
				destination: reflect.ValueOf(int(0)),
			},
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(int(0)), errors.New("can not converted from this type: "+reflect.Float32.String()+" beacuse this type not supported")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertBetweenPrimitiveTypes(tt.args.source, tt.args.destination)
			if got.GetNotAValue().Error != tt.want.GetNotAValue().Error {
				t.Log("completed")
			} else {
				t.Error("error while convert between int types because source is more than 64 bit value")
			}
		})
	}
}

func TestConvertBetweenPrimitiveTypesToIntFromStringSuccess(t *testing.T) {
	type args struct {
		source      reflect.Value
		destination reflect.Value
	}
	tests := []struct {
		name string
		args args
		want pipeline.GoStructorValue
	}{
		{
			name: "success converting from string into int",
			args: args{
				source:      reflect.ValueOf("123"),
				destination: reflect.ValueOf(0),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(123)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertBetweenPrimitiveTypes(tt.args.source, tt.args.destination)
			if got.Value.Interface().(int) == 123 {
				t.Log("completed")
			} else {
				t.Error("not setuped into destination")
			}
		})
	}
}

func TestConvertBetweenPrimitiveTypesToIntFromStringFailed(t *testing.T) {
	type args struct {
		source      reflect.Value
		destination reflect.Value
	}
	tests := []struct {
		name string
		args args
		want pipeline.GoStructorValue
	}{
		{
			name: "failed convert from string into int",
			args: args{
				source:      reflect.ValueOf("12f3"),
				destination: reflect.ValueOf(0),
			},
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(0), errors.New("can not converted to this type: "+reflect.Int.String())),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertBetweenPrimitiveTypes(tt.args.source, tt.args.destination)
			if got.GetNotAValue().Error != nil {
				t.Log("completed")
			} else {
				t.Error("can not be converted")
			}
		})
	}
}

func Test_convertToInt8FromStringSuccess(t *testing.T) {
	destination := reflect.ValueOf(int8(0))
	type args struct {
		source      reflect.Value
		destination reflect.Value
	}
	tests := []struct {
		name string
		args args
		want pipeline.GoStructorValue
	}{
		{
			name: "convert from string into int8 success",
			args: args{
				source:      reflect.ValueOf("111"),
				destination: reflect.ValueOf(int8(0)),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(int8(111))),
		},
		{
			name: "convert from stirng into int8 failed",
			args: args{
				source:      reflect.ValueOf("111.1"),
				destination: reflect.ValueOf(int8(0)),
			},
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(int8(0)), errors.New("can not converted to this type: "+reflect.Int8.String())),
		},
		{
			name: "convert from struct to int8 failed",
			args: args{
				source:      reflect.ValueOf(struct{ fieldTest string }{fieldTest: "test"}),
				destination: destination,
			},
			want: pipeline.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+reflect.Struct.String()+" beacuse this type not supported")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt8(tt.args.source, tt.args.destination); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToInt8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToInt16(t *testing.T) {
	type args struct {
		source      reflect.Value
		destination reflect.Value
	}
	tests := []struct {
		name string
		args args
		want pipeline.GoStructorValue
	}{
		{
			name: "convert from string into int16 success",
			args: args{
				source:      reflect.ValueOf("1234"),
				destination: reflect.ValueOf(int16(0)),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(int16(1234))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt16(tt.args.source, tt.args.destination); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToInt16() = %v, want %v", got, tt.want)
			}
		})
	}
}
