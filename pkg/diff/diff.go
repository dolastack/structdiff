package diff

import (
	"fmt"
	"go/parser"
	"reflect"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

func CompareFiles(file1, file2 string, quiet bool, filterPaths []string) (string, error) {
	d1, err := parser.ParseFile(file1)
	if err != nil {
		return "", err
	}
	d2, err := parser.ParseFile(file2)
	if err != nil {
		return "", err
	}

	if d1.Type != d2.Type {
		return "", fmt.Errorf("files must be of the same type")
	}

	var data1, data2 interface{}
	if len(filterPaths) == 0 {
		data1 = d1.Data
		data2 = d2.Data
	} else {
		data1 = applyFilters(d1.Data, filterPaths)
		data2 = applyFilters(d2.Data, filterPaths)
	}

	diff := compareRecursive(data1, data2, "")
	return formatDiff(diff, quiet), nil
}

func compareRecursive(a, b interface{}, prefix string) string {
	var sb strings.Builder

	ta := reflect.TypeOf(a)
	tb := reflect.TypeOf(b)
	if ta != tb {
		sb.WriteString(color.RedString("%sType mismatch: %v vs %v\n", prefix, ta, tb))
		return sb.String()
	}

	switch aVal := a.(type) {
	case map[string]interface{}:
		bVal := b.(map[string]interface{})
		sb.WriteString(compareMaps(aVal, bVal, prefix))

	case []interface{}:
		bVal := b.([]interface{})
		sb.WriteString(compareSlices(aVal, bVal, prefix))

	default:
		if fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b) {
			sb.WriteString(color.YellowString("%sValue mismatch: '%v' vs '%v'\n", prefix, a, b))
		}
	}

	return sb.String()
}

func compareMaps(a, b map[string]interface{}, prefix string) string {
	var sb strings.Builder

	allKeys := make(map[string]bool)
	for k := range a {
		allKeys[k] = true
	}
	for k := range b {
		allKeys[k] = true
	}

	for k := range allKeys {
		foundInA := false
		foundInB := false
		var keyA, keyB string

		for ka := range a {
			if strings.ToLower(ka) == strings.ToLower(k) {
				keyA = ka
				foundInA = true
				break
			}
		}
		for kb := range b {
			if strings.ToLower(kb) == strings.ToLower(k) {
				keyB = kb
				foundInB = true
				break
			}
		}

		if !foundInA {
			sb.WriteString(color.RedString("%sKey '%s' missing in first file\n", prefix, k))
			continue
		}
		if !foundInB {
			sb.WriteString(color.RedString("%sKey '%s' missing in second file\n", prefix, k))
			continue
		}

		newPrefix := k
		if prefix != "" {
			newPrefix = prefix + "." + k
		}

		sb.WriteString(compareRecursive(a[keyA], b[keyB], newPrefix))
	}

	return sb.String()
}

func compareSlices(a, b []interface{}, prefix string) string {
	var sb strings.Builder

	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
		sb.WriteString(color.RedString("%sSlice length mismatch: %d vs %d\n", prefix, len(a), len(b)))
	}

	for i := 0; i < minLen; i++ {
		newPrefix := fmt.Sprintf("%s[%d]", prefix, i)
		sb.WriteString(compareRecursive(a[i], b[i], newPrefix))
	}

	return sb.String()
}

func formatDiff(diff string, quiet bool) string {
	if diff == "" {
		if quiet {
			return ""
		}
		return color.GreenString("No differences found.\n")
	}

	if quiet {
		return ""
	}

	re := regexp.MustCompile(`(?m)^`)
	lines := re.Split(strings.TrimSpace(diff), -1)
	count := len(lines)

	summary := color.MagentaString("Found %d differences\n", count)
	return summary + diff
}
