package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/xsddz/whozere/internal/config"
)

// Slack implements Slack webhook notifications
type Slack struct {
	name    string
	webhook string
	client  *http.Client
}

// NewSlack creates a new Slack notifier
func NewSlack(cfg config.NotifierConfig) (*Slack, error) {
	webhook := cfg.Config["webhook"]
	if webhook == "" {
		return nil, fmt.Errorf("slack: webhook is required")
	}

	name := cfg.Name
	if name == "" {
		name = "Slack"
	}

	return &Slack{
		name:    name,
		webhook: webhook,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Name returns the notifier name
func (s *Slack) Name() string {
	return s.name
}

// Send sends a Slack notification
func (s *Slack) Send(event LoginEvent) error {
	payload := map[string]interface{}{
		"text": event.Format(),
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{
					"type": "plain_text",
					"text": "🔔 Login Alert",
				},
			},
			{
				"type": "section",
				"fields": []map[string]string{
					{"type": "mrkdwn", "text": fmt.Sprintf("*User:*\n%s", event.Username)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Host:*\n%s", event.Hostname)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Time:*\n%s", event.Timestamp.Format("2006-01-02 15:04:05"))},
					{"type": "mrkdwn", "text": fmt.Sprintf("*OS:*\n%s", event.OS)},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", s.webhook, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("slack: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
