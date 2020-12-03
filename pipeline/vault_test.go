package pipeline

import (
	"os"
	"testing"

	"github.com/goreflect/gostructor/infra"
)

type TestStruct struct {
	MyAPIKey            string  `cf_vault:"test/tururu#api-key"`
	MyIntKey            int64   `cf_vault:"test/tururu#kur"`
	MyCustomComplexType []int32 `cf_vault:"test/tururu#kur2"`
}

func TestConnectionToVault(t *testing.T) {
	os.Setenv("VAULT_URL", "http://localhost:1234")
	os.Setenv("VAULT_TOKEN", "myroot")
	result, err := Configure(&TestStruct{}, "", []infra.FuncType{infra.FunctionSetupVault}, "", true)
	if err != nil {
		t.Fail()
		t.Log(err)
		return
	}
	t.Log(result)
}
