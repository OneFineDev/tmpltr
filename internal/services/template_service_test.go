package services_test

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/OneFineDev/tmpltr/internal/services"
	"github.com/OneFineDev/tmpltr/internal/types"
)

func TestExtractTemplateKeys(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		wantKeys   types.TemplateValuesMap
		wantErrors int
	}{
		{
			name:       "Valid template with single key",
			template:   `{{.Key1}}`,
			wantKeys:   types.TemplateValuesMap{"Key1": ""},
			wantErrors: 0,
		},
		{
			name:       "Valid template with multiple keys",
			template:   `{{.Key1}} and {{.Key2}}`,
			wantKeys:   types.TemplateValuesMap{"Key1": "", "Key2": ""},
			wantErrors: 0,
		},
		{
			name:     "Valid template with 1 level of nested keys",
			template: `{{.Key1.Nested1}} and {{.Key1.Nested2}} and {{.Key2.Nested1}} and {{.Key2.Nested2}}`,
			wantKeys: types.TemplateValuesMap{
				"Key1": map[string]string{
					"Nested1": "",
					"Nested2": "",
				},
				"Key2": map[string]string{
					"Nested1": "",
					"Nested2": "",
				},
			},
			wantErrors: 0,
		},
		{
			name:     "Valid template with 2 levels of nested keys",
			template: `{{.Key1.Nested1.DoubleNested1}} and {{.Key1.Nested1.DoubleNested2}} and {{.Key2.Nested1.DoubleNested1}} and {{.Key2.Nested1.DoubleNested2}}`,
			wantKeys: types.TemplateValuesMap{
				"Key1": map[string]any{
					"Nested1": map[string]string{
						"DoubleNested1": "",
						"DoubleNested2": "",
					},
					"Nested2": map[string]string{
						"DoubleNested1": "",
						"DoubleNested2": "",
					},
				},
				"Key2": map[string]any{
					"Nested1": map[string]string{
						"DoubleNested1": "",
						"DoubleNested2": "",
					},
					"Nested2": map[string]string{
						"DoubleNested1": "",
						"DoubleNested2": "",
					},
				},
			},
			wantErrors: 0,
		},
		{
			name:       "Unsupported node type",
			template:   `{{.Key1}} {{if .Condition}}Conditional{{end}}`,
			wantKeys:   types.TemplateValuesMap{"Key1": ""},
			wantErrors: 1,
		},
		{
			name:       "Text-only template",
			template:   `This is a plain text template.`,
			wantKeys:   types.TemplateValuesMap{},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New(tt.name).Parse(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			valuesMap := make(types.TemplateValuesMap)
			ts := &services.TemplateService{}
			errors := ts.ExtractTemplateKeys(tmpl, valuesMap)

			if len(errors) != tt.wantErrors {
				t.Errorf("Expected %d errors\n, got %d", tt.wantErrors, len(errors))
			}

			if len(valuesMap) != len(tt.wantKeys) {
				t.Errorf("Expected keys:\n %v\n, got:\n %v", tt.wantKeys, valuesMap)
			}

			for key := range tt.wantKeys {
				if _, exists := valuesMap[key]; !exists {
					t.Errorf("Expected key %s not found in valuesMap", key)
				}
			}
		})
	}
}

