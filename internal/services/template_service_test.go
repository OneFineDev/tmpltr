//go:build !integration

package services_test

import (
	"reflect"
	"strings"
	"testing"
	"text/template"

	"github.com/OneFineDev/tmpltr/internal/services"
	"github.com/OneFineDev/tmpltr/internal/storage"
	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/spf13/afero"
)

func TestGetTemplateFiles(t *testing.T) { //nolint:gocognit
	// Arrange
	tests := []struct {
		name          string
		setupFs       func(fs afero.Fs)
		rootPath      string
		expectedFiles []string
		expectError   bool
	}{
		{
			name: "Valid template files",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("/templates", 0755)
				_ = afero.WriteFile(fs, "/templates/file1.template", []byte{}, 0755)
				_ = afero.WriteFile(fs, "/templates/file2.template", []byte{}, 0755)
				_ = afero.WriteFile(fs, "/templates/file3.txt", []byte{}, 0755)
			},
			rootPath:      "/templates",
			expectedFiles: []string{"/templates/file1.template", "/templates/file2.template"},
			expectError:   false,
		},
		{
			name: "No template files",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("/empty", 0755)
			},
			rootPath:      "/empty",
			expectedFiles: []string{},
			expectError:   false,
		},
		{
			name: "Root path does not exist",
			setupFs: func(_ afero.Fs) {
				// No setup needed
			},
			rootPath:    "/nonexistent",
			expectError: true,
		},
		{
			name: "Skip .git directory",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("/repo/.git", 0755)
				_ = fs.MkdirAll("/repo/templates", 0755)
				_ = afero.WriteFile(fs, "/repo/templates/file1.template", []byte{}, 0755)
				_ = afero.WriteFile(fs, "/repo/.git/file2.template", []byte{}, 0755)
			},
			rootPath:      "/repo",
			expectedFiles: []string{"/repo/templates/file1.template"},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}
			safeFs := &storage.SafeFs{Fs: fs}
			service := services.NewTemplateService(safeFs)

			// Act
			err := service.GetTemplateFiles(tt.rootPath)

			// Assert
			if tt.expectError { //nolint:nestif
				if err == nil {
					t.Errorf("expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(service.TemplateFiles) != len(tt.expectedFiles) {
					t.Errorf("expected %d files, got %d", len(tt.expectedFiles), len(service.TemplateFiles))
				}
				for _, expectedFile := range tt.expectedFiles {
					found := false
					for _, actualFile := range service.TemplateFiles {
						if expectedFile == actualFile {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected file %s not found in result", expectedFile)
					}
				}
			}
		})
	}
}
func TestParseTemplates(t *testing.T) { //nolint:gocognit
	// Arrange
	tests := []struct {
		name          string
		setupFs       func(fs afero.Fs)
		templateFiles []string
		expectError   bool
	}{
		{
			name: "Valid templates",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("templates", 0755)
				_ = afero.WriteFile(fs, "templates/file1.template", []byte("Hello, {{.Name}}!"), 0755)
				_ = afero.WriteFile(fs, "templates/file2.template", []byte("Welcome to {{.Place}}."), 0755)
			},
			templateFiles: []string{"templates/file1.template", "templates/file2.template"},
			expectError:   false,
		},
		{
			name: "Invalid template syntax",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("templates", 0755)
				_ = afero.WriteFile(fs, "templates/file1.template", []byte("Hello, {{.Name!"), 0755)
			},
			templateFiles: []string{"templates/file1.template"},
			expectError:   true,
		},
		{
			name: "No templates provided",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("templates", 0755)
			},
			templateFiles: []string{},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}
			safeFs := &storage.SafeFs{Fs: fs}
			service := services.NewTemplateService(safeFs)
			service.TemplateFiles = tt.templateFiles

			// Act
			err := service.ParseTemplates()

			// Assert
			if tt.expectError { //nolint:nestif
				if err == nil {
					t.Errorf("expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(service.Templates) != len(tt.templateFiles) {
					t.Errorf("expected %d templates, got %d", len(tt.templateFiles), len(service.Templates))
				}
				for _, tmpl := range service.Templates {
					if tmpl == nil {
						t.Errorf("expected a valid template but got nil")
					}
				}
			}
		})
	}
}
func TestExtractTemplateKeys(t *testing.T) {
	// Arrange
	tests := []struct {
		name            string
		templateContent string
		expectedValues  types.TemplateValuesMap
		expectErrors    bool
	}{
		{
			name:            "Flat keys",
			templateContent: "Hello, {{.Name}}!",
			expectedValues: types.TemplateValuesMap{
				"Name": "",
			},
			expectErrors: false,
		},
		{
			name:            "Nested keys",
			templateContent: "Welcome to {{.Location.City}}, {{.Location.Country}}!",
			expectedValues: types.TemplateValuesMap{
				"Location": map[string]any{
					"City":    "",
					"Country": "",
				},
			},
			expectErrors: false,
		},
		{
			name:            "Doubly nested keys",
			templateContent: "User: {{.User.Profile.Name}}",
			expectedValues: types.TemplateValuesMap{
				"User": map[string]any{
					"Profile": map[string]any{
						"Name": "",
					},
				},
			},
			expectErrors: false,
		},
		{
			name:            "Unsupported node",
			templateContent: "Hello, {{.Name}}! {{range .Items}}{{.}}{{end}}",
			expectedValues: types.TemplateValuesMap{
				"Name": "",
			},
			expectErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tmpl, err := template.New("test").Parse(tt.templateContent)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}
			valuesMap := make(types.TemplateValuesMap)
			service := &services.TemplateService{}

			// Act
			errors := service.ExtractTemplateKeys(tmpl, valuesMap)

			// Assert
			if tt.expectErrors {
				if len(errors) == 0 {
					t.Errorf("expected errors but got none")
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("unexpected errors: %v", errors)
				}
				if !reflect.DeepEqual(valuesMap, tt.expectedValues) {
					t.Errorf("expected values map %v, got %v", tt.expectedValues, valuesMap)
				}
			}
		})
	}
}
func TestValuesFromFile(t *testing.T) {
	// Arrange
	tests := []struct {
		name        string
		yamlContent string
		expectedMap map[string]any
		expectError bool
	}{
		{
			name: "Valid YAML content",
			yamlContent: `
name: John Doe
age: 30
address:
  city: New York
  country: USA
`,
			expectedMap: map[string]any{
				"name": "John Doe",
				"age":  30,
				"address": map[string]any{
					"city":    "New York",
					"country": "USA",
				},
			},
			expectError: false,
		},
		{
			name:        "Invalid YAML content",
			yamlContent: `name: John Doe: age: 30`,
			expectedMap: nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			reader := strings.NewReader(tt.yamlContent)
			service := &services.TemplateService{}

			// Act
			result, err := service.ValuesFromFile(reader)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(result, tt.expectedMap) {
					t.Errorf("expected map %v, got %v", tt.expectedMap, result)
				}
			}
		})
	}
}
func TestCreateTemplateValuesMap(t *testing.T) { //nolint:gocognit
	// Arrange
	tests := []struct {
		name           string
		templates      []string
		expectedValues types.TemplateValuesMap
		expectErrors   bool
	}{
		{
			name: "Single template with flat keys",
			templates: []string{
				"Hello, {{.Name}}!",
			},
			expectedValues: types.TemplateValuesMap{
				"Name": "",
			},
			expectErrors: false,
		},
		{
			name: "Multiple templates with nested keys",
			templates: []string{
				"Welcome to {{.Location.City}}, {{.Location.Country}}!",
				"User: {{.User.Profile.Name}}",
			},
			expectedValues: types.TemplateValuesMap{
				"Location": map[string]any{
					"City":    "",
					"Country": "",
				},
				"User": map[string]any{
					"Profile": map[string]any{
						"Name": "",
					},
				},
			},
			expectErrors: false,
		},
		{
			name: "Template with unsupported node",
			templates: []string{
				"Hello, {{.Name}}! {{range .Items}}{{.}}{{end}}",
			},
			expectedValues: types.TemplateValuesMap{
				"Name": "",
			},
			expectErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			fs := afero.NewMemMapFs()
			safeFs := &storage.SafeFs{Fs: fs}
			service := services.NewTemplateService(safeFs)

			templates := make([]*template.Template, len(tt.templates))
			for i, tmplContent := range tt.templates {
				tmpl, err := template.New("test").Parse(tmplContent)
				if err != nil {
					t.Fatalf("failed to parse template: %v", err)
				}
				templates[i] = tmpl
			}
			service.Templates = templates

			// Act
			service.CreateTemplateValuesMap()

			// Assert
			if tt.expectErrors {
				// Check if any errors occurred during key extraction
				for _, tmpl := range templates {
					errors := service.ExtractTemplateKeys(tmpl, service.TemplateValuesMap)
					if len(errors) == 0 {
						t.Errorf("expected errors but got none")
					}
				}
			} else if !reflect.DeepEqual(service.TemplateValuesMap, tt.expectedValues) {
				t.Errorf("expected values map %v, got %v", tt.expectedValues, service.TemplateValuesMap)
			}
		})
	}
}

