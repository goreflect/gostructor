module github.com/goreflect/gostructor/watch

go 1.24

require (
	github.com/fsnotify/fsnotify v1.8.0
	github.com/goreflect/gostructor v0.0.0
)

require golang.org/x/sys v0.13.0 // indirect

// Local-only until gostructor v1.0 is tagged; Go ignores replace directives
// from non-main modules.
replace github.com/goreflect/gostructor => ../