func TestGetTemplateFiles(t *testing.T) { //nolint:gocognit
	tests := []struct {
		name          string
		setupFiles    map[string]string // map of file path to content
		expectedFiles []string
		expectError   bool
	}{
		{
			name: "Directory with .template files",
			setupFiles: map[string]string{
				"testdata/file1.template":        "",
				"testdata/file2.template":        "",
				"testdata/subdir/file3.template": "",
				"testdata/subdir/file4.txt":      "",
			},
			expectedFiles: []string{
				"testdata/file1.template",
				"testdata/file2.template",
				"testdata/subdir/file3.template",
			},
			expectError: false,
		},
		{
			name: "Directory without .template files",
			setupFiles: map[string]string{
				"testdata/file1.txt": "",
				"testdata/file2.md":  "",
			},
			expectedFiles: []string{},
			expectError:   false,
		},
		{
			name: "Directory with .git folder",
			setupFiles: map[string]string{
				"testdata/.git/config":    "",
				"testdata/file1.template": "",
			},
			expectedFiles: []string{
				"testdata/file1.template",
			},
			expectError: false,
		},
		{
			name:          "Non-existent directory",
			setupFiles:    nil,
			expectedFiles: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			testDir := "testdata"
			if tt.setupFiles != nil {
				for path, content := range tt.setupFiles {
					fullPath := filepath.Join(testDir, path)
					if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
						t.Fatalf("Failed to create test directory: %v", err)
					}
					if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
						t.Fatalf("Failed to create test file: %v", err)
					}
				}
			}

			// Run the function
			ts := &services.TemplateService{}
			err := ts.GetTemplateFiles(testDir)

			// Check for errors
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
			}

			// Check the returned files
			if !tt.expectError {
				expectedFiles := make([]string, len(tt.expectedFiles))
				for i, file := range tt.expectedFiles {
					expectedFiles[i] = filepath.Join(testDir, file)
				}
				if len(ts.TemplateFiles) != len(expectedFiles) {
					t.Errorf("Expected %d files, got %d", len(expectedFiles), len(ts.TemplateFiles))
				}
				for _, file := range expectedFiles {
					found := false
					for _, f := range ts.TemplateFiles {
						if f == file {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected file %s not found in result", file)
					}
				}
			}

			// Cleanup test environment
			os.RemoveAll(testDir)
		})
	}
}

func TestValidateTemplateValues(t *testing.T) {
	tests := []struct {
		name         string
		valuesMap    types.TemplateValuesMap
		unknown      map[string]any
		expectedKeys []string
	}{
		{
			name: "All keys match",
			valuesMap: types.TemplateValuesMap{
				"Key1": "",
				"Key2": map[string]any{
					"Nested1": "",
					"Nested2": map[string]any{
						"DoubleNested1": "",
					},
				},
			},
			unknown: map[string]any{
				"Key1": "",
				"Key2": map[string]any{
					"Nested1": "",
					"Nested2": map[string]any{
						"DoubleNested1": "",
					},
				},
			},
			expectedKeys: []string{},
		},
		{
			name: "Missing top-level key",
			valuesMap: types.TemplateValuesMap{
				"Key1": "",
				"Key2": "",
			},
			unknown: map[string]any{
				"Key1": "",
			},
			expectedKeys: []string{"Key2"},
		},
		{
			name: "Missing nested key",
			valuesMap: types.TemplateValuesMap{
				"Key1": map[string]any{
					"Nested1": "",
					"Nested2": "",
				},
			},
			unknown: map[string]any{
				"Key1": map[string]any{
					"Nested1": "",
				},
			},
			expectedKeys: []string{"Key1.Nested2"},
		},
		{
			name: "Missing double-nested key",
			valuesMap: types.TemplateValuesMap{
				"Key1": map[string]any{
					"Nested1": map[string]any{
						"DoubleNested1": "",
						"DoubleNested2": "",
					},
				},
			},
			unknown: map[string]any{
				"Key1": map[string]any{
					"Nested1": map[string]any{
						"DoubleNested1": "",
					},
				},
			},
			expectedKeys: []string{"Key1.Nested1.DoubleNested2"},
		},
		{
			name: "Completely missing key",
			valuesMap: types.TemplateValuesMap{
				"Key1": "",
			},
			unknown:      map[string]any{},
			expectedKeys: []string{"Key1"},
		},
		{
			name:         "Empty valuesMap",
			valuesMap:    types.TemplateValuesMap{},
			unknown:      map[string]any{"Key1": ""},
			expectedKeys: []string{},
		},
		{
			name: "Empty unknown map",
			valuesMap: types.TemplateValuesMap{
				"Key1": "",
			},
			unknown:      map[string]any{},
			expectedKeys: []string{"Key1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ts := &services.TemplateService{}

			// Act
			missingKeys := ts.ValidateTemplateValues(tt.valuesMap, tt.unknown)

			// Assert
			if len(missingKeys) != len(tt.expectedKeys) {
				t.Errorf("Expected %d missing keys, got %d", len(tt.expectedKeys), len(missingKeys))
			}
			for _, expectedKey := range tt.expectedKeys {
				found := false
				for _, missingKey := range missingKeys {
					if missingKey == expectedKey {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected missing key %s not found", expectedKey)
				}
			}
		})
	}
}
func TestRenameTargetTemplateFiles(t *testing.T) { //nolint:gocognit
	tests := []struct {
		name                 string
		setupFiles           []string
		targetPaths          []string
		expectedRenamedFiles []string
		expectError          bool
	}{
		{
			name: "Rename .template files successfully",
			setupFiles: []string{
				"file1.template",
				"file2.template",
			},
			targetPaths: []string{
				"testdata/file1.template",
				"testdata/file2.template",
			},
			expectedRenamedFiles: []string{
				"testdata/file1",
				"testdata/file2",
			},
			expectError: false,
		},
		{
			name: "No .template files to rename",
			setupFiles: []string{
				"file1.txt",
				"file2.md",
			},
			targetPaths: []string{
				"testdata/file1.txt",
				"testdata/file2.md",
			},
			expectedRenamedFiles: []string{
				"testdata/file1.txt",
				"testdata/file2.md",
			},
			expectError: false,
		},
		{
			name: "Error renaming file",
			setupFiles: []string{
				"file1.template",
			},
			targetPaths: []string{
				"testdata/nonexistent.template",
			},
			expectedRenamedFiles: nil,
			expectError:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			testDir := "testdata"

			if tt.setupFiles != nil {
				for _, file := range tt.setupFiles {
					fullPath := filepath.Join(testDir, file)
					if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
						t.Fatalf("Failed to create test directory: %v", err)
					}
					if err := os.WriteFile(fullPath, []byte(""), 0644); err != nil {
						t.Fatalf("Failed to create test file: %v", err)
					}
				}
			}

			ts := &services.TemplateService{}

			// Act
			err := ts.RenameTargetTemplateFiles(tt.targetPaths)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
				for _, renamedFile := range tt.expectedRenamedFiles {
					if _, err := os.Stat(renamedFile); os.IsNotExist(err) {
						t.Errorf("Expected renamed file %s not found", renamedFile)
					}
				}
			}

			// Cleanup
			os.RemoveAll(testDir)
		})
	}
}

