package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xsddz/whozere/internal/config"
)

func TestLoginEventFormat(t *testing.T) {
	event := LoginEvent{
		Username:  "testuser",
		Hostname:  "testhost",
		IP:        "192.168.1.1",
		Terminal:  "pts/0",
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		OS:        "linux",
	}

	msg := event.Format()

	if msg == "" {
		t.Error("Format() returned empty string")
	}

	// Check if key information is included
	if !contains(msg, "testuser") {
		t.Error("Format() should contain username")
	}
	if !contains(msg, "testhost") {
		t.Error("Format() should contain hostname")
	}
	if !contains(msg, "192.168.1.1") {
		t.Error("Format() should contain IP")
	}
}

func TestWebhookNotifier(t *testing.T) {
	// Create a test server
	var receivedPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json, got %s", r.Header.Get("Content-Type"))
		}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&receivedPayload); err != nil {
			t.Errorf("Failed to decode payload: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.NotifierConfig{
		Type:    "webhook",
		Name:    "Test Webhook",
		Enabled: true,
		Config: map[string]string{
			"url": server.URL,
		},
	}

	webhook, err := NewWebhook(cfg)
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}

	event := LoginEvent{
		Username:  "testuser",
		Hostname:  "testhost",
		Timestamp: time.Now(),
		OS:        "darwin",
	}

	if err := webhook.Send(event); err != nil {
		t.Errorf("Send() failed: %v", err)
	}

	if receivedPayload["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got '%v'", receivedPayload["username"])
	}
}

func TestNewNotifier(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.NotifierConfig
		wantErr bool
	}{
		{
			name: "webhook",
			cfg: config.NotifierConfig{
				Type:   "webhook",
				Config: map[string]string{"url": "http://example.com"},
			},
			wantErr: false,
		},
		{
			name: "webhook without url",
			cfg: config.NotifierConfig{
				Type:   "webhook",
				Config: map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "unknown type",
			cfg: config.NotifierConfig{
				Type: "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
