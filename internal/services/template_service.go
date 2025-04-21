package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"text/template/parse"

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
	CurrentFS afero.Fs
}

func NewTemplateService(currentFs afero.Fs) *TemplateService {
	return &TemplateService{
		CurrentFS: currentFs,
	}
}

/*
GetTemplateFiles walks the rootPath and returns a list of all files with the .template extension.
*/
func (ts *TemplateService) GetTemplateFiles(rootPath string) error {
	if _, err := ts.CurrentFS.Stat(rootPath); err != nil {
		return err
	}
	templateFiles := []string{}

	_ = afero.Walk(ts.CurrentFS, rootPath, func(path string, info os.FileInfo, err error) error {
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
		t, err := template.ParseFS(afero.NewIOFS(ts.CurrentFS), file)
		if err != nil {
			return err
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
				if a, ok := c.Args[0].(*parse.FieldNode); ok {
					// flat node
					if len(a.Ident) == 1 {
						valuesMap[a.Ident[0]] = ""
					}
					// 1 level node
					//
					if len(a.Ident) == 2 { //nolint:mnd
						if nestedMap, ok := valuesMap[a.Ident[0]].(map[string]any); ok {
							nestedMap[a.Ident[1]] = ""
						} else {
							newNestedMap := make(map[string]any)
							newNestedMap[a.Ident[1]] = ""
							valuesMap[a.Ident[0]] = newNestedMap
						}
					}
					// 2 level node
					if len(a.Ident) == 3 { //nolint:mnd
						if nestedMap, ok := valuesMap[a.Ident[0]].(map[string]any); ok {
							if innerMap, ok := nestedMap[a.Ident[1]].(map[string]any); ok {
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
func (ts *TemplateService) ValuesFromFile(valuesFilePath string) (map[string]any, error) {
	yamlData, err := os.ReadFile(valuesFilePath)
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
	form, valuesPopualted := ui.RenderForm(ts.TemplateValuesMap)

	err := form.Run()
	if err != nil {
		return err
	}

	fmt.Println(valuesPopualted)
	return nil
}

func (ts *TemplateService) ValidateTemplateValues( //nolint:gocognit // complexity not avoidable
	valuesMap types.TemplateValuesMap,
	unknown map[string]any,
) []string {
	missingKeys := []string{}
	// check there are no keys in valuesMap that are not in unknown
	for key, value := range valuesMap {
		if nestedMap, ok := value.(map[string]any); ok { //nolint:nestif // complexity not avoidable
			// Check nested map keys
			if unknownNestedMap, ok := unknown[key].(map[string]any); ok {
				for nestedKey, nestedValue := range nestedMap {
					if innerMap, ok := nestedValue.(map[string]any); ok {
						// Check inner map keys
						if unknownInnerMap, ok := unknownNestedMap[nestedKey].(map[string]any); ok {
							for innerKey := range innerMap {
								if _, ok := unknownInnerMap[innerKey]; !ok {
									missingKeys = append(missingKeys, fmt.Sprintf("%s.%s.%s", key, nestedKey, innerKey))
								}
							}
						} else {
							missingKeys = append(missingKeys, fmt.Sprintf("%s.%s", key, nestedKey))
						}
					} else {
						if _, ok := unknownNestedMap[nestedKey]; !ok {
							missingKeys = append(missingKeys, fmt.Sprintf("%s.%s", key, nestedKey))
						}
					}
				}
			} else {
				missingKeys = append(missingKeys, key)
			}
		} else {
			// Check top-level keys
			if _, ok := unknown[key]; !ok {
				missingKeys = append(missingKeys, key)
			}
		}
	}

	return missingKeys
}

// RenameTargetTemplateFiles renames the target files by removing the .template suffix.
func (ts *TemplateService) RenameTargetTemplateFiles(files []string) error {
	for _, file := range files {
		name := strings.TrimSuffix(file, ".template")
		err := os.Rename(file, name)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExecuteTemplates executes the parsed templates and writes the output to the target files.
func ExecuteTemplates(tm types.TargetFileToTemplateMap, tvm types.TemplateValuesMap) error {
	// path = template
	for k, v := range tm {
		f, err := os.Create(k)
		if err != nil {
			if errors.Is(err, unix.EBADF) {
				return fmt.Errorf("bad file descriptor for file %s: %w", k, err)
			}
			return err
		}
		defer f.Close()
		err = v.Execute(f, tvm)
		if err != nil {
			return err
		}
	}
	return nil
}
