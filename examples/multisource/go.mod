module github.com/goreflect/gostructor/examples/multisource

go 1.24

require (
	github.com/goreflect/gostructor v0.0.0
	github.com/goreflect/gostructor/yaml v0.0.0
)

require (
	github.com/fatih/color v1.10.0 // indirect
	github.com/goccy/go-yaml v1.9.5 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

replace github.com/goreflect/gostructor => ../../

replace github.com/goreflect/gostructor/yaml => ../../yaml
