module github.com/goreflect/gostructor/consul

go 1.24

require (
	github.com/goreflect/gostructor v0.0.0
	github.com/goreflect/gostructor/snapshot v0.0.0
	github.com/hashicorp/consul/api v1.32.0
)

require (
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8 // indirect
	golang.org/x/sys v0.29.0 // indirect
)

// Local-only until gostructor v1.0 and the snapshot module are tagged; Go
// ignores replace directives from non-main modules, so these affect only builds
// rooted in this checkout.
replace github.com/goreflect/gostructor => ../

replace github.com/goreflect/gostructor/snapshot => ../snapshot
