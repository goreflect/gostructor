package gostructor

import (
	"os"
	"strings"
	"testing"
)

type reportConfig struct {
	Host   string `cf_env:"GOSTRUCTOR_RPT_HOST" cf_default:"0.0.0.0"`
	Port   int    `cf_env:"GOSTRUCTOR_RPT_PORT" cf_default:"8080"`
	Secret string `cf_env:"GOSTRUCTOR_RPT_SECRET" cf_secret:""`
}

func findField(r *Report, name string) *FieldResolution {
	for i := range r.Fields {
		if r.Fields[i].Field == name {
			return &r.Fields[i]
		}
	}
	return nil
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

	// Port resolved from cf_env; cf_default was skipped, never tried.
	port := findField(report, "Port")
	if port == nil {
		t.Fatal("no Port field in report")
	}
	if port.Winner != EnvTag || port.Outcome != outcomeResolved {
		t.Errorf("Port winner=%q outcome=%q, want cf_env/resolved", port.Winner, port.Outcome)
	}
	if port.Value != 9090 {
		t.Errorf("Port value = %v, want 9090", port.Value)
	}
	var sawSkippedDefault bool
	for _, a := range port.Attempts {
		if a.Tag == DefaultTag && a.Status == statusSkipped {
			sawSkippedDefault = true
		}
	}
	if !sawSkippedDefault {
		t.Errorf("expected cf_default recorded as skipped, got %+v", port.Attempts)
	}

	// Host fell through env (not-found) to cf_default.
	host := findField(report, "Host")
	if host.Winner != DefaultTag || host.Outcome != outcomeDefault {
		t.Errorf("Host winner=%q outcome=%q, want cf_default/default", host.Winner, host.Outcome)
	}

	// Provenance is the field -> winner projection.
	prov := report.Provenance()
	if prov["Port"] != EnvTag || prov["Host"] != DefaultTag {
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
}

func TestConvertErrorMasksSecretValue(t *testing.T) {
	type secretInt struct {
		Token int `cf_env:"GOSTRUCTOR_RPT_TOKEN" cf_secret:""`
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

func TestReportStringRendersTree(t *testing.T) {
	os.Setenv("GOSTRUCTOR_RPT_PORT", "9090")
	defer os.Unsetenv("GOSTRUCTOR_RPT_PORT")
	os.Unsetenv("GOSTRUCTOR_RPT_HOST")
	os.Setenv("GOSTRUCTOR_RPT_SECRET", "s3cr3t")
	defer os.Unsetenv("GOSTRUCTOR_RPT_SECRET")

	_, report, err := ConfigureWithReport(&reportConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := report.String()
	if !strings.Contains(out, "Port int  ⇐ cf_env(GOSTRUCTOR_RPT_PORT)=9090") {
		t.Errorf("unexpected Port line in:\n%s", out)
	}
	if !strings.Contains(out, "cf_default: skipped") {
		t.Errorf("expected skipped-default annotation in:\n%s", out)
	}
}
