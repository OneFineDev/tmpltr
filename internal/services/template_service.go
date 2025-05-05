package services

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/OneFineDev/tmpltr/internal/storage"
	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/OneFineDev/tmpltr/internal/ui"
	"github.com/spf13/afero"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v3"
)

type TemplateService struct {
	TemplateFiles []string
	types.TargetFileToTemplateMap
	Templates []*template.Template
	types.TemplateValuesMap
	CurrentFS *storage.SafeFs
}

func NewTemplateService(currentFs *storage.SafeFs) *TemplateService {
	return &TemplateService{
		CurrentFS: currentFs,
	}
}

func (ts *TemplateService) HandleTemplates() error {
	return nil
}

/*
GetTemplateFiles walks the rootPath and returns a list of all files with the .template extension.
*/
func (ts *TemplateService) GetTemplateFiles(rootPath string) error {
	if _, err := ts.CurrentFS.Fs.Stat(rootPath); err != nil {
		return err
	}
	templateFiles := []string{}

	_ = afero.Walk(ts.CurrentFS.Fs, rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		} else if filepath.Ext(path) == ".template" {
			templateFiles = append(templateFiles, path)
			return nil
		}
		return nil
	})
	ts.TemplateFiles = templateFiles
	return nil
}

// ParseTemplates parses the template files and returns a map of target file paths to its corresponding parsed template.
func (ts *TemplateService) ParseTemplates() error {
	targetFileToTemplateMap := make(types.TargetFileToTemplateMap)

	for _, file := range ts.TemplateFiles {
		// Read the file content directly from Afero filesystem
		content, err := afero.ReadFile(ts.CurrentFS.Fs, file)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", file, err)
		}

		// Parse the template from string content instead of using ParseFS
		t, err := template.New(filepath.Base(file)).Option("missingkey=error").Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", file, err)
		}

		targetFileToTemplateMap[file] = t
	}

	ts.TargetFileToTemplateMap = targetFileToTemplateMap

	templates := make([]*template.Template, len(targetFileToTemplateMap))
	i := 0
	for _, tmpl := range ts.TargetFileToTemplateMap {
		templates[i] = tmpl
		i++
	}

	ts.Templates = templates
	return nil
}

/*
ExtractTemplateKeys extracts the template keys from the parsed template and populates the valuesMap
with the keys. This map can then be populated either interactively or from a file, and then used
to execute a template.Can only currently handle unnested keys.
*/
func (ts *TemplateService) ExtractTemplateKeys( //nolint:gocognit // complexity not avoidable
	t *template.Template,
	valuesMap types.TemplateValuesMap,
) []error {
	var errors []error
	ln := t.Tree.Root
Node:
	for _, n := range ln.Nodes {
		if nn, ok := n.(*parse.ActionNode); ok { //nolint:nestif // complexity not avoidable
			p := nn.Pipe
			if len(p.Decl) > 0 {
				errors = append(errors, fmt.Errorf("node %v not supported", n))
				continue Node
			}
			for _, c := range p.Cmds {
				if len(c.Args) != 1 {
					errors = append(errors, fmt.Errorf("node %v not supported", n))
					continue Node
				}
				if a, fieldOk := c.Args[0].(*parse.FieldNode); fieldOk {
					// flat node
					if len(a.Ident) == 1 {
						valuesMap[a.Ident[0]] = ""
					}
					// 1 level node
					//
					if len(a.Ident) == 2 { //nolint:mnd
						if nestedMap, identOk := valuesMap[a.Ident[0]].(map[string]any); identOk {
							nestedMap[a.Ident[1]] = ""
						} else {
							newNestedMap := make(map[string]any)
							newNestedMap[a.Ident[1]] = ""
							valuesMap[a.Ident[0]] = newNestedMap
						}
					}
					// 2 level node
					if len(a.Ident) == 3 { //nolint:mnd
						if nestedMap, l2IdentOk := valuesMap[a.Ident[0]].(map[string]any); l2IdentOk {
							if innerMap, innerOk := nestedMap[a.Ident[1]].(map[string]any); innerOk {
								innerMap[a.Ident[2]] = ""
							} else {
								newInnerMap := make(map[string]any)
								newInnerMap[a.Ident[2]] = ""
								nestedMap[a.Ident[1]] = newInnerMap
							}
						} else {
							newNestedMap := make(map[string]any)
							newInnerMap := make(map[string]any)
							newInnerMap[a.Ident[2]] = ""
							newNestedMap[a.Ident[1]] = newInnerMap
							valuesMap[a.Ident[0]] = newNestedMap
						}
					}
				} else {
					errors = append(errors, fmt.Errorf("node %v not supported", n))
					continue Node
				}
			}
		} else {
			if _, innerOk := n.(*parse.TextNode); !innerOk {
				errors = append(errors, fmt.Errorf("node %v not supported", n))
				continue Node
			}
		}
	}
	return errors
}

