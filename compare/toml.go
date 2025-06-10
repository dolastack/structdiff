package compare

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

type TOMLValidator struct{}

func (t *TOMLValidator) Validate(content []byte) error {
	var dummy interface{}
	return toml.Unmarshal(content, &dummy)
}

func (t *TOMLValidator) ValidationHelp() string {
	return `TOML validation tips:
• Use key = "value" format
• Tables should be in [table] format
• Arrays use brackets: values = [1, 2, 3]
• Ensure proper quoting of strings`
}

type TOMLComparator struct {
	TOMLValidator
}

func (t *TOMLComparator) Compare(file1, file2 string, ignoreCase bool, config RemoteConfig) ([]Diff, error) {
	data1, err := readFileContent(file1, config)
	if err != nil {
		return nil, fmt.Errorf("error reading first file: %w", err)
	}

	data2, err := readFileContent(file2, config)
	if err != nil {
		return nil, fmt.Errorf("error reading second file: %w", err)
	}

	var obj1, obj2 interface{}
	if err := toml.Unmarshal(data1, &obj1); err != nil {
		return nil, fmt.Errorf("error parsing first TOML file: %w", err)
	}
	if err := toml.Unmarshal(data2, &obj2); err != nil {
		return nil, fmt.Errorf("error parsing second TOML file: %w", err)
	}

	return CompareValues(obj1, obj2, "", ignoreCase), nil
}

func (t *TOMLComparator) Validator() FileValidator {
	return &t.TOMLValidator
}
