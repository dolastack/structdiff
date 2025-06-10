package compare

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclparse"
)

type HCLJSONValidator struct{}

func (h *HCLJSONValidator) Validate(content []byte) error {
	parser := hclparse.NewParser()
	_, diags := parser.ParseJSON(content, "validation.hcl.json")
	return diags.Errs()[0]
}

func (h *HCLJSONValidator) ValidationHelp() string {
	return `HCL JSON validation tips:
• Must be valid JSON first
• Follow HCL's JSON schema
• Attributes must be properly formatted
• Check for correct JSON syntax`
}

type HCLJSONComparator struct {
	HCLJSONValidator
}

func (h *HCLJSONComparator) Compare(file1, file2 string, ignoreCase bool, config RemoteConfig) ([]Diff, error) {
	data1, err := readFileContent(file1, config)
	if err != nil {
		return nil, fmt.Errorf("error reading first file: %w", err)
	}

	data2, err := readFileContent(file2, config)
	if err != nil {
		return nil, fmt.Errorf("error reading second file: %w", err)
	}

	parser := hclparse.NewParser()
	f1, diags := parser.ParseJSON(data1, file1)
	if diags.HasErrors() {
		return nil, diags
	}

	f2, diags := parser.ParseJSON(data2, file2)
	if diags.HasErrors() {
		return nil, diags
	}

	obj1, err := hclToMap(f1)
	if err != nil {
		return nil, err
	}

	obj2, err := hclToMap(f2)
	if err != nil {
		return nil, err
	}

	return CompareValues(obj1, obj2, "", ignoreCase), nil
}

func (h *HCLJSONComparator) Validator() FileValidator {
	return &h.HCLJSONValidator
}
