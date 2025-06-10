package compare

import (
	"fmt"

	"gopkg.in/ini.v1"
)

type INIValidator struct{}

func (i *INIValidator) Validate(content []byte) error {
	_, err := ini.Load(content)
	return err
}

func (i *INIValidator) ValidationHelp() string {
	return `INI validation tips:
• Ensure each section is in brackets [section]
• Use key=value format for properties
• Comments start with ; or #
• Escape special characters properly`
}

type INIComparator struct {
	INIValidator
}

func (i *INIComparator) Compare(file1, file2 string, ignoreCase bool, config RemoteConfig) ([]Diff, error) {
	data1, err := readFileContent(file1, config)
	if err != nil {
		return nil, fmt.Errorf("error reading first file: %w", err)
	}

	data2, err := readFileContent(file2, config)
	if err != nil {
		return nil, fmt.Errorf("error reading second file: %w", err)
	}

	cfg1, err := ini.Load(data1)
	if err != nil {
		return nil, fmt.Errorf("error parsing first INI file: %w", err)
	}

	cfg2, err := ini.Load(data2)
	if err != nil {
		return nil, fmt.Errorf("error parsing second INI file: %w", err)
	}

	obj1 := iniToMap(cfg1)
	obj2 := iniToMap(cfg2)

	return CompareValues(obj1, obj2, "", ignoreCase), nil
}

func (i *INIComparator) Validator() FileValidator {
	return &i.INIValidator
}

func iniToMap(cfg *ini.File) map[string]interface{} {
	result := make(map[string]interface{})
	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}
		sectionMap := make(map[string]interface{})
		for _, key := range section.Keys() {
			sectionMap[key.Name()] = key.Value()
		}
		result[section.Name()] = sectionMap
	}
	return result
}
