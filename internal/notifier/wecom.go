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

// WeCom implements WeCom (企业微信) robot notifications
type WeCom struct {
	name    string
	webhook string
	client  *http.Client
}

// NewWeCom creates a new WeCom notifier
func NewWeCom(cfg config.NotifierConfig) (*WeCom, error) {
	webhook := cfg.Config["webhook"]
	if webhook == "" {
		return nil, fmt.Errorf("wecom: webhook is required")
	}

	name := cfg.Name
	if name == "" {
		name = "WeCom"
	}

	return &WeCom{
		name:    name,
		webhook: webhook,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Name returns the notifier name
func (w *WeCom) Name() string {
	return w.name
}

// Send sends a WeCom notification
func (w *WeCom) Send(event LoginEvent) error {
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": event.Format(),
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("wecom: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", w.webhook, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("wecom: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("wecom: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("wecom: failed to parse response: %w", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("wecom: error %d: %s", result.ErrCode, result.ErrMsg)
	}

	return nil
}
