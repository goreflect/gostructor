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

func TestConvertBetweenPrimitiveTypesInt8(t *testing.T) {
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
			name: "gostructor value with int8 converted type",
			args: args{
				source:      reflect.ValueOf(1),
				destination: reflect.ValueOf(int8(0)),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(int8(1))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertBetweenPrimitiveTypes(tt.args.source, tt.args.destination)
			if got.Value.Interface().(int8) != tt.want.Value.Interface().(int8) {
				t.Error("convertToInt16() = %w, want %w", got, tt.want)
			}
		})
	}
}

func TestConvertBetweenPrimitiveTypesInt16(t *testing.T) {
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
			name: "gostructor value with int16 converted type",
			args: args{
				source:      reflect.ValueOf(1),
				destination: reflect.ValueOf(int16(0)),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(int16(1))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertBetweenPrimitiveTypes(tt.args.source, tt.args.destination)
			if got.Value.Interface().(int16) != tt.want.Value.Interface().(int16) {
				t.Error("convertToInt16() = %w, want %w", got, tt.want)
			}
		})
	}
}

func TestConvertBetweenPrimitiveTypesInt32(t *testing.T) {
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
			name: "gostructor value with int32 converted type",
			args: args{
				source:      reflect.ValueOf(1),
				destination: reflect.ValueOf(int32(0)),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(int32(1))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertBetweenPrimitiveTypes(tt.args.source, tt.args.destination)
			if got.Value.Interface().(int32) != tt.want.Value.Interface().(int32) {
				t.Error("convertToInt16() = %w, want %w", got, tt.want)
			}
		})
	}
}

func TestConvertBetweenPrimitiveTypesInt64(t *testing.T) {
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
			name: "gostructor value with int64 converted type",
			args: args{
				source:      reflect.ValueOf(1),
				destination: reflect.ValueOf(int64(0)),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(int64(1))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertBetweenPrimitiveTypes(tt.args.source, tt.args.destination)
			if got.Value.Interface().(int64) != tt.want.Value.Interface().(int64) {
				t.Error("convertToInt16() = %w, want %w", got, tt.want)
			}
		})
	}
}

func TestConvertBetweenPrimitiveTypesString(t *testing.T) {
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
			name: "gostructor value with string converted type",
			args: args{
				source:      reflect.ValueOf(1),
				destination: reflect.ValueOf(""),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf("1")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertBetweenPrimitiveTypes(tt.args.source, tt.args.destination)
			if got.Value.Interface().(string) != tt.want.Value.Interface().(string) {
				t.Error("convertToInt16() = %w, want %w", got, tt.want)
			}
		})
	}
}

