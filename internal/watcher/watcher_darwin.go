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

// Watch monitors macOS system logs for login events
func (w *DarwinWatcher) Watch(ctx context.Context, events chan<- notifier.LoginEvent) error {
	// Use `log stream` to monitor login-related events
	// Filter for login/authentication related processes
	cmd := exec.CommandContext(ctx, "log", "stream",
		"--predicate", `process == "loginwindow" OR process == "sshd" OR process == "screensharingd" OR (process == "securityd" AND eventMessage CONTAINS "Session")`,
		"--style", "compact",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("darwin: failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("darwin: failed to start log stream: %w", err)
	}

	// Patterns to detect login events
	// SSH login pattern: "Accepted publickey for user from IP port ..."
	sshPattern := regexp.MustCompile(`sshd.*Accepted\s+\w+\s+for\s+(\w+)\s+from\s+([\d\.]+)`)
	// Console login pattern
	consolePattern := regexp.MustCompile(`loginwindow.*Login Window.*[Ll]ogin|User logged in`)
	// Screen sharing pattern
	screenSharePattern := regexp.MustCompile(`screensharingd.*[Aa]uthenticat|[Cc]onnect`)

	scanner := bufio.NewScanner(stdout)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()

			var event *notifier.LoginEvent

			// Check SSH login
			if matches := sshPattern.FindStringSubmatch(line); matches != nil {
				event = &notifier.LoginEvent{
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
				// Get current user
				user := os.Getenv("USER")
				if user == "" {
					user = "console"
				}
				event = &notifier.LoginEvent{
					Username:  user,
					Hostname:  w.hostname,
					Terminal:  "console",
					Timestamp: time.Now(),
					OS:        "darwin",
				}
			}

			// Check screen sharing
			if screenSharePattern.MatchString(line) {
				event = &notifier.LoginEvent{
					Username:  "screensharing",
					Hostname:  w.hostname,
					Terminal:  "vnc",
					Timestamp: time.Now(),
					OS:        "darwin",
				}
			}

			if event != nil {
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
