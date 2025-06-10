package compare

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

type HCLValidator struct{}

func (h *HCLValidator) Validate(content []byte) error {
	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL(content, "validation.hcl")
	if diags.HasErrors() {
		return diags
	}
	return nil
}

func (h *HCLValidator) ValidationHelp() string {
	return `HCL validation tips:
• Check for proper block syntax: block "name" { ... }
• Ensure proper attribute syntax: key = value
• Verify all braces and quotes are balanced
• Check for correct indentation`
}

type HCLComparator struct {
	HCLValidator
}

func (h *HCLComparator) Compare(file1, file2 string, ignoreCase bool, config RemoteConfig) ([]Diff, error) {
	data1, err := readFileContent(file1, config)
	if err != nil {
		return nil, fmt.Errorf("error reading first file: %w", err)
	}

	data2, err := readFileContent(file2, config)
	if err != nil {
		return nil, fmt.Errorf("error reading second file: %w", err)
	}

	parser := hclparse.NewParser()
	f1, diags := parser.ParseHCL(data1, file1)
	if diags.HasErrors() {
		return nil, fmt.Errorf("HCL parsing error: %w", diags.Errs()[0])
	}

	f2, diags := parser.ParseHCL(data2, file2)
	if diags.HasErrors() {
		return nil, fmt.Errorf("HCL parsing error: %w", diags.Errs()[0])
	}

	obj1, err := hclToMap(f1)
	if err != nil {
		return nil, fmt.Errorf("HCL conversion error: %w", err)
	}

	obj2, err := hclToMap(f2)
	if err != nil {
		return nil, fmt.Errorf("HCL conversion error: %w", err)
	}

	return CompareValues(obj1, obj2, "", ignoreCase), nil
}

// Add this missing method
func (h *HCLComparator) Validator() FileValidator {
	return &h.HCLValidator
}

func hclToMap(file *hcl.File) (map[string]interface{}, error) {
	val, diags := file.Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     []hcl.BlockHeaderSchema{},
	})
	if diags.HasErrors() {
		return nil, diags
	}

	result := make(map[string]interface{})
	for _, attr := range val.Attributes {
		ctyVal, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, diags
		}

		goVal, err := ctyToGo(ctyVal)
		if err != nil {
			return nil, err
		}

		result[attr.Name] = goVal
	}

	for _, block := range val.Blocks {
		blockMap, err := hclBlockToMap(block)
		if err != nil {
			return nil, err
		}

		if existing, exists := result[block.Type]; exists {
			if slice, ok := existing.([]interface{}); ok {
				result[block.Type] = append(slice, blockMap)
			} else {
				result[block.Type] = []interface{}{existing, blockMap}
			}
		} else {
			result[block.Type] = blockMap
		}
	}

	return result, nil
}

func hclBlockToMap(block *hcl.Block) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if len(block.Labels) > 0 {
		result["__labels__"] = block.Labels
	}

	val, diags := block.Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     []hcl.BlockHeaderSchema{},
	})
	if diags.HasErrors() {
		return nil, diags
	}

	for _, attr := range val.Attributes {
		ctyVal, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, diags
		}

		goVal, err := ctyToGo(ctyVal)
		if err != nil {
			return nil, err
		}

		result[attr.Name] = goVal
	}

	for _, nestedBlock := range val.Blocks {
		nestedMap, err := hclBlockToMap(nestedBlock)
		if err != nil {
			return nil, err
		}

		if existing, exists := result[nestedBlock.Type]; exists {
			if slice, ok := existing.([]interface{}); ok {
				result[nestedBlock.Type] = append(slice, nestedMap)
			} else {
				result[nestedBlock.Type] = []interface{}{existing, nestedMap}
			}
		} else {
			result[nestedBlock.Type] = nestedMap
		}
	}

	return result, nil
}

func ctyToGo(val cty.Value) (interface{}, error) {
	if val.IsNull() {
		return nil, nil
	}

	switch val.Type() {
	case cty.String:
		return val.AsString(), nil
	case cty.Bool:
		return val.True(), nil
	case cty.Number:
		f, _ := val.AsBigFloat().Float64()
		return f, nil
	}

	if val.Type().IsListType() || val.Type().IsSetType() || val.Type().IsTupleType() {
		var list []interface{}
		for it := val.ElementIterator(); it.Next(); {
			_, v := it.Element()
			elem, err := ctyToGo(v)
			if err != nil {
				return nil, err
			}
			list = append(list, elem)
		}
		return list, nil
	}

	if val.Type().IsMapType() || val.Type().IsObjectType() {
		m := make(map[string]interface{})
		for it := val.ElementIterator(); it.Next(); {
			k, v := it.Element()
			key, err := ctyToGo(k)
			if err != nil {
				return nil, err
			}
			keyStr, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("map key is not string: %v", key)
			}
			value, err := ctyToGo(v)
			if err != nil {
				return nil, err
			}
			m[keyStr] = value
		}
		return m, nil
	}

	return nil, fmt.Errorf("unsupported HCL type: %s", val.Type().FriendlyName())
}
