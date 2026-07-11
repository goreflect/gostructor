package toml

import (
	"reflect"
	"testing"
)

func TestParseFixture(t *testing.T) {
	data := []byte(`[postgres]
user = "pelletier"
password = "mypassword"
test1 = 1231
test2 = 43.52
test3 = ["myTest1", "myTest2"]
test4 = 123
`)
	doc, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	postgres, ok := doc["postgres"].(map[string]any)
	if !ok {
		t.Fatalf("postgres not parsed as a table: %v", doc["postgres"])
	}
	if postgres["user"] != "pelletier" {
		t.Errorf("user = %v", postgres["user"])
	}
	if postgres["password"] != "mypassword" {
		t.Errorf("password = %v", postgres["password"])
	}
	if postgres["test1"] != int64(1231) {
		t.Errorf("test1 = %v (%T)", postgres["test1"], postgres["test1"])
	}
	if postgres["test2"] != 43.52 {
		t.Errorf("test2 = %v", postgres["test2"])
	}
	if !reflect.DeepEqual(postgres["test3"], []any{"myTest1", "myTest2"}) {
		t.Errorf("test3 = %v", postgres["test3"])
	}
	if postgres["test4"] != int64(123) {
		t.Errorf("test4 = %v", postgres["test4"])
	}
}

func TestParseTopLevelKeys(t *testing.T) {
	doc, err := Parse([]byte("name = \"gostructor\"\nversion = 1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc["name"] != "gostructor" || doc["version"] != int64(1) {
		t.Errorf("got %v", doc)
	}
}

func TestParseNestedTable(t *testing.T) {
	doc, err := Parse([]byte("[a.b.c]\nkey = true"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a := doc["a"].(map[string]any)
	b := a["b"].(map[string]any)
	c := b["c"].(map[string]any)
	if c["key"] != true {
		t.Errorf("a.b.c.key = %v", c["key"])
	}
}

func TestParseComment(t *testing.T) {
	doc, err := Parse([]byte("# full line comment\nkey = \"value\" # trailing comment\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc["key"] != "value" {
		t.Errorf("key = %v", doc["key"])
	}
}

func TestParseCommentHashInsideString(t *testing.T) {
	doc, err := Parse([]byte(`key = "not a # comment"`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc["key"] != "not a # comment" {
		t.Errorf("key = %v", doc["key"])
	}
}

func TestParseSingleQuotedString(t *testing.T) {
	doc, err := Parse([]byte(`key = 'literal value'`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc["key"] != "literal value" {
		t.Errorf("key = %v", doc["key"])
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name string
		data string
	}{
		{"array of tables unsupported", "[[section]]\nkey = 1"},
		{"unterminated table header", "[section\nkey = 1"},
		{"empty table header", "[]"},
		{"missing equals", "key value"},
		{"unterminated string", `key = "unterminated`},
		{"unterminated array", "key = [1, 2"},
		{"unrecognized scalar", "key = notaliteral"},
		{"key redefined as table", "key = 1\n[key]\nsub = 2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse([]byte(tc.data)); err == nil {
				t.Errorf("expected an error for input %q", tc.data)
			}
		})
	}
}
