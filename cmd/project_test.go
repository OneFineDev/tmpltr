package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProjectCommand(t *testing.T) {
	// pathToTestConfigs, _ := filepath.Abs(path.Join("..", "test", ".tmpltr"))

	tests := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name: "valid with source-set",
			args: []string{
				"--config=/home/parisb/repos/PLT.PRODUCT.TMPLTR/tmpltr/test/.tmpltr",
				"--verbose",
				"create",
				"project",
				"--output-path",
				"/home/parisb/repos/PLT.PRODUCT.TMPLTR/tmpltr/test/output",
				"--project-name=test",
				"--source-set",
				"terraformChildSet",
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange

			// tt.args = append(tt.args)
			out := new(bytes.Buffer)
			cmd := NewRootCommand()
			cmd.SetOut(out)
			cmd.SetArgs(tt.args)

			// Act
			err := cmd.Execute()

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				outString := out.String()
				assert.Equal(t, "Project Called", outString)
			}

			t.Cleanup(func() {
				os.RemoveAll("/home/parisb/repos/PLT.PRODUCT.TMPLTR/tmpltr/test/output")
			})
		})
	}
}
