# Go_structor [![Actions Status](https://github.com/goreflect/go_structor/workflows/CI_dev/badge.svg)](https://github.com/goreflect/go_structor/actions?query=workflow%3ACI_dev) [![Go Report Card](https://goreportcard.com/badge/github.com/goreflect/go_structor)](https://goreportcard.com/report/github.com/goreflect/go_structor)
____
### Version: 0.1

hocon current configuration configuration library

## Current supporting input formats

- hocon
- environment variables
- default values from tag in structures

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

- [Reserved Tags](https://github.com/goreflect/go_structor/blob/master/tags)
- [Specifications](https://github.com/goreflect/go_structor/blob/master/specifications)

## Running configuring by easy way

For easy way you can chose go_structor.ConfigureEasy(). And if you have use one of file sources like: json, hocon, yaml, txt your files should ended by type (if your file not ended of type, this source will be ignored by function chainer)

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
    fieldTest1  string `cfg_hocon:"Examples.FieldExample"`
    Field1      string `cfg_default:"tururutest"`
}
...
go_structor.ConfigureSetup(&ExampleType{}, "test.hocon", []pipeline.FuncType{
    pipeline.FunctionSetupHocon,
    pipeline.FunctionSetupDefault,
})

```


## Many examples

In this section will added any examples of using this library