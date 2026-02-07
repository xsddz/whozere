package notifier

import (
	"fmt"
	"time"

	"github.com/xsddz/whozere/internal/config"
)

// LoginEvent represents a login event to be notified
type LoginEvent struct {
	Username  string    // user who logged in
	Hostname  string    // hostname of the machine
	IP        string    // source IP address (if available)
	Terminal  string    // terminal/session type (tty, pts, console, etc.)
	Timestamp time.Time // when the login occurred
	OS        string    // operating system
}

// Format returns a formatted message for the login event
func (e LoginEvent) Format() string {
	// Get timezone info
	zone, offset := e.Timestamp.Zone()
	offsetHours := offset / 3600
	var offsetStr string
	if offsetHours >= 0 {
		offsetStr = fmt.Sprintf("UTC+%d", offsetHours)
	} else {
		offsetStr = fmt.Sprintf("UTC%d", offsetHours)
	}

	msg := fmt.Sprintf("🔔 Login Alert\n\n"+
		"User: %s\n"+
		"Host: %s\n"+
		"Time: %s\n"+
		"Zone: %s (%s)\n"+
		"OS: %s",
		e.Username,
		e.Hostname,
		e.Timestamp.Format("2006-01-02 15:04:05"),
		zone,
		offsetStr,
		e.OS,
	)

	if e.IP != "" {
		msg += fmt.Sprintf("\nIP: %s", e.IP)
	}
	if e.Terminal != "" {
		msg += fmt.Sprintf("\nTerminal: %s", e.Terminal)
	}

	return msg
}

// Notifier is the interface for sending notifications
type Notifier interface {
	// Name returns the notifier name
	Name() string
	// Send sends a notification for a login event
	Send(event LoginEvent) error
}

// New creates a new notifier based on configuration
func New(cfg config.NotifierConfig) (Notifier, error) {
	switch cfg.Type {
	case "webhook":
		return NewWebhook(cfg)
	case "dingtalk":
		return NewDingTalk(cfg)
	case "wecom":
		return NewWeCom(cfg)
	case "telegram":
		return NewTelegram(cfg)
	case "slack":
		return NewSlack(cfg)
	case "email":
		return NewEmail(cfg)
	case "feishu":
		return NewFeishu(cfg)
	default:
		return nil, fmt.Errorf("unknown notifier type: %s", cfg.Type)
	}
}
