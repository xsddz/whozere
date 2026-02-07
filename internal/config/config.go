package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration
type Config struct {
	Notifiers []NotifierConfig `yaml:"notifiers"`
}

// NotifierConfig represents a notification channel configuration
type NotifierConfig struct {
	Type    string            `yaml:"type"`    // webhook, dingtalk, wecom, telegram, slack, email
	Name    string            `yaml:"name"`    // optional friendly name
	Enabled bool              `yaml:"enabled"` // enable/disable this notifier
	Config  map[string]string `yaml:"config"`  // type-specific configuration
}

// Load reads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Notifiers) == 0 {
		return fmt.Errorf("at least one notifier must be configured")
	}

	hasEnabled := false
	for i, n := range c.Notifiers {
		if n.Type == "" {
			return fmt.Errorf("notifier[%d]: type is required", i)
		}
		if n.Enabled {
			hasEnabled = true
		}
	}

	if !hasEnabled {
		return fmt.Errorf("at least one notifier must be enabled")
	}

	return nil
}
