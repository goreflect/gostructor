package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDataFromToml(t *testing.T) {
	test := TomlConfig{fileName: "../test_configs/config.toml"}
	test.typeSafeLoadConfigFile(nil)
	assert.Equal(t, "mypassword", test.parsedData.Get("postgres.password"))
}
