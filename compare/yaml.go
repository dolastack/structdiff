package compare

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

type YAMLValidator struct{}

func (y *YAMLValidator) Validate(content []byte) error {
	var dummy interface{}
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	return decoder.Decode(&dummy)
}

func (y *YAMLValidator) ValidationHelp() string {
	return `YAML validation tips:
• Use spaces for indentation (no tabs)
• Ensure proper colon usage in key: value pairs
• Quote strings with special characters
• Check for consistent indentation levels
• Verify multiline strings use proper syntax (| or >)`
}

type YAMLComparator struct {
	YAMLValidator
}

func (y *YAMLComparator) Compare(file1, file2 string, ignoreCase bool, config RemoteConfig) ([]Diff, error) {
	data1, err := readFileContent(file1, config)
	if err != nil {
		return nil, fmt.Errorf("error reading first file: %w", err)
	}

	data2, err := readFileContent(file2, config)
	if err != nil {
		return nil, fmt.Errorf("error reading second file: %w", err)
	}

	var obj1, obj2 interface{}
	if err := yaml.Unmarshal(data1, &obj1); err != nil {
		return nil, fmt.Errorf("error parsing first file: %w", err)
	}
	if err := yaml.Unmarshal(data2, &obj2); err != nil {
		return nil, fmt.Errorf("error parsing second file: %w", err)
	}

	return CompareValues(obj1, obj2, "", ignoreCase), nil
}

func (y *YAMLComparator) Validator() FileValidator {
	return &y.YAMLValidator
}