func TestConvertBetweenPrimitiveTypesBool(t *testing.T) {
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
			name: "gostructor value with bool converted type",
			args: args{
				source:      reflect.ValueOf(true),
				destination: reflect.ValueOf(bool(false)),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(true)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertBetweenPrimitiveTypes(tt.args.source, tt.args.destination)
			if got.Value.Interface().(bool) != tt.want.Value.Interface().(bool) {
				t.Error("convertToInt16() = %w, want %w", got, tt.want)
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

func TestConvertBetweenPrimitiveTypesToIntFromIntSuccess1(t *testing.T) {
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
			name: "success converting from int into int8",
			args: args{
				source:      reflect.ValueOf(int16(1)),
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
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(int8(0)),
				errors.New("can not converted to this type: "+reflect.Int8.String())),
		},
		{
			name: "convert from struct to int8 failed",
			args: args{
				source:      reflect.ValueOf(struct{ fieldTest string }{fieldTest: "test"}),
				destination: destination,
			},
			want: pipeline.NewGoStructorNoValue(destination, errors.New("can not converted from this type: "+
				reflect.Struct.String()+" beacuse this type not supported")),
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

func Test_convertToInt16FromString(t *testing.T) {
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
		// {
		// 	name: "convert from",
		// }
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt16(tt.args.source, tt.args.destination); got.Value.Interface().(int16) != int16(1234) {
				t.Errorf("convertToInt16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToInt16FromInt(t *testing.T) {
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
			name: "convert from int into int16 success",
			args: args{
				source:      reflect.ValueOf(int(1234)),
				destination: reflect.ValueOf(int16(0)),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(int16(1234))),
		},
		// {
		// 	name: "convert from",
		// }
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt16(tt.args.source, tt.args.destination); got.Value.Interface().(int16) != int16(1234) {
				t.Errorf("convertToInt16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToInt16FromStruct(t *testing.T) {
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
			name: "convert from struct into int16 failed",
			args: args{
				source:      reflect.ValueOf(struct{ field string }{field: "test"}),
				destination: reflect.ValueOf(int16(0)),
			},
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(int16(0)),
				errors.New("can not converted from this type: "+reflect.Struct.String()+" beacuse this type not supported")),
		},
		// {
		// 	name: "convert from",
		// }
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt16(tt.args.source, tt.args.destination); got.GetNotAValue().Error.Error() != tt.want.GetNotAValue().Error.Error() {
				t.Log("converted error: ", got.GetNotAValue().Error.Error())
				t.Log("expected error: ", tt.want.GetNotAValue().Error)
				t.Errorf("convertToInt16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToInt16FromStringFailed(t *testing.T) {
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
			name: "convert from string into int16 failed",
			args: args{
				source:      reflect.ValueOf("1234f"),
				destination: reflect.ValueOf(int16(0)),
			},
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(int16(0)),
				errors.New("can not converted to this type: "+reflect.Int16.String())),
		},
		// {
		// 	name: "convert from",
		// }
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt16(tt.args.source, tt.args.destination); got.GetNotAValue().Error.Error() != tt.want.GetNotAValue().Error.Error() {
				t.Log("converted error: ", got.GetNotAValue().Error.Error())
				t.Log("expected error: ", tt.want.GetNotAValue().Error)
				t.Errorf("convertToInt16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToInt32(t *testing.T) {
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
			name: "convert to int32 from string",
			args: args{
				source:      reflect.ValueOf("1234"),
				destination: reflect.ValueOf(int32(0)),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(int32(1234))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt32(tt.args.source, tt.args.destination); got.Value.Interface().(int32) != tt.want.Value.Interface().(int32) {
				t.Errorf("convertToInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToInt32FromInt(t *testing.T) {
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
			name: "convert to int32 from int",
			args: args{
				source:      reflect.ValueOf(int(1234)),
				destination: reflect.ValueOf(int32(0)),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(int32(1234))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt32(tt.args.source, tt.args.destination); got.Value.Interface().(int32) != tt.want.Value.Interface().(int32) {
				t.Errorf("convertToInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToInt32FromInt64(t *testing.T) {
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
			name: "convert to int32 from int",
			args: args{
				source:      reflect.ValueOf(int64(1234123123123123132)),
				destination: reflect.ValueOf(int32(0)),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(int32(450458556))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt32(tt.args.source, tt.args.destination); got.Value.Interface().(int32) != tt.want.Value.Interface().(int32) {
				t.Errorf("convertToInt32() = %v, want %v", got.Value.Interface(), tt.want.Value.Interface())
			}
		})
	}
}

func Test_convertToInt32FromStruct(t *testing.T) {
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
			name: "convert to int32 from struct",
			args: args{
				source:      reflect.ValueOf(struct{ field string }{field: "test"}),
				destination: reflect.ValueOf(int32(0)),
			},
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(int32(0)), errors.New("can not converted from this type: "+reflect.Struct.String()+" beacuse this type not supported")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt32(tt.args.source, tt.args.destination); got.GetNotAValue().Error.Error() != tt.want.GetNotAValue().Error.Error() {
				t.Errorf("convertToInt32() = %v, want %v", got.Value.Interface(), tt.want.Value.Interface())
			}
		})
	}
}

func Test_convertToInt32FromStringFailed(t *testing.T) {
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
			name: "convert to int32 from string failed",
			args: args{
				source:      reflect.ValueOf("1234f"),
				destination: reflect.ValueOf(int32(0)),
			},
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(int32(0)),
				errors.New("can not converted to this type: "+reflect.Int32.String()),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt32(tt.args.source, tt.args.destination); got.GetNotAValue().Error.Error() != tt.want.GetNotAValue().Error.Error() {
				t.Errorf("convertToInt32() = %v, want %v", got.Value.Interface(), tt.want.Value.Interface())
			}
		})
	}
}

func Test_convertToInt64(t *testing.T) {
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
			name: "convert from string into int64",
			args: args{
				source:      reflect.ValueOf("12345"),
				destination: reflect.ValueOf(int64(0)),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(int64(12345))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt64(tt.args.source, tt.args.destination); got.Value.Interface().(int64) != tt.want.Value.Interface().(int64) {
				t.Errorf("convertToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToInt64FromInt(t *testing.T) {
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
			name: "convert from int into int64",
			args: args{
				source:      reflect.ValueOf(int(12345)),
				destination: reflect.ValueOf(int64(0)),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(int64(12345))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt64(tt.args.source, tt.args.destination); got.Value.Interface().(int64) != tt.want.Value.Interface().(int64) {
				t.Errorf("convertToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToInt64FromStruct(t *testing.T) {
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
			name: "convert from struct into int64",
			args: args{
				source:      reflect.ValueOf(struct{ field string }{field: "test"}),
				destination: reflect.ValueOf(int64(0)),
			},
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(int64(0)), errors.New("can not converted from this type: "+reflect.Struct.String()+" beacuse this type not supported")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt64(tt.args.source, tt.args.destination); got.GetNotAValue().Error.Error() != tt.want.GetNotAValue().Error.Error() {
				t.Errorf("convertToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToInt64FromStringFailed(t *testing.T) {
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
			name: "convert from string into int64 failed",
			args: args{
				source:      reflect.ValueOf("1234124r"),
				destination: reflect.ValueOf(int64(0)),
			},
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(int64(0)),
				errors.New("can not converted to this type: "+reflect.Int64.String()),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToInt64(tt.args.source, tt.args.destination); got.GetNotAValue().Error.Error() != tt.want.GetNotAValue().Error.Error() {
				t.Errorf("convertToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToBool(t *testing.T) {
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
			name: "convert to bool from string",
			args: args{
				source:      reflect.ValueOf("true"),
				destination: reflect.ValueOf(false),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(true)),
		},
		{
			name: "convert to bool from string",
			args: args{
				source:      reflect.ValueOf("true1"),
				destination: reflect.ValueOf(false),
			},
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(false), errors.New("can not be converted from "+reflect.String.String()+" into bool")),
		},
		{
			name: "convert from bool to bool",
			args: args{
				source:      reflect.ValueOf(true),
				destination: reflect.ValueOf(false),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf(true)),
		},
		{
			name: "error convert from struct into bool",
			args: args{
				source:      reflect.ValueOf(struct{ field string }{field: "test"}),
				destination: reflect.ValueOf(false),
			},
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(false), errors.New("can not converted from this type: "+reflect.Struct.String()+" beacuse this type not supported")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToBool(tt.args.source, tt.args.destination); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToBoolFromString(t *testing.T) {
	type args struct {
		source      reflect.Value
		destination reflect.Value
	}
	tests := []struct {
		name string
		args args
		want pipeline.GoStructorValue
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToBool(tt.args.source, tt.args.destination); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToString(t *testing.T) {
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
			name: "convert to string from string",
			args: args{
				source:      reflect.ValueOf("1234"),
				destination: reflect.ValueOf(""),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf("1234")),
		},
		{
			name: "convert from int to string",
			args: args{
				source:      reflect.ValueOf(1234),
				destination: reflect.ValueOf(""),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf("1234")),
		},
		{
			name: "convert from bool to string",
			args: args{
				source:      reflect.ValueOf(true),
				destination: reflect.ValueOf(""),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf("true")),
		},
		{
			name: "convert from float32 to string",
			args: args{
				source:      reflect.ValueOf(float32(123.32)),
				destination: reflect.ValueOf(""),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf("1.2332E+02")),
		},
		{
			name: "convert from float64 to string",
			args: args{
				source:      reflect.ValueOf(float64(123.23)),
				destination: reflect.ValueOf(""),
			},
			want: pipeline.NewGoStructorTrueValue(reflect.ValueOf("1.2323E+02")),
		},
		{
			name: "convert from struct to string",
			args: args{
				source:      reflect.ValueOf(struct{ field string }{field: "test"}),
				destination: reflect.ValueOf(""),
			},
			want: pipeline.NewGoStructorNoValue(reflect.ValueOf(""),
				errors.New("can not converted from this type: "+reflect.Struct.String()+" beacuse this type not supported")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToString(tt.args.source, tt.args.destination); got.Value.String() != tt.want.Value.String() {
				t.Log(got.Value.Interface().(string))
				t.Errorf("convertToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
