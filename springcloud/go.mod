module github.com/goreflect/gostructor/springcloud

go 1.24

require (
	github.com/goreflect/gostructor v0.0.0
	github.com/goreflect/gostructor/snapshot v0.0.0
)

// Local-only until gostructor v1.0 and the snapshot module are tagged; Go
// ignores replace directives from non-main modules.
replace github.com/goreflect/gostructor => ../

replace github.com/goreflect/gostructor/snapshot => ../snapshot
