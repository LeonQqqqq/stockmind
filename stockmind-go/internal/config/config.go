package config

import (
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Claude      ClaudeConfig      `yaml:"claude"`
	DataService DataServiceConfig `yaml:"data_service"`
	Server      ServerConfig      `yaml:"server"`
	SQLite      SQLiteConfig      `yaml:"sqlite"`
}

type ClaudeConfig struct {
	APIKey    string `yaml:"api_key"`
	BaseURL   string `yaml:"base_url"`
	Model     string `yaml:"model"`
	MaxTokens int    `yaml:"max_tokens"`
}

type DataServiceConfig struct {
	BaseURL string `yaml:"base_url"`
	Timeout int    `yaml:"timeout"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type SQLiteConfig struct {
	Path string `yaml:"path"`
}

var envRe = regexp.MustCompile(`\$\{(\w+)\}`)

func expandEnv(s string) string {
	return envRe.ReplaceAllStringFunc(s, func(m string) string {
		key := strings.TrimSuffix(strings.TrimPrefix(m, "${"), "}")
		if v := os.Getenv(key); v != "" {
			return v
		}
		return m
	})
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	expanded := expandEnv(string(data))
	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
