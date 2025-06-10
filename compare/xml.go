package compare

import (
	"encoding/xml"
	"fmt"
)

type XMLValidator struct{}

func (x *XMLValidator) Validate(content []byte) error {
	var dummy interface{}
	return xml.Unmarshal(content, &dummy)
}

func (x *XMLValidator) ValidationHelp() string {
	return `XML validation tips:
• Ensure proper XML declaration <?xml version="1.0"?>
• All tags must be properly closed
• Attributes must be quoted
• Special characters must be escaped`
}

type XMLComparator struct {
	XMLValidator
}

func (x *XMLComparator) Compare(file1, file2 string, ignoreCase bool, config RemoteConfig) ([]Diff, error) {
	data1, err := readFileContent(file1, config)
	if err != nil {
		return nil, fmt.Errorf("error reading first file: %w", err)
	}

	data2, err := readFileContent(file2, config)
	if err != nil {
		return nil, fmt.Errorf("error reading second file: %w", err)
	}

	var obj1, obj2 interface{}
	if err := xml.Unmarshal(data1, &obj1); err != nil {
		return nil, fmt.Errorf("error parsing first XML file: %w", err)
	}
	if err := xml.Unmarshal(data2, &obj2); err != nil {
		return nil, fmt.Errorf("error parsing second XML file: %w", err)
	}

	return CompareValues(obj1, obj2, "", ignoreCase), nil
}

func (x *XMLComparator) Validator() FileValidator {
	return &x.XMLValidator
}
