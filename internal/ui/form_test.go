//go:build !integration

package ui

import (
	"reflect"
	"testing"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		src      map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "Empty source map",
			prefix:   "",
			src:      map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:   "Flat source map",
			prefix: "",
			src: map[string]interface{}{
				"key1": new(string),
				"key2": new(string),
			},
			expected: map[string]interface{}{
				"key1": new(string),
				"key2": new(string),
			},
		},
		{
			name:   "Nested source map",
			prefix: "",
			src: map[string]interface{}{
				"key1": map[string]interface{}{
					"key1_1": new(string),
					"key1_2": new(string),
				},
				"key2": new(string),
				"key3": map[string]interface{}{
					"key3_1": map[string]interface{}{
						"key3_1_1": new(string),
					},
				},
			},
			expected: map[string]interface{}{
				"key1.key1_1":          new(string),
				"key1.key1_2":          new(string),
				"key2":                 new(string),
				"key3.key3_1.key3_1_1": new(string),
			},
		},
		{
			name:   "Source map with prefix",
			prefix: "prefix",
			src: map[string]interface{}{
				"key1": new(string),
			},
			expected: map[string]interface{}{
				"prefix.key1": new(string),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			dest := make(map[string]*string)

			// Act
			Flatten(tt.prefix, tt.src, dest)

			// Assert
			if len(dest) != len(tt.expected) {
				t.Errorf("Flatten() produced map of length %d, expected %d", len(dest), len(tt.expected))
				return
			}

			for key, expectedValue := range tt.expected {
				actualValue, exists := dest[key]
				if !exists {
					t.Errorf("Flatten() missing key %q in result", key)
					continue
				}
				if reflect.TypeOf(actualValue) != reflect.TypeOf(expectedValue) ||
					reflect.ValueOf(actualValue).Kind() != reflect.Ptr {
					t.Errorf("Flatten() key %q has value of incorrect type or not a pointer", key)
				}
			}
		})
	}
}
func TestRebuild(t *testing.T) {
	tests := []struct {
		name              string
		formMap           map[string]*string
		templateValuesMap map[string]any
		expectedResult    map[string]any
	}{
		{
			name: "Single level keys",
			formMap: map[string]*string{
				"key1": ptr("value1"),
				"key2": ptr("value2"),
			},
			templateValuesMap: map[string]any{},
			expectedResult: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "Two level keys",
			formMap: map[string]*string{
				"key1.key2": ptr("value1"),
				"key3.key4": ptr("value2"),
			},
			templateValuesMap: map[string]any{},
			expectedResult: map[string]any{
				"key1": map[string]any{
					"key2": "value1",
				},
				"key3": map[string]any{
					"key4": "value2",
				},
			},
		},
		{
			name: "Three level keys",
			formMap: map[string]*string{
				"key1.key2.key3": ptr("value1"),
			},
			templateValuesMap: map[string]any{},
			expectedResult: map[string]any{
				"key1": map[string]any{
					"key2": map[string]any{
						"key3": "value1",
					},
				},
			},
		},
		{
			name: "Mixed levels",
			formMap: map[string]*string{
				"key1":           ptr("value1"),
				"key2.key3":      ptr("value2"),
				"key4.key5.key6": ptr("value3"),
			},
			templateValuesMap: map[string]any{},
			expectedResult: map[string]any{
				"key1": "value1",
				"key2": map[string]any{
					"key3": "value2",
				},
				"key4": map[string]any{
					"key5": map[string]any{
						"key6": "value3",
					},
				},
			},
		},
		{
			name: "Existing templateValuesMap",
			formMap: map[string]*string{
				"key1.key2": ptr("value1"),
			},
			templateValuesMap: map[string]any{
				"key1": map[string]any{
					"existingKey": "existingValue",
				},
			},
			expectedResult: map[string]any{
				"key1": map[string]any{
					"existingKey": "existingValue",
					"key2":        "value1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			formMap := tt.formMap
			templateValuesMap := tt.templateValuesMap

			// Act
			result := Rebuild(formMap, templateValuesMap)

			// Assert
			if !reflect.DeepEqual(result, tt.expectedResult) {
				t.Errorf("Rebuild() = %v, expected %v", result, tt.expectedResult)
			}
		})
	}
}

func ptr(s string) *string {
	return &s
}
