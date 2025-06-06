package parser

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4/device"
	"github.com/aws/aws-sdk-go-v2/aws/stscreds"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/dolastack/structdiff/pkg/cache"
	"github.com/go-ini/ini"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type ParsedData struct {
	Type string
	Data interface{}
}

var (
	auth struct {
		BasicUser        string
		BasicPassword    string
		BearerToken      string
		OAuthConfig      *device.Config
		DeviceConfig     *device.Config
		TokenSource      oauth2.TokenSource
		AwsSigv4Enabled  bool
		AwsService       string
		AwsRegion        string
		AwsCredentials   aws.CredentialsProvider
		AwsAssumeRoleARN string
		CacheToken       bool
		TokenCache       *cache.TokenCache
		Headers          map[string]string
	}
	mu sync.Mutex
)

func ParseFile(path string) (interface{}, error) {
	var reader io.ReadCloser
	var err error

	switch path {
	case "-":
		content, _ := io.ReadAll(os.Stdin)
		return parseContent(path, content)
	default:
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			reader, err = fetchRemoteFile(path)
		} else {
			reader, err = os.Open(path)
		}
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		content, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		return parseContent(path, content)
	}
}

func parseContent(path string, content []byte) (interface{}, error) {
	ext := detectExtension(path, content)

	switch ext {
	case ".json":
		var data interface{}
		err := json.Unmarshal(content, &data)
		return data, err
	case ".yaml", ".yml":
		var data interface{}
		err := yaml.Unmarshal(content, &data)
		return data, err
	case ".toml":
		var data interface{}
		err := toml.Unmarshal(content, &data)
		return data, err
	case ".xml":
		var data interface{}
		err := xml.Unmarshal(content, &data)
		return data, err
	case ".ini":
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
	case ".csv":
		r := csv.NewReader(bytes.NewReader(content))
		records, err := r.ReadAll()
		return records, err
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}
