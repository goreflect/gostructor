package gostructor

import (
	"reflect"
	"testing"
)

func TestScreamingSnake(t *testing.T) {
	cases := []struct{ in, want string }{
		{"port", "PORT"},
		{"maxConns", "MAX_CONNS"},
		{"HTTPServer", "HTTP_SERVER"},
		{"already_snake", "ALREADY_SNAKE"},
		{"kebab-case", "KEBAB_CASE"},
		{"dotted.name", "DOTTED_NAME"},
		{"with space", "WITH_SPACE"},
		{"v2Endpoint", "V2_ENDPOINT"},
		{"ID", "ID"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := ScreamingSnake(tc.in); got != tc.want {
			t.Errorf("ScreamingSnake(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestIdentity(t *testing.T) {
	if got := Identity("server.host"); got != "server.host" {
		t.Errorf("Identity() = %q, want unchanged", got)
	}
}

// fieldContextFor builds a FieldContext straight from a struct tag, exercising
// the lazy-parse path used when the engine hasn't attached a cached plan.
func fieldContextFor(tag string, typ reflect.Type) FieldContext {
	return FieldContext{StructField: reflect.StructField{
		Name: "Field",
		Tag:  reflect.StructTag(tag),
		Type: typ,
	}}
}

func TestFieldContextAccessors(t *testing.T) {
	f := fieldContextFor(`cfg:"port,env:SERVER_PORT" gos:"default:8080,sep:|,secret"`, reflect.TypeOf(0))

	if f.Base() != "port" {
		t.Errorf("Base() = %q, want port", f.Base())
	}
	if v, ok := f.Override("env"); !ok || v != "SERVER_PORT" {
		t.Errorf("Override(env) = %q,%v", v, ok)
	}
	if _, ok := f.Override("json"); ok {
		t.Errorf("Override(json) should be absent")
	}
	if got := f.SourceKey("env", ScreamingSnake); got != "SERVER_PORT" {
		t.Errorf("SourceKey(env) = %q, want the override", got)
	}
	if got := f.SourceKey("json", Identity); got != "port" {
		t.Errorf("SourceKey(json) = %q, want naming(base)", got)
	}
	if v, ok := f.Meta("default"); !ok || v != "8080" {
		t.Errorf("Meta(default) = %q,%v", v, ok)
	}
	if !f.IsSecret() {
		t.Errorf("IsSecret() = false, want true")
	}
	if f.Optional() {
		t.Errorf("Optional() = true, want false")
	}
	if f.Separator() != "|" {
		t.Errorf("Separator() = %q, want |", f.Separator())
	}
}

func TestFieldContextSourceKeyNoBase(t *testing.T) {
	// No cfg tag at all: a naming source has no key, so it does not apply.
	f := fieldContextFor(`gos:"default:x"`, reflect.TypeOf(""))
	if got := f.SourceKey("env", ScreamingSnake); got != "" {
		t.Errorf("SourceKey with no base = %q, want empty", got)
	}
	if got := f.SourceKey("json", nil); got != "" {
		t.Errorf("SourceKey with nil naming and no base = %q, want empty", got)
	}
}

func TestFieldContextDefaultSeparator(t *testing.T) {
	f := fieldContextFor(`cfg:"tags"`, reflect.TypeOf([]string{}))
	if f.Separator() != "," {
		t.Errorf("Separator() = %q, want comma default", f.Separator())
	}
}

func TestFieldContextConfigured(t *testing.T) {
	if fieldContextFor(``, reflect.TypeOf("")).configured() {
		t.Errorf("untagged field reported configured")
	}
	if !fieldContextFor(`cfg:"x"`, reflect.TypeOf("")).configured() {
		t.Errorf("cfg-tagged field reported unconfigured")
	}
	if !fieldContextFor(`gos:"optional"`, reflect.TypeOf("")).configured() {
		t.Errorf("gos-only field reported unconfigured")
	}
}
