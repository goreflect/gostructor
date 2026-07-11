module github.com/goreflect/gostructor/hocon

go 1.24

require github.com/goreflect/gostructor v0.0.0

// Local-only: gostructor v1.0 hasn't been tagged yet. `replace` is ignored
// by consumers (Go only applies replace directives from the main module of
// a build), so this only affects building this submodule directly out of
// this checkout. Bump the require above to a real tag and delete this line
// once gostructor v1.0 is released.
replace github.com/goreflect/gostructor => ../
