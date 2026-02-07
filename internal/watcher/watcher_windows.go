//go:build windows

package watcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/xsddz/whozere/internal/notifier"
)

// WindowsWatcher watches for login events on Windows
type WindowsWatcher struct {
	hostname string
}

func newPlatformWatcher() (Watcher, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return &WindowsWatcher{hostname: hostname}, nil
}

// Name returns the watcher name
func (w *WindowsWatcher) Name() string {
	return "windows"
}

// Watch monitors Windows Security Event Log for login events (new events only)
func (w *WindowsWatcher) Watch(ctx context.Context, events chan<- notifier.LoginEvent) error {
	return w.WatchWithOptions(ctx, events, Options{})
}

// WatchWithOptions monitors Windows event logs with specific options
func (w *WindowsWatcher) WatchWithOptions(ctx context.Context, events chan<- notifier.LoginEvent, opts Options) error {
	// Pattern to extract username from event
	userPattern := regexp.MustCompile(`Account Name:\s+(\S+)`)
	ipPattern := regexp.MustCompile(`Source Network Address:\s+([\d\.]+)`)
	logonTypePattern := regexp.MustCompile(`Logon Type:\s+(\d+)`)

	processEvent := func(eventData string) *notifier.LoginEvent {
		userMatches := userPattern.FindStringSubmatch(eventData)
		if userMatches == nil {
			return nil
		}

		username := userMatches[1]
		// Skip system accounts
		if strings.HasSuffix(username, "$") || username == "-" || username == "SYSTEM" {
			return nil
		}

		event := &notifier.LoginEvent{
			Username:  username,
			Hostname:  w.hostname,
			Terminal:  "windows",
			Timestamp: time.Now(),
			OS:        "windows",
		}

		// Extract IP if available
		if ipMatches := ipPattern.FindStringSubmatch(eventData); ipMatches != nil {
			if ipMatches[1] != "-" && ipMatches[1] != "" {
				event.IP = ipMatches[1]
			}
		}

		// Extract logon type
		if logonMatches := logonTypePattern.FindStringSubmatch(eventData); logonMatches != nil {
			switch logonMatches[1] {
			case "2":
				event.Terminal = "console"
			case "3":
				event.Terminal = "network"
			case "10":
				event.Terminal = "rdp"
			case "11":
				event.Terminal = "cached"
			}
		}

		return event
	}

	// If since is specified, query historical events
	if opts.Since > 0 {
		minutes := int(opts.Since.Minutes())
		if minutes < 1 {
			minutes = 1
		}

		// Query historical events using wevtutil
		psCmd := fmt.Sprintf(
			`Get-WinEvent -FilterHashtable @{LogName='Security';Id=4624;StartTime=(Get-Date).AddMinutes(-%d)} -ErrorAction SilentlyContinue | ForEach-Object { $_.Message }`,
			minutes,
		)
		cmd := exec.CommandContext(ctx, "powershell", "-Command", psCmd)
		if output, err := cmd.Output(); err == nil {
			eventBlocks := strings.Split(string(output), "\r\n\r\n")
			for _, block := range eventBlocks {
				if event := processEvent(block); event != nil {
					select {
					case events <- *event:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}
		}
	}

	// Now watch for new events using PowerShell event subscription
	// Use wevtutil to subscribe to new events
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastCheck := time.Now()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Query events since last check
			psCmd := fmt.Sprintf(
				`Get-WinEvent -FilterHashtable @{LogName='Security';Id=4624;StartTime='%s'} -ErrorAction SilentlyContinue | ForEach-Object { $_.Message }`,
				lastCheck.Format("2006-01-02T15:04:05"),
			)
			cmd := exec.CommandContext(ctx, "powershell", "-Command", psCmd)
			lastCheck = time.Now()

			output, err := cmd.Output()
			if err != nil {
				continue
			}

			eventBlocks := strings.Split(string(output), "\r\n\r\n")
			for _, block := range eventBlocks {
				if event := processEvent(block); event != nil {
					select {
					case events <- *event:
					case <-ctx.Done():
						return nil
					}
				}
			}
		}
	}
}
