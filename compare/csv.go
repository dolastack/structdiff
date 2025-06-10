package compare

import (
	"bytes"
	"encoding/csv"
	"fmt"
)

type CSVValidator struct{}

func (c *CSVValidator) Validate(content []byte) error {
	_, err := csv.NewReader(bytes.NewReader(content)).ReadAll()
	return err
}

func (c *CSVValidator) ValidationHelp() string {
	return `CSV validation tips:
• Ensure consistent column count per row
• Properly escape quotes in quoted fields
• Use consistent line endings
• Check for malformed records`
}

type CSVComparator struct {
	CSVValidator
}

func (c *CSVComparator) Compare(file1, file2 string, ignoreCase bool, config RemoteConfig) ([]Diff, error) {
	data1, err := readFileContent(file1, config)
	if err != nil {
		return nil, fmt.Errorf("error reading first file: %w", err)
	}

	data2, err := readFileContent(file2, config)
	if err != nil {
		return nil, fmt.Errorf("error reading second file: %w", err)
	}

	records1, err := csv.NewReader(bytes.NewReader(data1)).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error parsing first CSV file: %w", err)
	}

	records2, err := csv.NewReader(bytes.NewReader(data2)).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error parsing second CSV file: %w", err)
	}

	obj1 := csvToMaps(records1)
	obj2 := csvToMaps(records2)

	return CompareValues(obj1, obj2, "", ignoreCase), nil
}

func (c *CSVComparator) Validator() FileValidator {
	return &c.CSVValidator
}

func csvToMaps(records [][]string) []map[string]interface{} {
	if len(records) == 0 {
		return nil
	}

	headers := records[0]
	var result []map[string]interface{}

	for _, record := range records[1:] {
		row := make(map[string]interface{})
		for i, value := range record {
			if i < len(headers) {
				row[headers[i]] = value
			}
		}
		result = append(result, row)
	}

	return result
}