func TestExecuteTemplates(t *testing.T) { //nolint:gocognit
	// Arrange
	tests := []struct {
		name                 string
		setupFs              func(fs afero.Fs)
		targetFileToTemplate map[string]*template.Template
		templateValuesMap    types.TemplateValuesMap
		expectError          bool
		expectedOutputs      map[string]string
	}{
		{
			name: "Valid templates execution",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("/output", 0755)
			},
			targetFileToTemplate: map[string]*template.Template{
				"/output/file1.txt": template.Must(template.New("file1").Parse("Hello, {{.Name}}!")),
				"/output/file2.txt": template.Must(template.New("file2").Parse("Welcome to {{.Place}}.")),
			},
			templateValuesMap: types.TemplateValuesMap{
				"Name":  "John",
				"Place": "Earth",
			},
			expectError: false,
			expectedOutputs: map[string]string{
				"/output/file1.txt": "Hello, John!",
				"/output/file2.txt": "Welcome to Earth.",
			},
		},
		{
			name: "Template execution with missing values",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("/output", 0755)
			},
			targetFileToTemplate: map[string]*template.Template{
				"/output/file1.txt": template.Must(
					template.New("file1").Option("missingkey=error").Parse("Hello, {{.Name}}!"),
				),
			},
			templateValuesMap: types.TemplateValuesMap{},
			expectError:       true,
			expectedOutputs:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}
			safeFs := &storage.SafeFs{Fs: fs}
			service := services.NewTemplateService(safeFs)
			service.TargetFileToTemplateMap = tt.targetFileToTemplate
			service.TemplateValuesMap = tt.templateValuesMap

			// Act
			err := service.ExecuteTemplates()

			// Assert
			if tt.expectError { //nolint:nestif
				if err == nil {
					t.Errorf("expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				for filePath, expectedContent := range tt.expectedOutputs {
					actualContent, readErr := afero.ReadFile(fs, filePath)
					if readErr != nil {
						t.Errorf("failed to read file %s: %v", filePath, readErr)
					}
					if string(actualContent) != expectedContent {
						t.Errorf("expected content %q, got %q", expectedContent, string(actualContent))
					}
				}
			}
		})
	}
}
func TestRenameTargetTemplateFiles(t *testing.T) { //nolint:gocognit
	// Arrange
	tests := []struct {
		name          string
		setupFs       func(fs afero.Fs)
		targetFiles   map[string]*template.Template
		expectedFiles []string
		expectError   bool
	}{
		{
			name: "Rename valid template files",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "/file1.template", []byte{}, 0755)
				_ = afero.WriteFile(fs, "/file2.template", []byte{}, 0755)
			},
			targetFiles: map[string]*template.Template{
				"/file1.template": nil,
				"/file2.template": nil,
			},
			expectedFiles: []string{"/file1", "/file2"},
			expectError:   false,
		},
		{
			name: "File does not exist",
			setupFs: func(_ afero.Fs) {
				// No setup needed
			},
			targetFiles: map[string]*template.Template{
				"/nonexistent.template": nil,
			},
			expectedFiles: nil,
			expectError:   true,
		},
		{
			name: "Partial rename failure",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "/file1.template", []byte{}, 0755)
				// Simulate a file that cannot be renamed
			},
			targetFiles: map[string]*template.Template{
				"/file1.template": nil,
				"/file2.template": nil,
			},
			expectedFiles: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}
			safeFs := &storage.SafeFs{Fs: fs}
			service := services.NewTemplateService(safeFs)
			service.TargetFileToTemplateMap = tt.targetFiles

			// Act
			err := service.RenameTargetTemplateFiles()

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				for _, expectedFile := range tt.expectedFiles {
					exists, _ := afero.Exists(fs, expectedFile)
					if !exists {
						t.Errorf("expected file %s to exist but it does not", expectedFile)
					}
				}
			}
		})
	}
}
