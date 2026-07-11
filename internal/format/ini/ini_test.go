package ini

import "testing"

func TestParseBasic(t *testing.T) {
	data := []byte(`
; a comment
[TEST]
test = tururu
test2 = 614
test3 = 86.27
test4 = str1, str2, str3
test5 = 15
`)
	file, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cases := map[string]string{
		"test":  "tururu",
		"test2": "614",
		"test3": "86.27",
		"test4": "str1, str2, str3",
		"test5": "15",
	}
	for key, want := range cases {
		got, ok := file.Get("TEST", key)
		if !ok {
			t.Errorf("key %q: not found", key)
			continue
		}
		if got != want {
			t.Errorf("key %q: got %q, want %q", key, got, want)
		}
	}
}

func TestParseGlobalSection(t *testing.T) {
	data := []byte("standalone = value\n[Section]\nkey = other")
	file, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := file.Get(GlobalSection, "standalone")
	if !ok || got != "value" {
		t.Errorf("got %q, %v, want %q, true", got, ok, "value")
	}
}

func TestParseHashComment(t *testing.T) {
	data := []byte("# full line comment\n[S]\nkey = value")
	file, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := file.Get("S", "key")
	if !ok || got != "value" {
		t.Errorf("got %q, %v", got, ok)
	}
}

func TestParseColonSeparator(t *testing.T) {
	data := []byte("[S]\nkey: value")
	file, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := file.Get("S", "key")
	if !ok || got != "value" {
		t.Errorf("got %q, %v", got, ok)
	}
}

func TestParseQuotedValue(t *testing.T) {
	data := []byte(`[S]
double = "quoted value"
single = 'other value'`)
	file, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, _ := file.Get("S", "double"); got != "quoted value" {
		t.Errorf("double: got %q", got)
	}
	if got, _ := file.Get("S", "single"); got != "other value" {
		t.Errorf("single: got %q", got)
	}
}

func TestParseEmptySection(t *testing.T) {
	data := []byte("[empty]")
	file, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := file.Get("empty", "anything"); ok {
		t.Error("expected no keys in an empty section")
	}
}

func TestParseUnknownKeyOrSection(t *testing.T) {
	file, err := Parse([]byte("[S]\nkey = value"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := file.Get("S", "missing"); ok {
		t.Error("expected missing key to report ok=false")
	}
	if _, ok := file.Get("missing-section", "key"); ok {
		t.Error("expected missing section to report ok=false")
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name string
		data string
	}{
		{"unterminated section", "[section"},
		{"empty section name", "[]"},
		{"key with no separator", "[S]\njustaword"},
		{"empty key", "[S]\n = value"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse([]byte(tc.data)); err == nil {
				t.Errorf("expected an error for input %q", tc.data)
			}
		})
	}
}

func TestParseWhitespaceAndBlankLines(t *testing.T) {
	data := []byte("\n\n[S]\n\n   key   =   value with spaces   \n\n")
	file, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := file.Get("S", "key")
	if !ok || got != "value with spaces" {
		t.Errorf("got %q, %v", got, ok)
	}
}
