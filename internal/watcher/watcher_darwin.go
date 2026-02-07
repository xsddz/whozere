//go:build darwin

package watcher

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/xsddz/whozere/internal/notifier"
)

// DarwinWatcher watches for login events on macOS
type DarwinWatcher struct {
	hostname string
}

func newPlatformWatcher() (Watcher, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return &DarwinWatcher{hostname: hostname}, nil
}

// Name returns the watcher name
func (w *DarwinWatcher) Name() string {
	return "darwin"
}

// Watch monitors macOS system logs for login events (new events only)
func (w *DarwinWatcher) Watch(ctx context.Context, events chan<- notifier.LoginEvent) error {
	return w.WatchWithOptions(ctx, events, Options{})
}

// WatchWithOptions monitors macOS system logs with specific options
func (w *DarwinWatcher) WatchWithOptions(ctx context.Context, events chan<- notifier.LoginEvent, opts Options) error {
	// Patterns to detect login events
	sshPattern := regexp.MustCompile(`sshd.*Accepted\s+\w+\s+for\s+(\w+)\s+from\s+([\d\.]+)`)
	consolePattern := regexp.MustCompile(`loginwindow.*Login Window.*[Ll]ogin|User logged in`)
	screenSharePattern := regexp.MustCompile(`screensharingd.*[Aa]uthenticat|[Cc]onnect`)

	processLine := func(line string) *notifier.LoginEvent {
		// Check SSH login
		if matches := sshPattern.FindStringSubmatch(line); matches != nil {
			return &notifier.LoginEvent{
				Username:  matches[1],
				Hostname:  w.hostname,
				IP:        matches[2],
				Terminal:  "ssh",
				Timestamp: time.Now(),
				OS:        "darwin",
			}
		}

		// Check console login
		if consolePattern.MatchString(line) {
			user := os.Getenv("USER")
			if user == "" {
				user = "console"
			}
			return &notifier.LoginEvent{
				Username:  user,
				Hostname:  w.hostname,
				Terminal:  "console",
				Timestamp: time.Now(),
				OS:        "darwin",
			}
		}

		// Check screen sharing
		if screenSharePattern.MatchString(line) {
			return &notifier.LoginEvent{
				Username:  "screensharing",
				Hostname:  w.hostname,
				Terminal:  "vnc",
				Timestamp: time.Now(),
				OS:        "darwin",
			}
		}

		return nil
	}

	predicate := `process == "loginwindow" OR process == "sshd" OR process == "screensharingd" OR (process == "securityd" AND eventMessage CONTAINS "Session")`

	// If since is specified, first check historical logs
	if opts.Since > 0 {
		minutes := int(opts.Since.Minutes())
		if minutes < 1 {
			minutes = 1
		}
		showCmd := exec.CommandContext(ctx, "log", "show",
			"--last", fmt.Sprintf("%dm", minutes),
			"--predicate", predicate,
			"--style", "compact",
		)

		output, err := showCmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if event := processLine(line); event != nil {
					select {
					case events <- *event:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}
		}
	}

	// Now start streaming new events
	cmd := exec.CommandContext(ctx, "log", "stream",
		"--predicate", predicate,
		"--style", "compact",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("darwin: failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("darwin: failed to start log stream: %w", err)
	}

	scanner := bufio.NewScanner(stdout)

	go func() {
		for scanner.Scan() {
			if event := processLine(scanner.Text()); event != nil {
				select {
				case events <- *event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Wait for context cancellation or command exit
	<-ctx.Done()
	return cmd.Wait()
}

// GetCurrentUser returns the current logged-in user on macOS
func GetCurrentUser() string {
	// Try using 'stat' on the console
	cmd := exec.Command("stat", "-f", "%Su", "/dev/console")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}

	// Fallback to USER environment variable
	if user := os.Getenv("USER"); user != "" {
		return user
	}

	return "unknown"
}
