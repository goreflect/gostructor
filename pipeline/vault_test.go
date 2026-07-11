package pipeline

import "testing"

// TestParseVaultTagMalformed guards against a regression where a cf_vault
// tag missing the "#key" suffix caused an index-out-of-range panic instead
// of a returned error.
func TestParseVaultTagMalformed(t *testing.T) {
	cases := []string{"", "no-hash-here", "path#", "#key", "a#b#c"}
	for _, tagValue := range cases {
		t.Run(tagValue, func(t *testing.T) {
			_, _, err := parseVaultTag(tagValue)
			if tagValue == "a#b#c" {
				if err != nil {
					t.Errorf("expected 'a#b#c' to split into path='a', key='b#c', got error: %v", err)
				}
				return
			}
			if err == nil {
				t.Errorf("expected error for malformed tag %q", tagValue)
			}
		})
	}
}

func TestParseVaultTagWellFormed(t *testing.T) {
	path, key, err := parseVaultTag("secret/service/stage#my-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "secret/service/stage" || key != "my-key" {
		t.Errorf("got path=%q key=%q", path, key)
	}
}

type TestStruct struct {
	MyAPIKey            string  `cf_vault:"test/tururu#api-key"`
	MyIntKey            int64   `cf_vault:"test/tururu#kur"`
	MyCustomComplexType []int32 `cf_vault:"test/tururu#kur2"`
}

// func TestConnectionToVault(t *testing.T) {
// 	os.Setenv("VAULT_URL", "http://localhost:1234")
// 	os.Setenv("VAULT_TOKEN", "myroot")
// 	result, err := Configure(&TestStruct{}, "", []infra.FuncType{infra.FunctionSetupVault}, "", true)
// 	if err != nil {
// 		t.Fail()
// 		t.Log(err)
// 		return
// 	}
// 	t.Log(result)
// }