// ValuesFromFile reads the key:value pairs in values file and returns a map of these.
func (ts *TemplateService) ValuesFromFile(r io.Reader) (map[string]any, error) {
	yamlData, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var unknown map[string]any

	err = yaml.Unmarshal(yamlData, &unknown)
	if err != nil {
		return nil, err
	}

	return unknown, nil
}

// CreateTemplateValuesMap creates a map of template keys from parsed templates to empty strings to be populated later.
func (ts *TemplateService) CreateTemplateValuesMap() {
	m := make(types.TemplateValuesMap)

	ts.TemplateValuesMap = m

	for _, tmpl := range ts.Templates {
		ts.ExtractTemplateKeys(tmpl, m)
	}
}

func (ts *TemplateService) InteractiveInput() error {
	form, _ := ui.RenderForm(ts.TemplateValuesMap)

	err := form.Run()
	if err != nil {
		return err
	}
	return nil
}

func (ts *TemplateService) ValidateTemplateValues( //nolint:gocognit // complexity not avoidable
	valuesMap types.TemplateValuesMap,
	unknown map[string]any,
) []string {
	missingKeys := []string{}
	// check there are no keys in valuesMap that are not in unknown
	for key, value := range valuesMap {
		if nestedMap, okNested := value.(map[string]any); okNested { //nolint:nestif // complexity not avoidable
			// Check nested map keys
			if unknownNestedMap, okUnknownNested := unknown[key].(map[string]any); okUnknownNested {
				for nestedKey, nestedValue := range nestedMap {
					if innerMap, okInner := nestedValue.(map[string]any); okInner {
						// Check inner map keys
						if unknownInnerMap, okUnknownInner := unknownNestedMap[nestedKey].(map[string]any); okUnknownInner {
							for innerKey := range innerMap {
								if _, okInnerKey := unknownInnerMap[innerKey]; !okInnerKey {
									missingKeys = append(missingKeys, fmt.Sprintf("%s.%s.%s", key, nestedKey, innerKey))
								}
							}
						} else {
							missingKeys = append(missingKeys, fmt.Sprintf("%s.%s", key, nestedKey))
						}
					} else {
						if _, okNestedKey := unknownNestedMap[nestedKey]; !okNestedKey {
							missingKeys = append(missingKeys, fmt.Sprintf("%s.%s", key, nestedKey))
						}
					}
				}
			} else {
				missingKeys = append(missingKeys, key)
			}
		} else {
			// Check top-level keys
			if _, okTopLevel := unknown[key]; !okTopLevel {
				missingKeys = append(missingKeys, key)
			}
		}
	}

	return missingKeys
}

// RenameTargetTemplateFiles renames the target files by removing the .template suffix.
func (ts *TemplateService) RenameTargetTemplateFiles() error {
	for k := range ts.TargetFileToTemplateMap {
		name := strings.TrimSuffix(k, ".template")
		err := ts.CurrentFS.Fs.Rename(k, name)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExecuteTemplates executes the parsed templates and writes the output to the target files.
func (ts *TemplateService) ExecuteTemplates() error {
	// path = template
	for k, v := range ts.TargetFileToTemplateMap {
		f, err := ts.CurrentFS.Fs.Create(k)
		if err != nil {
			if errors.Is(err, unix.EBADF) {
				return fmt.Errorf("bad file descriptor for file %s: %w", k, err)
			}
			return err
		}
		defer f.Close()
		err = v.Execute(f, ts.TemplateValuesMap)
		if err != nil {
			return err
		}
	}
	return nil
}
