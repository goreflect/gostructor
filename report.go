package gostructor

import (
	"fmt"
	"strings"
)

// Report is a structured, per-field account of how Configure resolved a
// struct: which sources it tried for each field, in what order, which one
// won, and what value came out. Obtain one from ConfigureWithReport. It is
// nil-cost when unused - the plain Configure entry point builds no report.
//
// The Report is the machine view; String() renders the same data as a
// human-readable per-field tree suitable for logs or a support ticket.
// Values of fields marked cf_secret are masked everywhere they appear here.
type Report struct {
	// Fields holds one FieldResolution per struct field walked, in struct
	// order (nested fields flattened, matching resolution order).
	Fields []FieldResolution
}

// FieldResolution records the resolution of a single struct field: every
// source considered, the winner, and the raw and converted values.
type FieldResolution struct {
	Field    string    // struct field name
	Type     string    // Go type, e.g. "int", "time.Duration"
	Attempts []Attempt // every source considered, in the order tried
	Winner   string    // tag that produced the value, or "" if none
	Raw      any       // value the winning source returned (masked if secret)
	Value    any       // converted, field-typed value    (masked if secret)
	Outcome  string    // resolved | default | unresolved | error | skipped
}

// Attempt records a single source's turn at resolving a field.
type Attempt struct {
	Tag    string // e.g. "cf_env"
	Status string // not-found | used | skipped | error
	Detail string // the key looked up (env var, file key) or an error summary
}

const (
	outcomeResolved   = "resolved"
	outcomeDefault    = "default"
	outcomeUnresolved = "unresolved"
	outcomeError      = "error"
	outcomeSkipped    = "skipped"

	statusNotFound = "not-found"
	statusUsed     = "used"
	statusSkipped  = "skipped"
	statusError    = "error"
)

// Provenance projects the report down to the "field -> winning source tag"
// map: the dependency view of how this environment assembled the config.
// Fields that resolved from no source (unresolved or untagged) are omitted.
func (r *Report) Provenance() map[string]string {
	out := make(map[string]string, len(r.Fields))
	for _, f := range r.Fields {
		if f.Winner != "" {
			out[f.Field] = f.Winner
		}
	}
	return out
}

// String renders the report as a per-field tree, one field per line:
//
//	Port int  ⇐ cf_env(APP_PORT)=8080  [cf_json: not-found, cf_default: skipped]
//	Host string  ⇐ cf_default=0.0.0.0
//	Secret string  ⇐ cf_env(API_KEY)=••••••
func (r *Report) String() string {
	var b strings.Builder
	for i, f := range r.Fields {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "%s %s  ", f.Field, f.Type)
		switch f.Outcome {
		case outcomeResolved, outcomeDefault:
			// Winner is the last attempt with status "used".
			winDetail := f.Winner
			for _, a := range f.Attempts {
				if a.Status == statusUsed {
					if a.Detail != "" && a.Detail != f.Winner {
						winDetail = fmt.Sprintf("%s(%s)", a.Tag, a.Detail)
					} else {
						winDetail = a.Tag
					}
					break
				}
			}
			fmt.Fprintf(&b, "⇐ %s=%v", winDetail, f.Value)
		case outcomeUnresolved:
			b.WriteString("⇐ unresolved")
		case outcomeError:
			b.WriteString("⇐ error")
		case outcomeSkipped:
			b.WriteString("⇐ (untagged, skipped)")
		}
		if losing := losingAttempts(f); losing != "" {
			fmt.Fprintf(&b, "  [%s]", losing)
		}
	}
	return b.String()
}

// losingAttempts renders the non-winning attempts as "tag: status" pairs.
func losingAttempts(f FieldResolution) string {
	var parts []string
	for _, a := range f.Attempts {
		if a.Status == statusUsed {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %s", a.Tag, a.Status))
	}
	return strings.Join(parts, ", ")
}
