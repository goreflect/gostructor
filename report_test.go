package gostructor

import (
	"os"
	"strings"
	"testing"
)

type reportConfig struct {
	Host   string `cfg:"host,env:GOSTRUCTOR_RPT_HOST,json:host" gos:"default:0.0.0.0"`
	Port   int    `cfg:"port,env:GOSTRUCTOR_RPT_PORT,json:port" gos:"default:8080"`
	Secret string `cfg:"secret,env:GOSTRUCTOR_RPT_SECRET" gos:"secret"`
}

func findField(r *Report, name string) *FieldResolution {
	for i := range r.Fields {
		if r.Fields[i].Field == name {
			return &r.Fields[i]
		}
	}
	return nil
}

// writeJSON writes a JSON config file and returns its path, for reports that
// need a primary file source behind the env overrides.
func writeReportJSON(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := dir + "/report.json"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestConfigureWithReportRecordsWinnerAndLosers(t *testing.T) {
	os.Setenv("GOSTRUCTOR_RPT_PORT", "9090")
	defer os.Unsetenv("GOSTRUCTOR_RPT_PORT")
	os.Unsetenv("GOSTRUCTOR_RPT_HOST")
	os.Setenv("GOSTRUCTOR_RPT_SECRET", "hunter2")
	defer os.Unsetenv("GOSTRUCTOR_RPT_SECRET")

	cfg, report, err := ConfigureWithReport(&reportConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 9090 {
		t.Fatalf("Port = %d, want 9090", cfg.Port)
	}

	// Port resolved from env; the default was skipped, never tried.
	port := findField(report, "Port")
	if port == nil {
		t.Fatal("no Port field in report")
	}
	if port.Winner != SourceEnv || port.Outcome != outcomeResolved {
		t.Errorf("Port winner=%q outcome=%q, want env/resolved", port.Winner, port.Outcome)
	}
	if port.Value != 9090 {
		t.Errorf("Port value = %v, want 9090", port.Value)
	}
	var sawSkippedDefault bool
	for _, a := range port.Attempts {
		if a.Source == SourceDefault && a.Status == statusSkipped {
			sawSkippedDefault = true
		}
	}
	if !sawSkippedDefault {
		t.Errorf("expected default recorded as skipped, got %+v", port.Attempts)
	}

	// Host fell through env (not-found) to the default.
	host := findField(report, "Host")
	if host.Winner != SourceDefault || host.Outcome != outcomeDefault {
		t.Errorf("Host winner=%q outcome=%q, want default/default", host.Winner, host.Outcome)
	}

	// Provenance is the field -> winner projection.
	prov := report.Provenance()
	if prov["Port"] != SourceEnv || prov["Host"] != SourceDefault {
		t.Errorf("provenance = %+v", prov)
	}
}

func TestReportMasksSecretValue(t *testing.T) {
	os.Setenv("GOSTRUCTOR_RPT_SECRET", "hunter2")
	defer os.Unsetenv("GOSTRUCTOR_RPT_SECRET")
	os.Unsetenv("GOSTRUCTOR_RPT_HOST")
	os.Unsetenv("GOSTRUCTOR_RPT_PORT")

	cfg, report, err := ConfigureWithReport(&reportConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The real value still lands on the struct...
	if cfg.Secret != "hunter2" {
		t.Fatalf("Secret = %q, want the real value on the struct", cfg.Secret)
	}
	// ...but never appears in the report or its rendering.
	secret := findField(report, "Secret")
	if secret.Value == "hunter2" || secret.Raw == "hunter2" {
		t.Errorf("secret leaked into report: raw=%v value=%v", secret.Raw, secret.Value)
	}
	if !secret.IsSecret {
		t.Errorf("Secret field's IsSecret flag not set in report")
	}
	if s := report.String(); strings.Contains(s, "hunter2") {
		t.Errorf("secret leaked into String():\n%s", s)
	}
}

func TestWithMaskerCustomRendering(t *testing.T) {
	os.Setenv("GOSTRUCTOR_RPT_SECRET", "abcd1234")
	defer os.Unsetenv("GOSTRUCTOR_RPT_SECRET")
	os.Unsetenv("GOSTRUCTOR_RPT_HOST")
	os.Unsetenv("GOSTRUCTOR_RPT_PORT")

	_, report, err := ConfigureWithReport(&reportConfig{},
		WithMasker(func(_ FieldContext, v any) string {
			s, _ := v.(string)
			if len(s) >= 4 {
				return "****" + s[len(s)-4:]
			}
			return "****"
		}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := findField(report, "Secret").Value; got != "****1234" {
		t.Errorf("masked value = %v, want ****1234", got)
	}
	// The custom masker also drives String()'s secret rendering.
	if s := report.String(); !strings.Contains(s, "****1234") {
		t.Errorf("custom mask not used in String():\n%s", s)
	}
}

func TestConvertErrorMasksSecretValue(t *testing.T) {
	type secretInt struct {
		Token int `cfg:"token,env:GOSTRUCTOR_RPT_TOKEN" gos:"secret"`
	}
	os.Setenv("GOSTRUCTOR_RPT_TOKEN", "not-a-number")
	defer os.Unsetenv("GOSTRUCTOR_RPT_TOKEN")

	_, err := Configure(&secretInt{})
	if err == nil {
		t.Fatal("expected a convert error")
	}
	if strings.Contains(err.Error(), "not-a-number") {
		t.Errorf("secret raw value leaked into error: %v", err)
	}
}

// TestReportStringFocusedSections checks the three-part focused trace: the
// primary source line, the defaults line, and an Overrides & Secrets section
// listing only the overridden and secret fields. JSON is the majority source
// (host + region), env overrides one field (port) and supplies the secret.
func TestReportStringFocusedSections(t *testing.T) {
	type focusConfig struct {
		Host   string `cfg:"host,json:host"`
		Region string `cfg:"region,json:region"`
		Level  string `cfg:"level,json:level" gos:"default:info"`
		Port   int    `cfg:"port,env:GOSTRUCTOR_RPT_PORT,json:port"`
		Secret string `cfg:"secret,env:GOSTRUCTOR_RPT_SECRET" gos:"secret"`
	}
	path := writeReportJSON(t, `{"host":"10.0.0.1","region":"eu","port":8080}`)
	os.Setenv("GOSTRUCTOR_RPT_PORT", "9090") // overrides the JSON port
	defer os.Unsetenv("GOSTRUCTOR_RPT_PORT")
	os.Setenv("GOSTRUCTOR_RPT_SECRET", "s3cr3t")
	defer os.Unsetenv("GOSTRUCTOR_RPT_SECRET")

	_, report, err := ConfigureWithReport(&focusConfig{},
		WithSources(Env(), JSONFile(path), Default()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := report.String()

	if report.PrimarySource() != SourceJSON {
		t.Errorf("PrimarySource = %q, want json", report.PrimarySource())
	}
	for _, want := range []string{
		"[Primary Source] json (loaded 2 fields)", // host + region
		"[Defaults] applied for 1 field",          // level
		"Overrides & Secrets:",
		"Port", // overridden by env
		"(override)",
		"Secret", // secret field
		"(secret)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("String() missing %q in:\n%s", want, out)
		}
	}
	// Host resolved straight from the primary source, so it is NOT listed as
	// an override.
	if strings.Contains(out, "Host ") {
		t.Errorf("Host should not appear in the focused trace:\n%s", out)
	}
}

// TestReportStringAllDefaults covers the branch where there is no primary
// source (every field came from a default) - just the summary + defaults line.
func TestReportStringAllDefaults(t *testing.T) {
	os.Unsetenv("GOSTRUCTOR_RPT_HOST")
	os.Unsetenv("GOSTRUCTOR_RPT_PORT")
	os.Setenv("GOSTRUCTOR_RPT_SECRET", "x")
	defer os.Unsetenv("GOSTRUCTOR_RPT_SECRET")

	_, report, err := ConfigureWithReport(&reportConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := report.String()
	if report.PrimarySource() != "" {
		t.Errorf("PrimarySource = %q, want empty (no non-default source)", report.PrimarySource())
	}
	if !strings.Contains(out, "[Defaults] applied for 2 fields") {
		t.Errorf("want a defaults line for Host+Port in:\n%s", out)
	}
	if strings.Contains(out, "[Primary Source]") {
		t.Errorf("no primary source expected in:\n%s", out)
	}
}
