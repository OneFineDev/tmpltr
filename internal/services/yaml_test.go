package services //nolint:testpackage // unexported

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockYamlStruct struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}

func (m MockYamlStruct) Yamafiable() {}

func TestReadYamlFromFile(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		fileContent   string
		expectError   bool
		expectedName  string
		expectedAge   int
		errorContains string
	}{
		{
			name:         "Success",
			filePath:     "test.yaml",
			fileContent:  "name: John Doe\nage: 30\n",
			expectError:  false,
			expectedName: "John Doe",
			expectedAge:  30,
		},
		{
			name:          "FileNotFound",
			filePath:      "nonexistent.yaml",
			expectError:   true,
			errorContains: "failed to read file",
		},
		{
			name:          "InvalidYaml",
			filePath:      "invalid.yaml",
			fileContent:   "name: John Doe\nage: thirty\n",
			expectError:   true,
			errorContains: "failed to unmarshal YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			if tt.fileContent != "" {
				err := os.WriteFile(tt.filePath, []byte(tt.fileContent), 0644)
				require.NoError(t, err)
				defer os.Remove(tt.filePath)
			}

			// Act
			result, err := readYamlFromFile[MockYamlStruct](tt.filePath)

			// Assert
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedName, result.Name)
				assert.Equal(t, tt.expectedAge, result.Age)
			}
		})
	}
}
