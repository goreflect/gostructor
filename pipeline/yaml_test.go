package pipeline

import (
	"reflect"
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

// TestYamlGetComplexTypeNestedList guards against a regression where reading
// a slice-typed field out of a nested yaml section panicked instead of
// converting, because the values are not always strings after parsing.
func TestYamlGetComplexTypeNestedList(t *testing.T) {
	strct := struct {
		Test6 []int `cf_yaml:"test5.test6"`
	}{}
	fieldType := reflect.ValueOf(strct).Type().Field(0)
	fieldValue := reflect.ValueOf(strct).Field(0)

	config := YamlConfig{
		fileName: "../test_configs/config.yml",
	}
	got := config.GetComplexType(&structContext{
		Value:       fieldValue,
		StructField: fieldType,
	})
	if got.GetNotAValue() != nil {
		t.Fatalf("unexpected error: %v", got.GetNotAValue().Error)
	}
	assert.Equal(t, []int{1231, 15123}, got.Value.Interface())
}

// TestYamlGetComplexTypeTopLevelList covers a plain top-level list field.
func TestYamlGetComplexTypeTopLevelList(t *testing.T) {
	strct := struct {
		Test2 []string `cf_yaml:"test2"`
	}{}
	fieldType := reflect.ValueOf(strct).Type().Field(0)
	fieldValue := reflect.ValueOf(strct).Field(0)

	config := YamlConfig{
		fileName: "../test_configs/config.yml",
	}
	got := config.GetComplexType(&structContext{
		Value:       fieldValue,
		StructField: fieldType,
	})
	if got.GetNotAValue() != nil {
		t.Fatalf("unexpected error: %v", got.GetNotAValue().Error)
	}
	assert.Equal(t, []string{"string", "string2", "string3"}, got.Value.Interface())
}
