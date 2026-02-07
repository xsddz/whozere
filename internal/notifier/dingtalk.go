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
	"net/url"
	"time"

	"github.com/xsddz/whozere/internal/config"
)

// DingTalk implements DingTalk robot notifications
type DingTalk struct {
	name    string
	webhook string
	secret  string
	client  *http.Client
}

// NewDingTalk creates a new DingTalk notifier
func NewDingTalk(cfg config.NotifierConfig) (*DingTalk, error) {
	webhook := cfg.Config["webhook"]
	if webhook == "" {
		return nil, fmt.Errorf("dingtalk: webhook is required")
	}

	name := cfg.Name
	if name == "" {
		name = "DingTalk"
	}

	return &DingTalk{
		name:    name,
		webhook: webhook,
		secret:  cfg.Config["secret"],
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Name returns the notifier name
func (d *DingTalk) Name() string {
	return d.name
}

// Send sends a DingTalk notification
func (d *DingTalk) Send(event LoginEvent) error {
	webhookURL := d.webhook

	// Add signature if secret is configured
	if d.secret != "" {
		timestamp := time.Now().UnixMilli()
		sign := d.sign(timestamp)
		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, url.QueryEscape(sign))
	}

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": event.Format(),
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("dingtalk: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("dingtalk: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("dingtalk: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("dingtalk: failed to parse response: %w", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("dingtalk: error %d: %s", result.ErrCode, result.ErrMsg)
	}

	return nil
}

func (d *DingTalk) sign(timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, d.secret)
	h := hmac.New(sha256.New, []byte(d.secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
