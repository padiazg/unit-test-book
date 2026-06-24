package temp_files_and_parsing

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Debug   bool   `json:"debug"`
	Timeout int    `json:"timeout"`
}

func DefaultConfig() Config {
	return Config{
		Host:    "localhost",
		Port:    8080,
		Debug:   false,
		Timeout: 30,
	}
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("config file is empty")
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

func LoadConfigWithDefaults(path string) (*Config, error) {
	if path == "" {
		cfg := DefaultConfig()
		return &cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			return &cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30
	}

	return &cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
