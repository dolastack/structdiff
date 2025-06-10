package compare

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RemoteConfig struct {
	Timeout      time.Duration
	MaxFileSize  int64
	Username     string
	Password     string
	Token        string
	SkipValidate bool
}

var DefaultRemoteConfig = RemoteConfig{
	Timeout:     30 * time.Second,
	MaxFileSize: 10 * 1024 * 1024,
}

type Comparator interface {
	Compare(file1, file2 string, ignoreCase bool, config RemoteConfig) ([]Diff, error)
	Validator() FileValidator
}

func readFileContent(source string, config RemoteConfig) ([]byte, error) {
	if source == "-" {
		return io.ReadAll(os.Stdin)
	}
	if strings.HasPrefix(source, "https://") {
		client := &http.Client{
			Timeout: config.Timeout,
		}

		req, err := http.NewRequest("GET", source, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		if config.Username != "" && config.Password != "" {
			req.SetBasicAuth(config.Username, config.Password)
		} else if config.Token != "" {
			req.Header.Add("Authorization", "Bearer "+config.Token)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch remote file: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("remote server returned status: %d", resp.StatusCode)
		}

		if config.MaxFileSize > 0 && resp.ContentLength > config.MaxFileSize {
			return nil, fmt.Errorf("file size %d exceeds limit of %d bytes",
				resp.ContentLength, config.MaxFileSize)
		}

		reader := io.LimitReader(resp.Body, config.MaxFileSize)
		content, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("error reading remote content: %v", err)
		}

		if config.MaxFileSize > 0 && int64(len(content)) == config.MaxFileSize {
			_, err := io.CopyN(io.Discard, resp.Body, 1)
			if err == nil {
				return nil, fmt.Errorf("file size exceeds limit of %d bytes", config.MaxFileSize)
			}
		}

		return content, nil
	}

	if config.MaxFileSize > 0 {
		fileInfo, err := os.Stat(source)
		if err != nil {
			return nil, err
		}
		if fileInfo.Size() > config.MaxFileSize {
			return nil, fmt.Errorf("file size %d exceeds limit of %d bytes",
				fileInfo.Size(), config.MaxFileSize)
		}
	}

	return os.ReadFile(source)
}

func DetectFormat(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return "json", nil
	case ".yaml", ".yml":
		return "yaml", nil
	case ".toml":
		return "toml", nil
	case ".xml":
		return "xml", nil
	case ".ini", ".cfg":
		return "ini", nil
	case ".csv":
		return "csv", nil
	case ".hcl":
		return "hcl", nil
	case ".hcl.json", ".json.hcl":
		return "hcljson", nil
	default:
		return "", fmt.Errorf("unsupported file format: %s", ext)
	}
}

func CompareFiles(file1, file2, format string, ignoreCase bool, config RemoteConfig) ([]Diff, error) {
	var comparator Comparator

	switch strings.ToLower(format) {
	case "json":
		comparator = &JSONComparator{}
	case "yaml":
		comparator = &YAMLComparator{}
	case "toml":
		comparator = &TOMLComparator{}
	case "xml":
		comparator = &XMLComparator{}
	case "ini":
		comparator = &INIComparator{}
	case "csv":
		comparator = &CSVComparator{}
	case "hcl":
		comparator = &HCLComparator{}
	case "hcljson":
		comparator = &HCLJSONComparator{}
	default:
		return nil, errors.New("unsupported format")
	}

	if !config.SkipValidate && comparator.Validator() != nil {
		data1, err := readFileContent(file1, config)
		if err != nil {
			return nil, fmt.Errorf("error reading first file: %w", err)
		}
		if err := comparator.Validator().Validate(data1); err != nil {
			return nil, fmt.Errorf("first file validation failed: %w\n%s",
				err, comparator.Validator().ValidationHelp())
		}

		data2, err := readFileContent(file2, config)
		if err != nil {
			return nil, fmt.Errorf("error reading second file: %w", err)
		}
		if err := comparator.Validator().Validate(data2); err != nil {
			return nil, fmt.Errorf("second file validation failed: %w\n%s",
				err, comparator.Validator().ValidationHelp())
		}
	}

	return comparator.Compare(file1, file2, ignoreCase, config)
}
