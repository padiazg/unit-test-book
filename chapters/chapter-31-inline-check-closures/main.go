package inline_check_closures

import (
	"errors"
	"fmt"
)

type Config struct {
	Host string
	Port int
	Key  string
}

func ValidateConfig(cfg Config) error {
	if cfg.Host == "" {
		return errors.New("host is required")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", cfg.Port)
	}
	if cfg.Key == "" {
		return errors.New("api key is required")
	}
	return nil
}
