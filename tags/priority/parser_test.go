package priority

import (
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	reader := strings.NewReader(`stage1:cf_env,cf_default,cf_vault;stage2:cf_default,cf_vault;stage3:cf_env,cf_default`)
	ast, errParsing := NewParser(reader).Parse()
	if errParsing != nil {
		t.Error(errParsing)
	}
	t.Log(ast)
}
