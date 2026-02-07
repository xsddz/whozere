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

// Telegram implements Telegram bot notifications
type Telegram struct {
	name   string
	token  string
	chatID string
	client *http.Client
}

// NewTelegram creates a new Telegram notifier
func NewTelegram(cfg config.NotifierConfig) (*Telegram, error) {
	token := cfg.Config["token"]
	if token == "" {
		return nil, fmt.Errorf("telegram: token is required")
	}

	chatID := cfg.Config["chat_id"]
	if chatID == "" {
		return nil, fmt.Errorf("telegram: chat_id is required")
	}

	name := cfg.Name
	if name == "" {
		name = "Telegram"
	}

	return &Telegram{
		name:   name,
		token:  token,
		chatID: chatID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Name returns the notifier name
func (t *Telegram) Name() string {
	return t.name
}

// Send sends a Telegram notification
func (t *Telegram) Send(event LoginEvent) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)

	payload := map[string]interface{}{
		"chat_id":    t.chatID,
		"text":       event.Format(),
		"parse_mode": "HTML",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("telegram: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("telegram: failed to parse response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("telegram: %s", result.Description)
	}

	return nil
}
