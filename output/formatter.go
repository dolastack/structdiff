package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dolastack/structdiff/compare"
	"github.com/fatih/color"
)

func FormatText(diffs []compare.Diff, useColor bool) (string, error) {
	var sb strings.Builder

	if useColor {
		color.NoColor = false
	} else {
		color.NoColor = true
	}

	summary := generateSummary(diffs)
	sb.WriteString(summary + "\n\n")

	added := color.New(color.FgGreen).SprintFunc()
	removed := color.New(color.FgRed).SprintFunc()
	modified := color.New(color.FgYellow).SprintFunc()
	moved := color.New(color.FgBlue).SprintFunc()

	for _, diff := range diffs {
		switch diff.Type {
		case compare.DiffAdded:
			sb.WriteString(fmt.Sprintf("%s %s: %v\n", added("+"), diff.Path, diff.NewValue))
		case compare.DiffRemoved:
			sb.WriteString(fmt.Sprintf("%s %s: %v\n", removed("-"), diff.Path, diff.OldValue))
		case compare.DiffModified:
			sb.WriteString(fmt.Sprintf("%s %s: %v → %v\n", modified("~"), diff.Path, diff.OldValue, diff.NewValue))
		case compare.DiffMoved:
			sb.WriteString(fmt.Sprintf("%s %s moved\n", moved(">"), diff.Path))
		}
	}
	return sb.String(), nil
}

func FormatJSON(diffs []compare.Diff) (string, error) {
	output := struct {
		Summary struct {
			Total    int `json:"total"`
			Added    int `json:"added"`
			Removed  int `json:"removed"`
			Modified int `json:"modified"`
			Moved    int `json:"moved"`
		} `json:"summary"`
		Diffs []compare.Diff `json:"diffs"`
	}{
		Summary: generateSummaryStruct(diffs),
		Diffs:   diffs,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func generateSummary(diffs []compare.Diff) string {
	summary := generateSummaryStruct(diffs)

	var parts []string
	if summary.Added > 0 {
		parts = append(parts, fmt.Sprintf("%d added", summary.Added))
	}
	if summary.Removed > 0 {
		parts = append(parts, fmt.Sprintf("%d removed", summary.Removed))
	}
	if summary.Modified > 0 {
		parts = append(parts, fmt.Sprintf("%d modified", summary.Modified))
	}
	if summary.Moved > 0 {
		parts = append(parts, fmt.Sprintf("%d moved", summary.Moved))
	}

	return fmt.Sprintf("Found %d differences (%s)", summary.Total, strings.Join(parts, ", "))
}

func generateSummaryStruct(diffs []compare.Diff) struct {
	Total    int `json:"total"`
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Modified int `json:"modified"`
	Moved    int `json:"moved"`
} {
	var added, removed, modified, moved int
	for _, diff := range diffs {
		switch diff.Type {
		case compare.DiffAdded:
			added++
		case compare.DiffRemoved:
			removed++
		case compare.DiffModified:
			modified++
		case compare.DiffMoved:
			moved++
		}
	}

	return struct {
		Total    int `json:"total"`
		Added    int `json:"added"`
		Removed  int `json:"removed"`
		Modified int `json:"modified"`
		Moved    int `json:"moved"`
	}{
		Total:    len(diffs),
		Added:    added,
		Removed:  removed,
		Modified: modified,
		Moved:    moved,
	}
}
