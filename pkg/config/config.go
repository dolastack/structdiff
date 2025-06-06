package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type OAuthConfig struct {
	ClientID     string `yaml:"client-id"`
	ClientSecret string `yaml:"client-secret"`
	TokenURL     string `yaml:"token-url"`
}

type AwsSigv4 struct {
	Service string `yaml:"service"`
	Region  string `yaml:"region"`
	Role    string `yaml:"role"`
}

type Config struct {
	BasicUser     string      `yaml:"basic-username"`
	BasicPassword string      `yaml:"basic-password"`
	BearerToken   string      `yaml:"bearer-token"`
	OAuth         OAuthConfig `yaml:"oauth"`
	AwsSigv4      AwsSigv4    `yaml:"aws-sigv4"`
	Filters       []string    `yaml:"filter"`
	Color         string      `yaml:"color"`
	Quiet         bool        `yaml:"quiet"`
	Check         bool        `yaml:"check"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	cfg.BasicPassword = resolveValue(cfg.BasicPassword)
	cfg.BearerToken = resolveValue(cfg.BearerToken)

	return &cfg, nil
}

func resolveValue(val string) string {
	if val == "" {
		return ""
	}

	switch {
	case strings.HasPrefix(val, "$"):
		return os.Getenv(strings.TrimPrefix(val, "$"))
	case strings.HasPrefix(val, "@"):
		path := strings.TrimPrefix(val, "@")
		content, _ := os.ReadFile(path)
		return strings.TrimSpace(string(content))
	default:
		return val
	}
}