func TestExecuteTemplates(t *testing.T) { //nolint:gocognit
	tests := []struct {
		name              string
		targetToTemplate  types.TargetFileToTemplateMap
		templateValuesMap types.TemplateValuesMap
		expectedContent   map[string]string
	}{
		{
			name: "Execute templates with nested values",
			targetToTemplate: types.TargetFileToTemplateMap{
				"test1.txt": template.Must(
					template.New("test1").
						Parse("Simple value: {{.key1}}\nNested value: {{.key2.key3}}\nDeep nested value: {{.key4.key5.key6}}"),
				),
				"test2.txt": template.Must(template.New("test2").Parse("Key1: {{.key1}}")),
			},
			templateValuesMap: types.TemplateValuesMap{
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
			expectedContent: map[string]string{
				"test1.txt": "Simple value: value1\nNested value: value2\nDeep nested value: value3",
				"test2.txt": "Key1: value1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tempDir, err := os.MkdirTemp("", "template-test")
			if err != nil {
				t.Fatalf("Failed to create temporary directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Convert paths in targetToTemplate to be in tempDir
			targetToTemplate := types.TargetFileToTemplateMap{}
			for target, tmpl := range tt.targetToTemplate {
				fullPath := filepath.Join(tempDir, target)
				targetToTemplate[fullPath] = tmpl
			}

			// Act
			err = services.ExecuteTemplates(targetToTemplate, tt.templateValuesMap)

			// Assert
			if err != nil {
				t.Fatalf("ExecuteTemplates() error = %v", err)
			}

			// Verify each file has expected content
			for target, expectedContent := range tt.expectedContent {
				fullPath := filepath.Join(tempDir, target)
				content, err := os.ReadFile(fullPath)
				if err != nil {
					t.Errorf("Failed to read file %s: %v", fullPath, err)
					continue
				}

				if string(content) != expectedContent {
					t.Errorf("File %s content = %q, want %q", target, string(content), expectedContent)
				}
			}
		})
	}
}
func TestParseTemplates(t *testing.T) { //nolint:gocognit
	tests := []struct {
		name          string
		setupFiles    map[string]string // map of file path to content
		expectedFiles []string
		expectError   bool
	}{
		{
			name: "Valid template files",
			setupFiles: map[string]string{
				"testdata/file1.template": "{{.Key1}}",
				"testdata/file2.template": "{{.Key2}}",
			},
			expectedFiles: []string{
				"testdata/file1.template",
				"testdata/file2.template",
			},
			expectError: false,
		},
		{
			name: "Invalid template file",
			setupFiles: map[string]string{
				"testdata/file1.template": "{{.Key1}",
			},
			expectedFiles: nil,
			expectError:   true,
		},
		{
			name:          "No template files",
			setupFiles:    nil,
			expectedFiles: nil,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			testDir := "testdata"
			if tt.setupFiles != nil {
				for path, content := range tt.setupFiles {
					fullPath := filepath.Join(testDir, path)
					if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
						t.Fatalf("Failed to create test directory: %v", err)
					}
					if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
						t.Fatalf("Failed to create test file: %v", err)
					}
				}
			}

			ts := &services.TemplateService{}
			ts.TemplateFiles = []string{}
			for path := range tt.setupFiles {
				ts.TemplateFiles = append(ts.TemplateFiles, filepath.Join(testDir, path))
			}

			// Act
			err := ts.ParseTemplates()

			// Assert
			if tt.expectError { //nolint:nestif // test
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
				if len(ts.Templates) != len(tt.expectedFiles) {
					t.Errorf("Expected %d templates, got %d", len(tt.expectedFiles), len(ts.Templates))
				}
				for _, file := range tt.expectedFiles {
					found := false
					for _, tmpl := range ts.Templates {
						if path.Join("testdata", tmpl.Name()) == file {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected template for file %s not found", file)
					}
				}
			}

			// Cleanup
			os.RemoveAll(testDir)
		})
	}
}

func TestCreateTemplateValuesMap(t *testing.T) {
	tests := []struct {
		name           string
		templates      []*template.Template
		expectedValues types.TemplateValuesMap
		expectedErrors int
	}{
		{
			name: "Single template with flat keys",
			templates: []*template.Template{
				template.Must(template.New("tmpl1").Parse("{{.Key1}} {{.Key2}}")),
			},
			expectedValues: types.TemplateValuesMap{
				"Key1": "",
				"Key2": "",
			},
			expectedErrors: 0,
		},
		{
			name: "Multiple templates with nested keys",
			templates: []*template.Template{
				template.Must(template.New("tmpl1").Parse("{{.Key1.Nested1}} {{.Key1.Nested2}}")),
				template.Must(template.New("tmpl2").Parse("{{.Key2.Nested1}} {{.Key2.Nested2}}")),
			},
			expectedValues: types.TemplateValuesMap{
				"Key1": map[string]any{
					"Nested1": "",
					"Nested2": "",
				},
				"Key2": map[string]any{
					"Nested1": "",
					"Nested2": "",
				},
			},
			expectedErrors: 0,
		},
		{
			name: "Template with unsupported node",
			templates: []*template.Template{
				template.Must(template.New("tmpl1").Parse("{{.Key1}} {{if .Condition}}Conditional{{end}}")),
			},
			expectedValues: types.TemplateValuesMap{
				"Key1": "",
			},
			expectedErrors: 1,
		},
		{
			name:           "Empty templates list",
			templates:      []*template.Template{},
			expectedValues: types.TemplateValuesMap{},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ts := &services.TemplateService{
				Templates: tt.templates,
			}

			// Act
			ts.CreateTemplateValuesMap()

			// Assert
			if len(ts.TemplateValuesMap) != len(tt.expectedValues) {
				t.Errorf("Expected values map size %d, got %d", len(tt.expectedValues), len(ts.TemplateValuesMap))
			}

			for key, expectedValue := range tt.expectedValues {
				actualValue, exists := ts.TemplateValuesMap[key]
				if !exists {
					t.Errorf("Expected key %s not found in values map", key)
				} else if fmt.Sprintf("%v", actualValue) != fmt.Sprintf("%v", expectedValue) {
					t.Errorf("For key %s, expected value %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}
