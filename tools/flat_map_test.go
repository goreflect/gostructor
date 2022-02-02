package tools

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFlatMap(t *testing.T) {
	result := FlatMap(map[string]interface{}{
		"test": []string{"test1", "test2"},
		"test2": map[string]interface{}{
			"test4": []int{1, 2, 3},
			"test5": map[string]interface{}{
				"1": "test",
			},
		},
	})

	logrus.Info(result)
	assert.Equal(t, "test", result["test2.test5.1"])
}
