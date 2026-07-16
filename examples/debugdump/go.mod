module github.com/goreflect/gostructor/examples/debugdump

go 1.24

require (
	github.com/goreflect/gostructor v0.0.0
	github.com/goreflect/gostructor/watch v0.0.0
)

require (
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
)

replace github.com/goreflect/gostructor => ../../

replace github.com/goreflect/gostructor/watch => ../../watch
