module github.com/goreflect/gostructor/examples/springcloud

go 1.24

require (
	github.com/goreflect/gostructor v0.0.0
	github.com/goreflect/gostructor/snapshot v0.0.0
	github.com/goreflect/gostructor/springcloud v0.0.0
)

replace github.com/goreflect/gostructor => ../../

replace github.com/goreflect/gostructor/springcloud => ../../springcloud

replace github.com/goreflect/gostructor/snapshot => ../../snapshot
