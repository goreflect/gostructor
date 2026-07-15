module github.com/goreflect/gostructor/yaml

go 1.24

require (
	github.com/goccy/go-yaml v1.9.5
	github.com/goreflect/gostructor v0.0.0
)

require (
	github.com/fatih/color v1.10.0 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

// Local-only: gostructor v1.0 hasn't been tagged yet. `replace` is ignored
// by consumers (Go only applies replace directives from the main module of
// a build), so this only affects building this submodule directly out of
// this checkout. Bump the require above to a real tag and delete this line
// once gostructor v1.0 is released.
replace github.com/goreflect/gostructor => ../
