package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration
type Config struct {
	Notifiers []NotifierConfig `yaml:"notifiers"`
	Filters   FilterConfig     `yaml:"filters"`
}

// FilterConfig defines event filtering rules
type FilterConfig struct {
	// IgnoreTerminals is a list of terminal types to ignore (e.g., cron, su, sudo)
	IgnoreTerminals []string `yaml:"ignore_terminals"`
	// IgnoreUsers is a list of users to ignore
	IgnoreUsers []string `yaml:"ignore_users"`
	// IgnoreCombinations is a list of user+terminal combinations to ignore
	IgnoreCombinations []FilterCombination `yaml:"ignore_combinations"`
}

// FilterCombination defines a specific user+terminal combination to ignore
type FilterCombination struct {
	User     string `yaml:"user"`
	Terminal string `yaml:"terminal"`
}

// ShouldIgnore checks if an event should be ignored based on filter rules
func (f *FilterConfig) ShouldIgnore(username, terminal string) bool {
	// Check ignore_terminals
	for _, t := range f.IgnoreTerminals {
		if t == terminal {
			return true
		}
	}

	// Check ignore_users
	for _, u := range f.IgnoreUsers {
		if u == username {
			return true
		}
	}

	// Check ignore_combinations
	for _, c := range f.IgnoreCombinations {
		if c.User == username && c.Terminal == terminal {
			return true
		}
	}

	return false
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
