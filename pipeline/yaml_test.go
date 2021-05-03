package pipeline

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestYamlParse(t *testing.T) {
	config := YamlConfig{
		fileName: "../test_configs/config.yml",
	}
	loadedConfig, err := config.typeSafeLoadConfigFile(&structContext{})
	logrus.Info(err)
	assert.Equal(t, true, loadedConfig)
}

func TestYamlParseByKey(t *testing.T) {
	config := YamlConfig{
		fileName: "../test_configs/config.yml",
	}
	config.typeSafeLoadConfigFile(&structContext{})
	assert.Equal(t, "str1", config.parsedData["test5.test4"])
}
