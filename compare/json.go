package compare

import (
	"encoding/json"
	"fmt"
)

type JSONValidator struct{}

func (j *JSONValidator) Validate(content []byte) error {
	var dummy interface{}
	return json.Unmarshal(content, &dummy)
}

func (j *JSONValidator) ValidationHelp() string {
	return `JSON validation tips:
• Check for proper quoting of all strings
• Verify all brackets and braces are balanced
• Ensure trailing commas are removed
• Validate correct use of null, true, false literals`
}

type JSONComparator struct {
	JSONValidator
}

func (j *JSONComparator) Compare(file1, file2 string, ignoreCase bool, config RemoteConfig) ([]Diff, error) {
	data1, err := readFileContent(file1, config)
	if err != nil {
		return nil, fmt.Errorf("error reading first file: %w", err)
	}

	data2, err := readFileContent(file2, config)
	if err != nil {
		return nil, fmt.Errorf("error reading second file: %w", err)
	}

	var obj1, obj2 interface{}
	if err := json.Unmarshal(data1, &obj1); err != nil {
		return nil, fmt.Errorf("error parsing first file: %w", err)
	}
	if err := json.Unmarshal(data2, &obj2); err != nil {
		return nil, fmt.Errorf("error parsing second file: %w", err)
	}

	return CompareValues(obj1, obj2, "", ignoreCase), nil
}

func (j *JSONComparator) Validator() FileValidator {
	return &j.JSONValidator
}
