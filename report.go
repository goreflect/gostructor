package gostructor

import (
	"fmt"
	"sort"
	"strings"
)

// Report is a per-field account of how Configure resolved a struct: which
// sources it tried for each field, in what order, which won, and what value
// came out. Obtain one from ConfigureWithReport; the plain Configure builds
// no report, so it is free when unused.
//
// Report is the machine view. String() renders a focused summary of the
// actionable data: the primary source, the defaults, and an Overrides &
// Secrets section. Secret fields (gos:"secret") are masked throughout.
type Report struct {
	// Type is the configured struct's Go type, e.g. "main.Config".
	Type string
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
	Winner   string    // source name that produced the value, or "" if none
	Raw      any       // value the winning source returned (masked if secret)
	Value    any       // converted, field-typed value    (masked if secret)
	Outcome  string    // resolved | default | unresolved | error | skipped
	IsSecret bool      // field carries gos:"secret"
}

// Attempt records a single source's turn at resolving a field.
type Attempt struct {
	Source string // e.g. "env"
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

// Provenance projects the report down to the "field -> winning source name"
// map: the dependency view of how this environment assembled the config.
// Fields that resolved from no source (unresolved or unconfigured) are omitted.
func (r *Report) Provenance() map[string]string {
	out := make(map[string]string, len(r.Fields))
	for _, f := range r.Fields {
		if f.Winner != "" {
			out[f.Field] = f.Winner
		}
	}
	return out
}

// PrimarySource reports the source that resolved the most non-secret fields,
// i.e. the environment's main configuration origin (typically a YAML/JSON/TOML
// file). The Default source is never the primary. It returns "" when no
// non-default source won a non-secret field.
func (r *Report) PrimarySource() string {
	counts := map[string]int{}
	for _, f := range r.Fields {
		if f.Winner == "" || f.Winner == SourceDefault || f.IsSecret {
			continue
		}
		counts[f.Winner]++
	}
	best, bestN := "", 0
	// Sort source names for a stable winner on ties.
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if counts[name] > bestN {
			best, bestN = name, counts[name]
		}
	}
	return best
}

// String renders a focused resolution trace: a one-line summary, a one-liner
// for the primary source and for defaults, then an Overrides & Secrets section
// listing only the fields that were resolved by a non-primary, non-default
// source (an override) or that are secret. Fields resolved straight from the
// primary source or from a plain default are summarized by count, not listed.
//
//	Configuring main.Config: 5 fields
//	[Primary Source] json (loaded 3 fields)
//	[Defaults] applied for 1 field
//
//	Overrides & Secrets:
//	  Port     int    ⇐ env (override; json had it too)
//	  Password string ⇐ vault  •••••••• (secret)
func (r *Report) String() string {
	primary := r.PrimarySource()

	var (
		primaryCount int
		defaultCount int
		highlights   []FieldResolution
	)
	for _, f := range r.Fields {
		switch {
		case f.IsSecret && f.Winner != "":
			highlights = append(highlights, f)
		case f.Winner == "" || f.Outcome == outcomeSkipped:
			// unresolved / optional / unconfigured: not actionable here.
		case f.Winner == SourceDefault:
			defaultCount++
		case f.Winner == primary:
			primaryCount++
		default:
			// resolved by a non-primary, non-default source: an override.
			highlights = append(highlights, f)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Configuring %s: %s", orUnknownType(r.Type), pluralFields(len(r.Fields)))
	if primary != "" {
		fmt.Fprintf(&b, "\n[Primary Source] %s (loaded %s)", primary, pluralFields(primaryCount))
	}
	if defaultCount > 0 {
		fmt.Fprintf(&b, "\n[Defaults] applied for %s", pluralFields(defaultCount))
	}

	if len(highlights) == 0 {
		return b.String()
	}
	b.WriteString("\n\nOverrides & Secrets:")
	nameW, typeW := highlightWidths(highlights)
	for _, f := range highlights {
		name := fmt.Sprintf("%-*s %-*s", nameW, f.Field, typeW, f.Type)
		if f.IsSecret {
			// f.Value is already masked (via the configured Masker), so the
			// real secret never reaches this string.
			fmt.Fprintf(&b, "\n  %s  ⇐ %s = %v (secret)", name, f.Winner, f.Value)
			continue
		}
		fmt.Fprintf(&b, "\n  %s  ⇐ %s (override) = %v", name, f.Winner, f.Value)
	}
	return b.String()
}

// highlightWidths returns the column widths to align the field name and type
// columns in the Overrides & Secrets section.
func highlightWidths(fields []FieldResolution) (nameW, typeW int) {
	for _, f := range fields {
		if len(f.Field) > nameW {
			nameW = len(f.Field)
		}
		if len(f.Type) > typeW {
			typeW = len(f.Type)
		}
	}
	return nameW, typeW
}

func pluralFields(n int) string {
	if n == 1 {
		return "1 field"
	}
	return fmt.Sprintf("%d fields", n)
}

func orUnknownType(t string) string {
	if t == "" {
		return "struct"
	}
	return t
}
