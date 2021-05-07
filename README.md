
# Gostructor [![Actions Status](https://github.com/goreflect/gostructor/workflows/CI_dev/badge.svg)](https://github.com/goreflect/gostructor/actions?query=workflow%3ACI_dev) [![Go Report Card](https://goreportcard.com/badge/github.com/goreflect/gostructor)](https://goreportcard.com/report/github.com/goreflect/gostructor) [![codecov](https://codecov.io/gh/goreflect/gostructor/branch/master/graph/badge.svg)](https://codecov.io/gh/goreflect/gostructor)

____

## Version: v0.7

Universal configuration library by tags

## Current supporting input formats

- hocon values
- default values
- environment variables
- vault configs
- json values
- yaml values
- ini values (developing)
- toml values (developing)

## Current supporting types

- int, int8, int16, int32, int64
- uint, uint8, uint16, uint32, uint64
- float32, float64
- string
- bool
- map[string\int]string\int\float32\float64
- slices of any types from (int32, int64, int, string, bool, float32, float64)

## Plan of upgrading
0.8:
1. Adding support file store fetching
0.9:
1. adding support key\value store backends (with callback)
2. adding support config server fetching content like Spring Cloud Config Server

1.0:
1. adding support uint types (done 0.6.6)
2. change transition dependencies like go-yaml, lfjson by native map[string]interface{}\string

## Ideas for future

1. Live watching for contract by git (maybe mechanism for watching changes like do it spring)
2. CodeGen plugin for protoc for generating models with predefined tags

### Tags

- [x] cf_hocon - setup value for this field from hocon
- [x] cf_default - setup default value for this field
- [x] cf_env - setup value from env variable by name in this tag
- [x] cf_yaml - setup value for this field from yaml 
- [x] cf_json - setup value for this field from json
- [x] cf_ini - setup value for this field from ini (version > 0.7)
- [x] cf_toml - setup value for this field from toml (version > 0.7)
- [ ] cf_server_file - setup value from configuration server like spring cloud config server or others (version>0.8)
- [ ] cf_server_kv - setup value from configuration key\value store (version > 0.9)
- [x] cf_vault - setup secret for this field from hashi corp vault

## Running configuring by smart variant

For Run configuration by smart variant autostart analysing of using tags. Library will start configuring  your structure by pipeline with all founded tags.

```go
type Test struct {
    MyValue1 string `cf_default:"turur" cf_hocon:"mySourceValue1"`
    MySlice1 []bool `cf_default:"true,false,false,true" cf_env:"MY_SIGNALS"`
}

// in this example do use 3 tags: cf_default (using default values which setup inline tag)
// cf_env - using environment variable
// cf_hocon - using hocon source file 

//....

func myConfigurator() {
    os.Setenv(tags.HoconFile, , "testhocon.hocon")
    myStruct, errConfiguring := gostructir.ConfigureSmart(&Test{})
    // check errConfiguring for any errors
    if errConfiguring != nil {
        /// action for error
    }

    // cast interface{} into Test structure
    myValues := myStruct.(*Test)
    // now, u structure already filled
} 

```

## Running configuring by setup

You can also setting configuring pipeline like this:

```go
type Test struct {
    MyValue1 string `cf_default:"turur" cf_hocon:"mySourceValue1"`
    MySlice1 []bool `cf_default:"true,false,false,true" cf_env:"MY_SIGNALS"`
}

func myConfigurator() {
    os.Setenv(tags.HoconFile, , "testhocon.hocon")
    myStruct, errConfiguring := gostructir.ConfigureSetup(&Test{}, []infra.FuncType{
        infra.FunctionSetupEnvironment,
    })// you should setup only by order configure
    // check errConfiguring for any errors
    if errConfiguring != nil {
        /// action for error
    }

    // cast interface{} into Test structure
    myValues := myStruct.(*Test)
    // now, u structure already filled
} 

```

## Fetching secrets from vault

For fetching secrets you should add 2 environment variables: VAULT_ADDRESS, VAULT_TOKEN. After all, you can add cf_vault tag into your structure tags. 

Now vault configure support all basic types and also complex type: slice

For Example:

```go
type Test struct {
    MySecretKey string `cf_vault:"my-secret-service/stage/tururu#my-key"`
    MyCustomIntKey int16 `cf_vault:"my-secret-service/stage/tururur#my-key2"`
    TestSda []int32 `cf_vault:"my-secret-service/stage/tururu#my-key3"`
}
```

## Setting up files or any other sources

By the way u can configuring files by environment variables:

1. For hocon files - GOSTRUCTOR_HOCON
2. For json files - GOSTRUCTOR_JSON
3. For yaml files - GOSTRUCTOR_YAML
4. For ini files - GOSTRUCTOR_INI
5. For toml files - GOSTRUCTOR_TOML

## Infrastructure

For the best way to automatic publish versions of patch added github workflow for publish in master changes