module github.com/goreflect/gostructor/etcd

go 1.25.0

require (
	github.com/goreflect/gostructor v0.0.0
	github.com/goreflect/gostructor/snapshot v0.0.0
	go.etcd.io/etcd/client/v3 v3.5.16
)

require (
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	go.etcd.io/etcd/api/v3 v3.5.16 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.16 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/grpc v1.83.2 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

// Local-only until gostructor v1.0 and the snapshot module are tagged; Go
// ignores replace directives from non-main modules.
replace github.com/goreflect/gostructor => ../

replace github.com/goreflect/gostructor/snapshot => ../snapshot
