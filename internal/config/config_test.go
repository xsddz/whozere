package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	content := `
notifiers:
  - type: webhook
    name: "Test Webhook"
    enabled: true
    config:
      url: "https://example.com/webhook"
`
	tmpfile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Test loading
	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg.Notifiers) != 1 {
		t.Errorf("Expected 1 notifier, got %d", len(cfg.Notifiers))
	}

	if cfg.Notifiers[0].Type != "webhook" {
		t.Errorf("Expected type 'webhook', got '%s'", cfg.Notifiers[0].Type)
	}

	if cfg.Notifiers[0].Config["url"] != "https://example.com/webhook" {
		t.Errorf("Expected url 'https://example.com/webhook', got '%s'", cfg.Notifiers[0].Config["url"])
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "empty notifiers",
			config:  Config{},
			wantErr: true,
		},
		{
			name: "no enabled notifier",
			config: Config{
				Notifiers: []NotifierConfig{
					{Type: "webhook", Enabled: false},
				},
			},
			wantErr: true,
		},
		{
			name: "missing type",
			config: Config{
				Notifiers: []NotifierConfig{
					{Enabled: true},
				},
			},
			wantErr: true,
		},
		{
			name: "valid config",
			config: Config{
				Notifiers: []NotifierConfig{
					{Type: "webhook", Enabled: true},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
