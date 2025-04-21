package services

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type YamlStruct interface {
	Yamafiable()
}

func readYamlFromFile[T YamlStruct](filePath string) (T, error) {
	var yamlStruct T

	data, err := os.ReadFile(filePath)
	if err != nil {
		return yamlStruct, fmt.Errorf("failed to read file: %w", err)
	}

	err = yaml.Unmarshal(data, &yamlStruct)
	if err != nil {
		return yamlStruct, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return yamlStruct, nil
}
