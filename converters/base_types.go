package converters

import (
	"errors"
	"reflect"
	"strconv"

	"github.com/goreflect/gostructor/infra"
	logrus "github.com/sirupsen/logrus"
)

func convertToInt(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		logrus.Debug("start convert value ", source.String(), " into int type")
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, 64)
		if errorConvert != nil {
			return infra.NewGoStructorNoValue(destination, errors.New("can not be converted to this type: "+destination.Kind().String()))
		}
		return infra.NewGoStructorTrueValue(reflect.ValueOf(int(convertedValue)))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		logrus.Debug("start convert value ", source.String(), " into int type")
		if source.Type().ConvertibleTo(destination.Type()) {
			return infra.NewGoStructorTrueValue(source.Convert(destination.Type()))
		}
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into int"))
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from this type: "+source.Kind().String()+" because this type not supported"))
	}
}

func convertToIntOrder(source reflect.Value, destination reflect.Value, order int) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		logrus.Debug("start convert value ", source.String(), " into int64 type")
		convertedValue, errorConvert := strconv.ParseInt(source.String(), 10, order)
		if errorConvert != nil {
			return infra.NewGoStructorNoValue(destination, errors.New("can not be converted to this type: "+destination.Kind().String()))
		}
		return infra.NewGoStructorTrueValue(chooseIntByOrder(int64(convertedValue), order))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		logrus.Debug("start convert value ", source.String(), " into int64 type")
		if source.Type().ConvertibleTo(destination.Type()) {
			return infra.NewGoStructorTrueValue(source.Convert(destination.Type()))
		}
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into int64"))
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from this type: "+source.Kind().String()+" because this type not supported"))
	}
}

func chooseIntByOrder(value int64, order int) reflect.Value {
	switch order {
	case 8:
		return reflect.ValueOf(int8(value))
	case 16:
		return reflect.ValueOf(int16(value))
	case 32:
		return reflect.ValueOf(int32(value))
	default:
		return reflect.ValueOf(value)
	}
}

func convertToBool(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		converted, errConvert := strconv.ParseBool(source.String())
		if errConvert != nil {
			return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from "+source.Kind().String()+" into bool"))
		}
		return infra.NewGoStructorTrueValue(reflect.ValueOf(converted))
	case reflect.Bool:
		return infra.NewGoStructorTrueValue(source)
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from this type: "+source.Kind().String()+" because this type not supported"))
	}
}

func convertToString(source reflect.Value, destination reflect.Value) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.String:
		return infra.NewGoStructorTrueValue(source)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return infra.NewGoStructorTrueValue(reflect.ValueOf(
			strconv.FormatInt(source.Int(), 10),
		))
	case reflect.Float32:
		return infra.NewGoStructorTrueValue(reflect.ValueOf(
			strconv.FormatFloat(source.Float(), 'E', -1, 32),
		))
	case reflect.Float64:
		return infra.NewGoStructorTrueValue(reflect.ValueOf(
			strconv.FormatFloat(source.Float(), 'E', -1, 64),
		))
	case reflect.Bool:
		return infra.NewGoStructorTrueValue(reflect.ValueOf(
			strconv.FormatBool(source.Bool()),
		))
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from this type: "+source.Kind().String()+" because this type not supported"))
	}
}

func convertToFloatOrder(source reflect.Value, destination reflect.Value, order int) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return infra.NewGoStructorTrueValue(choseTypeByOrder(float64(source.Int()), order))
	case reflect.String:
		parsed, errParsed := strconv.ParseFloat(source.String(), order)
		if errParsed != nil {
			return infra.NewGoStructorNoValue(destination, errParsed)
		}
		return infra.NewGoStructorTrueValue(choseTypeByOrder(float64(parsed), order))
	case reflect.Float32, reflect.Float64:
		return infra.NewGoStructorTrueValue(choseTypeByOrder(source.Float(), order))
	default:
		return infra.NewGoStructorNoValue(destination, errors.New("can not be converted from this type: "+source.Kind().String()+" because this type not supported"))
	}
}

func choseTypeByOrder(value float64, order int) reflect.Value {
	if order == 32 {
		return reflect.ValueOf(float32(value))
	}
	return reflect.ValueOf(value)
}

func convertToUintOrder(source reflect.Value, destination reflect.Value, order int) infra.GoStructorValue {
	switch source.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return infra.NewGoStructorTrueValue(chooseTypeByOrder(source.Uint(), order))
	case reflect.String:
		parsed, errParsed := strconv.ParseUint(source.String(), 10, order)
		if errParsed != nil {
			return infra.NewGoStructorNoValue(destination, errParsed)
		}
		return infra.NewGoStructorTrueValue(chooseTypeByOrder(parsed, order))
	default:
		return infra.NewGoStructorNoValue(source, errors.New("not supported convertation"))
	}
}

func chooseTypeByOrder(value uint64, order int) reflect.Value {
	switch order {
	case 8:
		return reflect.ValueOf(uint8(value))
	case 16:
		return reflect.ValueOf(uint16(value))
	case 32:
		return reflect.ValueOf(uint32(value))
	default:
		return reflect.ValueOf(value)
	}
}
