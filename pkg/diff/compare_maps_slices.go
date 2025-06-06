// pkg/diff/compare_maps_slices.go

package diff

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// compareMaps compares two maps recursively
func compareMaps(a, b map[string]interface{}, prefix string, filterPaths []string, visited map[uintptr]bool) {
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

		// Fuzzy match keys (case-insensitive)
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

		newPrefix := k
		if prefix != "" {
			newPrefix = prefix + "." + k
		}

		if foundInA && foundInB {
			compareRecursive(a[keyA], b[keyB], newPrefix, filterPaths, visited)
		} else if !foundInA {
			recordDiff(newPrefix, color.RedString("%sKey '%s' missing in first file\n", prefix, k))
		} else if !foundInB {
			recordDiff(newPrefix, color.RedString("%sKey '%s' missing in second file\n", prefix, k))
		}
	}
}

// compareSlices compares slices element-wise
func compareSlices(a, b []interface{}, prefix string, filterPaths []string, visited map[uintptr]bool) {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
		recordDiff(prefix, color.RedString("%sSlice length mismatch: %d vs %d\n", prefix, len(a), len(b)))
	}

	for i := 0; i < minLen; i++ {
		newPrefix := fmt.Sprintf("%s[%d]", prefix, i)
		compareRecursive(a[i], b[i], newPrefix, filterPaths, visited)
	}
}
