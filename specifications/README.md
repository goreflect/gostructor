# Specification

[Back to main](https://github.com/goreflect/gostructor/blob/master)

## Base Scenarious

You can use this library for configurating your golang structures from any available sources. It is this feature that distinguishes this library from everyone else. 

## Priority configurating

### Default

For configure your structure you can choose the priority of configuration functions by adding a slice with function indices to the interface used.

### Custom

For configure your structure with custom order you can create this order in special tag:
```golang
cf_priority:"<local:cf_env, cf_default>,<cloud: cf_hocon,cf_vault,cf_default>"
```
After you should specify environment parameter for choosing current key priority like this:

For example: 
```bash
GOSTRUCTOR_PRIORITY=local
```