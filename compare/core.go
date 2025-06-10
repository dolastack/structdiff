package compare

import (
	"fmt"
	"reflect"
	"strings"
)

type DiffType string

const (
	DiffAdded    DiffType = "added"
	DiffRemoved  DiffType = "removed"
	DiffModified DiffType = "modified"
	DiffMoved    DiffType = "moved"
)

type Diff struct {
	Type     DiffType    `json:"type"`
	Path     string      `json:"path"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
}

type FileValidator interface {
	Validate(content []byte) error
	ValidationHelp() string
}

func CompareValues(a, b interface{}, path string, ignoreCase bool) []Diff {
	var diffs []Diff

	if a == nil || b == nil {
		if a != b {
			diffType := DiffAdded
			oldVal, newVal := a, b
			if a == nil {
				diffType = DiffRemoved
				oldVal, newVal = b, a
			}
			diffs = append(diffs, Diff{
				Type:     diffType,
				Path:     path,
				OldValue: oldVal,
				NewValue: newVal,
			})
		}
		return diffs
	}

	aType := reflect.TypeOf(a)
	bType := reflect.TypeOf(b)
	if aType != bType {
		diffs = append(diffs, Diff{
			Type:     DiffModified,
			Path:     path,
			OldValue: a,
			NewValue: b,
		})
		return diffs
	}

	switch aVal := a.(type) {
	case map[string]interface{}:
		bVal := b.(map[string]interface{})
		diffs = append(diffs, CompareMaps(aVal, bVal, path, ignoreCase)...)
	case []interface{}:
		bVal := b.([]interface{})
		diffs = append(diffs, CompareSlices(aVal, bVal, path, ignoreCase)...)
	default:
		if !DeepEqual(a, b, ignoreCase) {
			diffs = append(diffs, Diff{
				Type:     DiffModified,
				Path:     path,
				OldValue: a,
				NewValue: b,
			})
		}
	}
	return diffs
}

func CompareMaps(a, b map[string]interface{}, path string, ignoreCase bool) []Diff {
	var diffs []Diff
	allKeys := make(map[string]struct{})

	for key := range a {
		allKeys[key] = struct{}{}
	}
	for key := range b {
		allKeys[key] = struct{}{}
	}

	for key := range allKeys {
		fullPath := path
		if fullPath != "" {
			fullPath += "."
		}
		fullPath += key

		aVal, aExists := a[key]
		bVal, bExists := b[key]

		switch {
		case aExists && bExists:
			diffs = append(diffs, CompareValues(aVal, bVal, fullPath, ignoreCase)...)
		case aExists:
			diffs = append(diffs, Diff{
				Type:     DiffRemoved,
				Path:     fullPath,
				OldValue: aVal,
			})
		case bExists:
			diffs = append(diffs, Diff{
				Type:     DiffAdded,
				Path:     fullPath,
				NewValue: bVal,
			})
		}
	}
	return diffs
}

func CompareSlices(a, b []interface{}, path string, ignoreCase bool) []Diff {
	var diffs []Diff
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	for i := 0; i < maxLen; i++ {
		fullPath := fmt.Sprintf("%s[%d]", path, i)
		switch {
		case i >= len(a):
			diffs = append(diffs, Diff{
				Type:     DiffAdded,
				Path:     fullPath,
				NewValue: b[i],
			})
		case i >= len(b):
			diffs = append(diffs, Diff{
				Type:     DiffRemoved,
				Path:     fullPath,
				OldValue: a[i],
			})
		default:
			diffs = append(diffs, CompareValues(a[i], b[i], fullPath, ignoreCase)...)
		}
	}
	return diffs
}

func DeepEqual(a, b interface{}, ignoreCase bool) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	switch aVal := a.(type) {
	case string:
		bVal, ok := b.(string)
		if !ok {
			return false
		}
		if ignoreCase {
			return strings.EqualFold(aVal, bVal)
		}
		return aVal == bVal
	case bool:
		bVal, ok := b.(bool)
		return ok && aVal == bVal
	case float64:
		bVal, ok := b.(float64)
		return ok && aVal == bVal
	default:
		return a == b
	}
}
