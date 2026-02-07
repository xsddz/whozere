package notifier

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/xsddz/whozere/internal/config"
)

// Feishu implements Feishu (Lark) robot notifications
type Feishu struct {
	name    string
	webhook string
	secret  string
	client  *http.Client
}

// NewFeishu creates a new Feishu notifier
func NewFeishu(cfg config.NotifierConfig) (*Feishu, error) {
	webhook := cfg.Config["webhook"]
	if webhook == "" {
		return nil, fmt.Errorf("feishu: webhook is required")
	}

	name := cfg.Name
	if name == "" {
		name = "Feishu"
	}

	return &Feishu{
		name:    name,
		webhook: webhook,
		secret:  cfg.Config["secret"],
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Name returns the notifier name
func (f *Feishu) Name() string {
	return f.name
}

// sign generates the signature for Feishu webhook
func (f *Feishu) sign(timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, f.secret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Send sends a Feishu notification
func (f *Feishu) Send(event LoginEvent) error {
	timestamp := time.Now().Unix()

	// Build message payload
	payload := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": event.Format(),
		},
	}

	// Add signature if secret is configured
	if f.secret != "" {
		payload["timestamp"] = fmt.Sprintf("%d", timestamp)
		payload["sign"] = f.sign(timestamp)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("feishu: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", f.webhook, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("feishu: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("feishu: failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("feishu: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// Check response
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		if result.Code != 0 {
			return fmt.Errorf("feishu: API error: %s", result.Msg)
		}
	}

	return nil
}
