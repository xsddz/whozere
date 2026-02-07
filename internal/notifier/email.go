package notifier

import (
	"fmt"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/xsddz/whozere/internal/config"
)

// Email implements email notifications via SMTP
type Email struct {
	name     string
	host     string
	port     int
	username string
	password string
	from     string
	to       []string
}

// NewEmail creates a new Email notifier
func NewEmail(cfg config.NotifierConfig) (*Email, error) {
	host := cfg.Config["smtp_host"]
	if host == "" {
		return nil, fmt.Errorf("email: smtp_host is required")
	}

	portStr := cfg.Config["smtp_port"]
	if portStr == "" {
		portStr = "587"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("email: invalid smtp_port: %w", err)
	}

	username := cfg.Config["username"]
	password := cfg.Config["password"]

	from := cfg.Config["from"]
	if from == "" {
		from = username
	}

	toStr := cfg.Config["to"]
	if toStr == "" {
		return nil, fmt.Errorf("email: to is required")
	}
	to := strings.Split(toStr, ",")
	for i := range to {
		to[i] = strings.TrimSpace(to[i])
	}

	name := cfg.Name
	if name == "" {
		name = "Email"
	}

	return &Email{
		name:     name,
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		to:       to,
	}, nil
}

// Name returns the notifier name
func (e *Email) Name() string {
	return e.name
}

// Send sends an email notification
func (e *Email) Send(event LoginEvent) error {
	subject := fmt.Sprintf("Login Alert: %s logged in to %s", event.Username, event.Hostname)

	body := fmt.Sprintf(`Login detected on your system:

User: %s
Hostname: %s
Time: %s
OS: %s`,
		event.Username,
		event.Hostname,
		event.Timestamp.Format("2006-01-02 15:04:05 MST"),
		event.OS,
	)

	if event.IP != "" {
		body += fmt.Sprintf("\nIP Address: %s", event.IP)
	}
	if event.Terminal != "" {
		body += fmt.Sprintf("\nTerminal: %s", event.Terminal)
	}

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s",
		e.from,
		strings.Join(e.to, ", "),
		subject,
		body,
	)

	addr := fmt.Sprintf("%s:%d", e.host, e.port)

	var auth smtp.Auth
	if e.username != "" && e.password != "" {
		auth = smtp.PlainAuth("", e.username, e.password, e.host)
	}

	if err := smtp.SendMail(addr, auth, e.from, e.to, []byte(msg)); err != nil {
		return fmt.Errorf("email: failed to send: %w", err)
	}

	return nil
}
