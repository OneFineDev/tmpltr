package ui

import (
	"strings"

	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/charmbracelet/huh"
)

func RenderForm(valuesMap types.TemplateValuesMap) (*huh.Form, map[string]*string) {
	outMap := make(map[string]*string)

	Flatten("", valuesMap, outMap)

	inputSlice := []huh.Field{}

	for k := range outMap {
		value := outMap[k]
		i := huh.NewInput().Description(k).Inline(true).Value(value)

		inputSlice = append(inputSlice, i)
	}

	form := huh.NewForm(
		huh.NewGroup(inputSlice...),
	)

	return form, outMap
}

func Flatten(prefix string, src map[string]interface{}, dest map[string]*string) {
	if len(prefix) > 0 {
		prefix += "."
	}
	for k, v := range src {
		switch child := v.(type) {
		case map[string]interface{}:
			Flatten(prefix+k, child, dest)
		// case []interface{}:
		// 	for i := 0; i < len(child); i++ {
		// 		dest[prefix+k+"."+strconv.Itoa(i)] = child[i]
		// 	}
		default:
			dest[prefix+k] = new(string)
		}
	}
}

func Rebuild(formMap map[string]*string, templateValuesMap map[string]any) map[string]any {
	for k, v := range formMap {
		keys := strings.Split(k, ".")

		switch len(keys) {
		case 1:
			key := keys[0]
			templateValuesMap[key] = *v
		case 2: //nolint:mnd // it's fine
			key1 := keys[0]
			key2 := keys[1]
			if nestedMap, okNested := templateValuesMap[key1].(map[string]any); okNested {
				nestedMap[key2] = *v
			} else {
				nMap := make(map[string]any)
				nMap[key2] = *v
				templateValuesMap[key1] = nMap
			}
		case 3: //nolint:mnd // it's fine
			key1 := keys[0]
			key2 := keys[1]
			key3 := keys[2]
			if nestedMap, okNested := templateValuesMap[key1].(map[string]any); okNested {
				if innerMap, okInner := nestedMap[key2].(map[string]any); okInner {
					innerMap[key3] = *v
				} else {
					inMap := make(map[string]any)
					inMap[key3] = *v
					nestedMap[key2] = inMap
				}
			} else {
				nMap := make(map[string]any)
				innerMap := make(map[string]any)
				innerMap[key3] = *v
				nMap[key2] = innerMap
				templateValuesMap[key1] = nMap
			}
		default:
			// Handle deeper nesting if required
		}
	}
	return templateValuesMap
}
