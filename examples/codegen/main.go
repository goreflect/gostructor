package main

//go:generate go run github.com/goreflect/gostructor/cmd/gostructor-gen -dir .

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/goreflect/gostructor"
)

// AppConfig is the root config. The //gostructor:gen marker tells gostructor-gen
// to generate a Fill for it; the tool only needs the directory (or, across a
// whole project, `-dir . -recursive`).
//
//gostructor:gen
type AppConfig struct {
	Service ServiceConfig // untagged nested struct: flattened into its leaves
	DB      DatabaseConfig

	// Compound leaves come from an object source (here the in-memory Map); env
	// and default are flat string sources, so these stay optional.
	Labels map[string]string `cfg:"labels" gos:"optional"`
	Ports  []int             `cfg:"ports" gos:"optional"`

	Tags []string `cfg:"tags" gos:"default:a|b,sep:|"`
}

type ServiceConfig struct {
	Name    string        `cfg:"svcName" gos:"default:api"`
	Port    int           `cfg:"svcPort,env:SVC_PORT" gos:"default:8080"`
	Timeout time.Duration `cfg:"svcTimeout" gos:"default:5s"`
}

type DatabaseConfig struct {
	Host     string `cfg:"dbHost" gos:"default:localhost"`
	Replicas uint   `cfg:"dbReplicas" gos:"default:2"`
}

func main() {
	os.Setenv("SVC_PORT", "9090")

	// A highest-priority object source supplies the nested host and the compound
	// map/slice fields; env overrides the service port; defaults fill the rest.
	overrides := gostructor.Map("map", map[string]any{
		"dbHost": "db.internal",
		"labels": map[string]any{"team": "core", "tier": "gold"},
		"ports":  []any{8080, 8081, 8082},
	})

	// Adaptive by default: because AppConfig has a generated Fill, this runs the
	// reflection-free path. Drop the generated file (or force EngineReflection)
	// and the exact same call keeps working through the reflective engine.
	cfg, err := gostructor.Configure(&AppConfig{},
		gostructor.WithSources(overrides, gostructor.Env(), gostructor.Default()))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", *cfg)
}
