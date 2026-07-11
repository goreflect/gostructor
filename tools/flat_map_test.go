package tools

import (
	"testing"

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

	assert.Equal(t, "test", result["test2.test5.1"])
}

// TestFlatMapDoesNotDropNestedKeys guards against a regression where a nested
// map processed before any scalar top-level key was silently dropped instead
// of merged, because Go map iteration order is randomized. Run repeatedly to
// exercise both possible iteration orders deterministically.
func TestFlatMapDoesNotDropNestedKeys(t *testing.T) {
	for i := 0; i < 50; i++ {
		result := FlatMap(map[string]interface{}{
			"nested": map[string]interface{}{
				"a": "1",
				"b": "2",
			},
			"top": "value",
		})
		assert.Equal(t, "1", result["nested.a"])
		assert.Equal(t, "2", result["nested.b"])
		assert.Equal(t, "value", result["top"])
		assert.Equal(t, 3, len(result))
	}
}

// TestFlatMapOnlyNestedKey covers the deterministic case that always failed
// before the fix: a source map whose only key is itself a nested map.
func TestFlatMapOnlyNestedKey(t *testing.T) {
	result := FlatMap(map[string]interface{}{
		"nested": map[string]interface{}{
			"a": "1",
		},
	})
	assert.Equal(t, "1", result["nested.a"])
	assert.Equal(t, 1, len(result))
}
