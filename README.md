# Gostructor [![Actions Status](https://github.com/goreflect/gostructor/workflows/CI_dev/badge.svg)](https://github.com/goreflect/gostructor/actions?query=workflow%3ACI_dev) [![Go Report Card](https://goreportcard.com/badge/github.com/goreflect/gostructor)](https://goreportcard.com/report/github.com/goreflect/gostructor) [![codecov](https://codecov.io/gh/goreflect/gostructor/branch/master/graph/badge.svg)](https://codecov.io/gh/goreflect/gostructor)
____
### Version: 0.1.2

hocon current configuration configuration library

## Current supporting input formats

- hocon

## Current supporting types

- int32, int64
- float32, float64
- string
- bool
- map[string\int]string\int\float32\float64
- slices of any types from (int32, int64, int, string, bool, float32, float64)

## Tags

In current library using any of this tags:

1. cf_hocon - setup value for this field from hocon
2. cf_default - setup default value for this field
3. cf_env - setup value from env variable by name in this tag
4. cf_yaml - setup value for this field from yaml (version > 0.2)
5. cf_json - setup value for this field from json (version > 0.3)
6. cf_server - setup value from configuration server like spring cloud config server or others (version>0.4)
7.  cf_vault - setup secret for this field from hashi corp vault (version>0.5)

## Validation(optional version > 0.6)

If you have validate your data you should write in source tag validate after comma and after : you can choose validation type: 
- email
- phoneNumber8

## Menu

- [Reserved Tags](https://github.com/goreflect/gostructor/blob/master/tags)
- [Specifications](https://github.com/goreflect/gostructor/blob/master/specifications)

## Running configuring by easy way

For easy way you can chose gostructor.ConfigureEasy(). And if you have use one of file sources like: json, hocon, yaml, txt your files should ended by type (if your file not ended of type, this source will be ignored by function chainer)

## Running configuring by setup way

For this way you should set up chain functions will used in configuring pipeline. For example: 
`test.hocon`
```hocon
Examples = {
    FieldExample = test1

    ExampleType = {
        Field1 = test2
    }
}
```

```golang
type ExampleType struct {
    fieldTest1  string `cf_hocon:"Examples.FieldExample"`
    Field1      string `cf_default:"tururutest"`
}
...
value, err := gostructor.ConfigureSetup(&ExampleType{}, "test.hocon", []pipeline.FuncType{
    pipeline.FunctionSetupHocon,
    pipeline.FunctionSetupDefault,
})

```

Where Value is Interface therefore you should cast to your type with Pointer like this:
```golang
value.(*ExampleType)
```


## Many examples

In this section will added any examples of using this library

## TODO

1. Implement self converters for all sources
2. Write unit tests for any cases
3. Move getFieldName methods for any source func parsers because in any of source have available any naming sources