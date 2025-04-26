package services

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type YamlStruct interface {
	Yamafiable()
}

// ReadYamlFromFile reads YAML data from a reader and unmarshals it into a struct that implements YamlStruct.
// It returns an error if the reader is empty or if the YAML cannot be unmarshaled.
func ReadYamlFromFile[T YamlStruct](r io.Reader) (T, error) {
	var yamlStruct T

	// Read all data from the reader
	data, err := io.ReadAll(r)
	if err != nil {
		return yamlStruct, fmt.Errorf("failed to read file: %w", err)
	}

	// Check if there is any content
	if len(data) == 0 {
		return yamlStruct, fmt.Errorf("no content in file, data length is 0")
	}

	// Debug output - can be removed if not needed
	// fmt.Printf("Read %d bytes of data\n", len(data))

	err = yaml.Unmarshal(data, &yamlStruct)
	if err != nil {
		return yamlStruct, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return yamlStruct, nil
}
