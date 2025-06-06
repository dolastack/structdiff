// pkg/diff/diff.go

package diff

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/fatih/color"
)

var (
	diffLines = make(map[string]string)
	mu        sync.Mutex
)

// Compare compares two structured values and returns human-readable diff
func Compare(a, b interface{}, filterPaths []string) string {
	mu.Lock()
	defer mu.Unlock()

	diffLines = make(map[string]string)
	visited := make(map[uintptr]bool)

	compareRecursive(a, b, "", filterPaths, visited)

	if len(diffLines) == 0 {
		return color.GreenString("No differences found.\n")
	}

	var sb strings.Builder
	sb.WriteString(color.MagentaString("Found %d differences\n", len(diffLines)))

	for _, line := range diffLines {
		sb.WriteString(line)
	}

	return sb.String()
}

// compareRecursive walks through maps/slices and records mismatches
func compareRecursive(a, b interface{}, prefix string, filterPaths []string, visited map[uintptr]bool) {
	if shouldSkip(prefix, filterPaths) {
		return
	}

	ta := reflect.TypeOf(a)
	tb := reflect.TypeOf(b)
	if ta != tb {
		recordDiff(prefix, color.RedString("%sType mismatch: %v vs %v\n", prefix, ta, tb))
		return
	}

	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	// Dereference pointers
	if valA.Kind() == reflect.Ptr && !valA.IsNil() {
		valA = valA.Elem()
	}
	if valB.Kind() == reflect.Ptr && !valB.IsNil() {
		valB = valB.Elem()
	}

	aInterface := valA.Interface()
	bInterface := valB.Interface()

	addrA := reflect.ValueOf(aInterface).Pointer()
	if isVisited(addrA, visited) {
		return
	}
	markVisited(addrA, visited)

	addrB := reflect.ValueOf(bInterface).Pointer()
	markVisited(addrB, visited)

	switch aTyped := aInterface.(type) {
	case map[string]interface{}:
		bTyped, ok := bInterface.(map[string]interface{})
		if !ok {
			recordDiff(prefix, color.RedString("%sType mismatch: expected map\n", prefix))
			return
		}
		compareMaps(aTyped, bTyped, prefix, filterPaths, visited)

	case []interface{}:
		bTyped, ok := bInterface.([]interface{})
		if !ok {
			recordDiff(prefix, color.RedString("%sType mismatch: expected slice\n", prefix))
			return
		}
		compareSlices(aTyped, bTyped, prefix, filterPaths, visited)

	default:
		if fmt.Sprintf("%v", aInterface) != fmt.Sprintf("%v", bInterface) {
			recordDiff(prefix, color.YellowString("%sValue mismatch: '%v' vs '%v'\n", prefix, aInterface, bInterface))
		}
	}
}

// recordDiff adds a message to diffLines only once per path
func recordDiff(path, message string) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := diffLines[path]; !exists {
		diffLines[path] = message
	}
}

// isVisited checks if a pointer was already processed
func isVisited(ptr uintptr, visited map[uintptr]bool) bool {
	return visited[ptr]
}

// markVisited marks a pointer as visited
func markVisited(ptr uintptr, visited map[uintptr]bool) {
	visited[ptr] = true
}

// shouldSkip returns true if current path doesn't match any filters
func shouldSkip(path string, filterPaths []string) bool {
	if len(filterPaths) == 0 || path == "" {
		return false
	}

	matched := false
	for _, fpath := range filterPaths {
		if strings.HasPrefix(path+".", fpath+".") || path == fpath {
			matched = true
			break
		}
	}
	return !matched
}
