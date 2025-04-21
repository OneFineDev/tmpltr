package types

import "text/template"

type TemplateValuesMap map[string]any
type TargetFileToTemplateMap map[string]*template.Template
type ValuesInputType string
