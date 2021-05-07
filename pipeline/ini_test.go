package pipeline

import (
	"os"
	"testing"

	"github.com/goreflect/gostructor/tags"
	"github.com/stretchr/testify/assert"
)

func Test_tryParseIniConfig(t *testing.T) {
	os.Setenv(tags.IniFile, "../test_configs/config.ini")
	config := IniConfig{}
	config.typeSafeLoadConfigFile(nil)
	loadedFile, err := config.iniFile.GetSection("TEST")
	if err != nil {
		t.Error(err)
	}
	loadedKey, err := loadedFile.GetKey("test")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "tururu", loadedKey.String())
}
