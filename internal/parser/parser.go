// internal/parser/parser.go

package parser

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-ini/ini"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// ParseOptions holds authentication and request options for remote files
type ParseOptions struct {
	BasicUser     string
	BasicPassword string
	BearerToken   string
	CustomHeaders map[string]string
}

// ParseFile parses a file or URL using provided options
func ParseFile(path string, opts ParseOptions) (interface{}, error) {
	var data interface{}
	var err error

	switch path {
	case "-":
		data, err = parseStdin()
	default:
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			data, err = parseRemoteFile(path, opts)
		} else {
			data, err = parseLocalFile(path)
		}
	}

	return data, err
}

// parseStdin reads input from standard input
func parseStdin() (interface{}, error) {
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %v", err)
	}
	return parseContent(content)
}

// parseLocalFile reads and parses a local file
func parseLocalFile(path string) (interface{}, error) {
	ext := filepath.Ext(path)
	if ext == "" {
		return nil, fmt.Errorf("unable to detect file type")
	}
	ext = strings.ToLower(ext)

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()

	switch ext {
	case ".json":
		var data interface{}
		err := json.NewDecoder(f).Decode(&data)
		return data, err
	case ".yaml", ".yml":
		var data interface{}
		err := yaml.NewDecoder(f).Decode(&data)
		return data, err
	case ".toml":
		b, _ := io.ReadAll(f)
		var data interface{}
		err := toml.Unmarshal(b, &data)
		return data, err
	case ".xml":
		b, _ := io.ReadAll(f)
		var data interface{}
		err := xml.Unmarshal(b, &data)
		return data, err
	case ".ini":
		cfg, err := ini.Load(path)
		if err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for _, section := range cfg.Sections() {
			secMap := make(map[string]string)
			for _, key := range section.Keys() {
				secMap[key.Name()] = key.String()
			}
			m[section.Name()] = secMap
		}
		return m, nil
	case ".csv":
		r := csv.NewReader(f)
		records, err := r.ReadAll()
		return records, err
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// parseRemoteFile fetches and parses a remote file with optional auth and headers
func parseRemoteFile(url string, opts ParseOptions) (interface{}, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set Basic Auth if provided
	if opts.BasicUser != "" || opts.BasicPassword != "" {
		req.SetBasicAuth(opts.BasicUser, opts.BasicPassword)
	}

	// Set Bearer Token if provided
	if opts.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+opts.BearerToken)
	}

	// Set custom headers
	for k, v := range opts.CustomHeaders {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return parseContent(body)
}

// parseContent detects format and unmarshals structured content
func parseContent(content []byte) (interface{}, error) {
	content = bytes.TrimSpace(content)
	if len(content) == 0 {
		return nil, fmt.Errorf("empty content")
	}

	// Try to detect format
	if isJSON(content) {
		var data interface{}
		err := json.Unmarshal(content, &data)
		return data, err
	}

	if isYAML(content) {
		var data interface{}
		err := yaml.Unmarshal(content, &data)
		return data, err
	}

	if isTOML(content) {
		var data interface{}
		err := toml.Unmarshal(content, &data)
		return data, err
	}

	if isXML(content) {
		var data interface{}
		err := xml.Unmarshal(content, &data)
		return data, err
	}

	if isINI(content) {
		cfg, err := ini.LoadSources(ini.LoadOptions{AllowBooleanKeys: true}, bytes.NewReader(content))
		if err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for _, section := range cfg.Sections() {
			secMap := make(map[string]string)
			for _, key := range section.Keys() {
				secMap[key.Name()] = key.String()
			}
			m[section.Name()] = secMap
		}
		return m, nil
	}

	if isCSV(content) {
		r := csv.NewReader(bytes.NewReader(content))
		records, err := r.ReadAll()
		return records, err
	}

	return nil, fmt.Errorf("unsupported content type")
}

// Utility functions to detect format by content

func isJSON(data []byte) bool {
	return len(data) > 0 && data[0] == '{'
}

func isYAML(data []byte) bool {
	return bytes.Contains(data, []byte(":")) && !bytes.Contains(data, []byte("<?xml"))
}

func isTOML(data []byte) bool {
	return bytes.Contains(data, []byte("=")) && bytes.Contains(data, []byte("["))
}

func isXML(data []byte) bool {
	return bytes.HasPrefix(bytes.TrimSpace(data), []byte("<?xml"))
}

func isINI(data []byte) bool {
	return bytes.Contains(data, []byte("=")) && bytes.Contains(data, []byte("["))
}

func isCSV(data []byte) bool {
	r := csv.NewReader(bytes.NewReader(data))
	_, err := r.Read()
	return err == nil
}
