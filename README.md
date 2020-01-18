# Go_structor [![Actions Status](https://github.com/goreflect/go_structor/workflows/CI_dev/badge.svg)](https://github.com/goreflect/go_structor/actions?query=workflow%3ACI_dev) [![Go Report Card](https://goreportcard.com/badge/github.com/goreflect/go_structor)](https://goreportcard.com/report/github.com/goreflect/go_structor)

hocon current configuration configuration library

## Current supporting input formats

- hocon
- environment variables
- default values from tag in structures

### Version: 0.1

## Tags

In current library using any of this tags:

1. cf_hocon - setup value for this field from hocon
2. cf_default - setup default value for this field
3. cf_validate - setup validation function for this field
4. cf_env - setup value from env variable by name in this tag
5. cf_yaml - setup value for this field from yaml (version > 0.2)
6. cf_json - setup value for this field from json (version > 0.3)
7. cf_server - setup value from configuration server like spring cloud config server or others (version>0.4)
8.  cf_vault - setup secret for this field from hashi corp vault (version>0.5)

## Menu

- [Reserved Tags](https://github.com/goreflect/go_structor/blob/master/tags)
- [Specifications](https://github.com/goreflect/go_structor/blob/master/specifications)
