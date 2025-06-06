package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type TokenCache struct {
	TokenPath string
}

func NewTokenCache(path string) *TokenCache {
	if path == "" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, ".structdiff", "token.json")
	}
	return &TokenCache{TokenPath: path}
}

func (c *TokenCache) SaveToken(token map[string]interface{}) error {
	dir := filepath.Dir(c.TokenPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create token dir: %v", err)
	}

	file, err := os.OpenFile(c.TokenPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open token file: %v", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(token); err != nil {
		return fmt.Errorf("failed to write token: %v", err)
	}

	return nil
}

func (c *TokenCache) LoadToken() (map[string]interface{}, error) {
	data, err := os.ReadFile(c.TokenPath)
	if err != nil {
		return nil, err
	}

	var token map[string]interface{}
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return token, nil
}

func (c *TokenCache) ClearToken() error {
	return os.Remove(c.TokenPath)
}
