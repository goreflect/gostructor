// Command errors tours gostructor's typed error taxonomy: it triggers each
// failure mode and classifies it with errors.Is / errors.As, so a caller can
// tell what went wrong and whose fault it is.
//
// Run it:
//
//	go run ./examples/errors
package main

import (
	"errors"
	"fmt"

	"github.com/goreflect/gostructor"
)

func main() {
	classify("unresolved field (no source produced a value)", func() error {
		type C struct {
			APIKey string `cfg:"apiKey,env:DEFINITELY_UNSET_VAR_XYZ"`
		}
		_, err := gostructor.Configure(&C{})
		return err
	})

	classify("convert error (value doesn't fit the field type)", func() error {
		type C struct {
			Port int `gos:"default:not-a-number"`
		}
		_, err := gostructor.Configure(&C{})
		return err
	})

	classify("source error (backing store failed)", func() error {
		type C struct {
			Name string `cfg:"name,json:service.name"`
		}
		// Point the JSON source at a file that doesn't exist.
		_, err := gostructor.Configure(&C{},
			gostructor.WithSources(gostructor.JSONFile("/no/such/config.json")))
		return err
	})

	classify("hook error (validation rejected the value)", func() error {
		type C struct {
			Port int `gos:"default:80"`
		}
		_, err := gostructor.Configure(&C{},
			gostructor.WithHook(func(_ gostructor.FieldContext, v any) (any, error) {
				return nil, fmt.Errorf("port %v is privileged", v)
			}))
		return err
	})

	classify("invalid target (a programming bug, not a config gap)", func() error {
		var nilPtr *struct{ X int }
		_, err := gostructor.Configure(nilPtr)
		return err
	})
}

// classify prints how each error is recognised through the public taxonomy.
func classify(label string, run func() error) {
	err := run()
	fmt.Printf("• %s\n  err: %v\n", label, err)

	switch {
	case err == nil:
		fmt.Println("  (no error?!)")
	case errors.Is(err, gostructor.ErrInvalidTarget):
		fmt.Println("  → ErrInvalidTarget: the call itself is wrong; fix the code.")
	case errors.Is(err, gostructor.ErrFieldNotResolved):
		var e *gostructor.NotResolvedError
		errors.As(err, &e)
		fmt.Printf("  → NotResolvedError on %q, tried %v\n", e.Field, e.Sources)
	default:
		// Every field-scoped error implements FieldError, so we can always
		// recover which field failed without knowing the concrete type.
		var fe gostructor.FieldError
		if errors.As(err, &fe) {
			var (
				ce *gostructor.ConvertError
				se *gostructor.SourceError
				he *gostructor.HookError
			)
			switch {
			case errors.As(err, &ce):
				fmt.Printf("  → ConvertError on %q (bad value in config)\n", fe.FieldName())
			case errors.As(err, &se):
				fmt.Printf("  → SourceError on %q via %s (backing store)\n", fe.FieldName(), se.Source)
			case errors.As(err, &he):
				fmt.Printf("  → HookError on %q (your validation)\n", fe.FieldName())
			}
		}
	}
	fmt.Println()
}
