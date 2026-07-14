package gostructor

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Errors returned by Configure fall into a small, closed set of categories so
// a caller can tell exactly what went wrong and whose fault it is:
//
//   - ErrInvalidTarget - the call itself is wrong (target isn't a non-nil
//     pointer to a struct). A programming bug; nothing an operator can fix at
//     runtime.
//   - *NotResolvedError (wraps ErrFieldNotResolved) - a field carries source
//     tags but none of them produced a value. Usually a missing env var,
//     file key, or secret: an operational/config gap.
//   - *SourceError - a source failed while producing a value (file missing or
//     malformed, Vault unreachable, ...). The backing store is the problem.
//   - *ConvertError - a value was produced but doesn't fit the field's Go
//     type (a fractional float into an int, an overflow, a bad duration).
//     The value in the config is the problem.
//   - *HookError - a WithHook callback rejected or mistyped the value. Your
//     validation/transformation is the problem.
//
// Match the sentinels with errors.Is and the struct types with errors.As;
// every struct type unwraps to its underlying Cause, and every field-scoped
// type satisfies FieldError so you can recover the field name uniformly:
//
//	cfg, err := gostructor.Configure(&Config{})
//	switch {
//	case err == nil:
//		// ok
//	case errors.Is(err, gostructor.ErrFieldNotResolved):
//		var nre *gostructor.NotResolvedError
//		errors.As(err, &nre)
//		log.Fatalf("no value for %s (tried %s)", nre.Field, strings.Join(nre.Tags, ", "))
//	default:
//		var fe gostructor.FieldError
//		if errors.As(err, &fe) {
//			log.Fatalf("field %s: %v", fe.FieldName(), err)
//		}
//		log.Fatal(err)
//	}

// ErrInvalidTarget is returned, wrapped, when Configure is called with a
// target that is not a non-nil pointer to a struct. Match it with errors.Is.
var ErrInvalidTarget = errors.New("gostructor: target must be a non-nil pointer to a struct")

// ErrFieldNotResolved is the sentinel wrapped by *NotResolvedError: a field
// carried at least one recognized source tag, but no configured source
// produced a value for it. Match it with errors.Is:
//
//	if errors.Is(err, gostructor.ErrFieldNotResolved) { ... }
var ErrFieldNotResolved = errors.New("no configured source produced a value")

// FieldError is implemented by every Configure error scoped to one struct
// field - NotResolvedError, SourceError, ConvertError, and HookError. It lets
// a caller recover which field failed without switching on the concrete type:
//
//	var fe gostructor.FieldError
//	if errors.As(err, &fe) { ... fe.FieldName() ... }
type FieldError interface {
	error
	FieldName() string
}

// NotResolvedError reports that a field carried recognized source tags but
// none produced a value. Tags lists the source tags that were present and
// tried, so the message can say precisely what was looked at (e.g. tried
// cf_env, cf_json). It wraps ErrFieldNotResolved.
type NotResolvedError struct {
	Field string   // struct field name
	Tags  []string // the source tags present on the field, in the order tried
}

func (e *NotResolvedError) Error() string {
	if len(e.Tags) == 0 {
		return fmt.Sprintf("gostructor: field %q: %v", e.Field, ErrFieldNotResolved)
	}
	return fmt.Sprintf("gostructor: field %q: %v (tried %s)", e.Field, ErrFieldNotResolved, strings.Join(e.Tags, ", "))
}

func (e *NotResolvedError) Unwrap() error { return ErrFieldNotResolved }

// FieldName reports the struct field this error concerns.
func (e *NotResolvedError) FieldName() string { return e.Field }

// SourceError reports that a configured Source failed while resolving a field
// - a JSON file that doesn't exist or doesn't parse, a Vault server that
// can't be reached, and so on. Tag names the failing source (e.g. "cf_json")
// so you can tell which of several sources on a field went wrong. It unwraps
// to the error the Source itself returned.
type SourceError struct {
	Field string // struct field name being resolved
	Tag   string // the failing source's tag, e.g. "cf_json"
	Cause error  // the error the Source returned
}

func (e *SourceError) Error() string {
	return fmt.Sprintf("gostructor: field %q: source %s failed: %v", e.Field, e.Tag, e.Cause)
}

func (e *SourceError) Unwrap() error { return e.Cause }

// FieldName reports the struct field this error concerns.
func (e *SourceError) FieldName() string { return e.Field }

// ConvertError reports a failure to convert a resolved raw value into a
// struct field's Go type - a fractional float into an int field, an
// out-of-range number, a malformed duration string, and so on. It carries
// the offending value and target type for diagnostics and unwraps to the
// underlying cause (a *strconv.NumError, a time.ParseDuration error, a
// TextUnmarshaler error, ...), reachable with a further errors.As.
type ConvertError struct {
	Field  string       // struct field name being filled
	Value  any          // the raw value a source produced
	Target reflect.Type // the field's Go type
	Cause  error        // the underlying conversion failure, if any
}

func (e *ConvertError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("gostructor: field %q: cannot convert %#v into %s: %v", e.Field, e.Value, e.Target, e.Cause)
	}
	return fmt.Sprintf("gostructor: field %q: cannot convert %#v into %s", e.Field, e.Value, e.Target)
}

func (e *ConvertError) Unwrap() error { return e.Cause }

// FieldName reports the struct field this error concerns.
func (e *ConvertError) FieldName() string { return e.Field }

// HookError reports that a WithHook callback rejected or failed to transform a
// field's value. A hook that returns a non-nil error produces a HookError
// wrapping it; a hook that returns a value of the wrong type produces a
// HookError whose Cause explains the type mismatch. It unwraps to that cause.
type HookError struct {
	Field string // struct field name the hook ran on
	Cause error  // the hook's error, or a type-mismatch description
}

func (e *HookError) Error() string {
	return fmt.Sprintf("gostructor: field %q: hook rejected value: %v", e.Field, e.Cause)
}

func (e *HookError) Unwrap() error { return e.Cause }

// FieldName reports the struct field this error concerns.
func (e *HookError) FieldName() string { return e.Field }
