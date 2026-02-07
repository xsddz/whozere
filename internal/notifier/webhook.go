package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xsddz/whozere/internal/config"
)

// Webhook implements generic webhook notifications
type Webhook struct {
	name        string
	url         string
	method      string
	contentType string
	client      *http.Client
}

// NewWebhook creates a new Webhook notifier
func NewWebhook(cfg config.NotifierConfig) (*Webhook, error) {
	url := cfg.Config["url"]
	if url == "" {
		return nil, fmt.Errorf("webhook: url is required")
	}

	method := strings.ToUpper(cfg.Config["method"])
	if method == "" {
		method = "POST"
	}
	if method != "GET" && method != "POST" {
		return nil, fmt.Errorf("webhook: method must be GET or POST")
	}

	contentType := cfg.Config["content_type"]
	if contentType == "" {
		contentType = "application/json"
	}

	name := cfg.Name
	if name == "" {
		name = "Webhook"
	}

	return &Webhook{
		name:        name,
		url:         url,
		method:      method,
		contentType: contentType,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Name returns the notifier name
func (w *Webhook) Name() string {
	return w.name
}

// Send sends a webhook notification
func (w *Webhook) Send(event LoginEvent) error {
	payload := map[string]interface{}{
		"event":     "login",
		"username":  event.Username,
		"hostname":  event.Hostname,
		"ip":        event.IP,
		"terminal":  event.Terminal,
		"timestamp": event.Timestamp.Format(time.RFC3339),
		"os":        event.OS,
		"message":   event.Format(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: failed to marshal payload: %w", err)
	}

	var req *http.Request
	if w.method == "GET" {
		req, err = http.NewRequest("GET", w.url, nil)
	} else {
		req, err = http.NewRequest("POST", w.url, bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", w.contentType)
	}
	if err != nil {
		return fmt.Errorf("webhook: failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "whozere/1.0")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
