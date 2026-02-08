//go:build linux

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

// LinuxWatcher watches for login events on Linux
type LinuxWatcher struct {
	hostname string
	logFile  string
}

func newPlatformWatcher() (Watcher, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Determine which log file to watch
	logFile := "/var/log/auth.log" // Debian/Ubuntu
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		logFile = "/var/log/secure" // RHEL/CentOS
	}

	return &LinuxWatcher{
		hostname: hostname,
		logFile:  logFile,
	}, nil
}

// Name returns the watcher name
func (w *LinuxWatcher) Name() string {
	return "linux"
}

// Watch monitors Linux auth logs for login events (new events only)
func (w *LinuxWatcher) Watch(ctx context.Context, events chan<- notifier.LoginEvent) error {
	return w.WatchWithOptions(ctx, events, Options{})
}

// WatchWithOptions monitors Linux auth logs with specific options
func (w *LinuxWatcher) WatchWithOptions(ctx context.Context, events chan<- notifier.LoginEvent, opts Options) error {
	// Patterns to detect login events
	// SSH login: "Accepted password for user from IP port ..."
	// SSH login: "Accepted publickey for user from IP port ..."
	sshPattern := regexp.MustCompile(`sshd\[\d+\]:\s+Accepted\s+\w+\s+for\s+(\w+)\s+from\s+([\d\.]+)\s+port\s+\d+`)
	// PAM session opened: "pam_unix(sshd:session): session opened for user xxx"
	pamPattern := regexp.MustCompile(`pam_unix\((\w+):session\):\s+session opened for user\s+(\w+)`)
	// TTY login: "LOGIN ON ttyX BY user"
	ttyPattern := regexp.MustCompile(`LOGIN ON\s+(\w+)\s+BY\s+(\w+)`)

	processLine := func(line string) *notifier.LoginEvent {
		// Check SSH login
		if matches := sshPattern.FindStringSubmatch(line); matches != nil {
			return &notifier.LoginEvent{
				Username:  matches[1],
				Hostname:  w.hostname,
				IP:        matches[2],
				Terminal:  "ssh",
				Timestamp: time.Now(),
				OS:        "linux",
			}
		}

		// Check PAM session
		if matches := pamPattern.FindStringSubmatch(line); matches != nil {
			service := matches[1]
			user := matches[2]
			// Avoid duplicate with SSH pattern
			if service != "sshd" {
				return &notifier.LoginEvent{
					Username:  user,
					Hostname:  w.hostname,
					Terminal:  service,
					Timestamp: time.Now(),
					OS:        "linux",
				}
			}
		}

		// Check TTY login
		if matches := ttyPattern.FindStringSubmatch(line); matches != nil {
			return &notifier.LoginEvent{
				Username:  matches[2],
				Hostname:  w.hostname,
				Terminal:  matches[1],
				Timestamp: time.Now(),
				OS:        "linux",
			}
		}

		return nil
	}

	// If since is specified, first check historical logs using journalctl or tail
	if opts.Since > 0 {
		minutes := int(opts.Since.Minutes())
		if minutes < 1 {
			minutes = 1
		}

		// Try journalctl first (systemd)
		journalCmd := exec.CommandContext(ctx, "journalctl",
			"--since", fmt.Sprintf("%d minutes ago", minutes),
			"-u", "sshd",
			"--no-pager",
		)

		if output, err := journalCmd.Output(); err == nil {
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
		} else {
			// Fallback: read log file with tail
			tailCmd := exec.CommandContext(ctx, "tail", "-n", "1000", w.logFile)
			if output, err := tailCmd.Output(); err == nil {
				lines := strings.Split(string(output), "\n")
				cutoff := time.Now().Add(-opts.Since)
				for _, line := range lines {
					if event := processLine(line); event != nil {
						// Note: We can't accurately parse log timestamps easily,
						// so we send all matching events from tail output
						if event.Timestamp.After(cutoff) || true {
							select {
							case events <- *event:
							case <-ctx.Done():
								return ctx.Err()
							}
						}
					}
				}
			}
		}
	}

	// Now watch for new events by tailing the log file
	// Use a reopenable file watcher to handle log rotation/truncation
	go func() {
		var file *os.File
		var reader *bufio.Reader
		var currentInode uint64
		var currentSize int64

		openFile := func() error {
			if file != nil {
				file.Close()
			}
			var err error
			file, err = os.Open(w.logFile)
			if err != nil {
				return err
			}
			// Seek to end of file to only watch new entries
			pos, err := file.Seek(0, 2)
			if err != nil {
				file.Close()
				return err
			}
			currentSize = pos

			// Get current inode
			if info, err := file.Stat(); err == nil {
				currentInode = getInode(info)
			}

			reader = bufio.NewReader(file)
			return nil
		}

		// Initial open
		if err := openFile(); err != nil {
			return
		}
		defer func() {
			if file != nil {
				file.Close()
			}
		}()

		checkTicker := time.NewTicker(5 * time.Second)
		defer checkTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-checkTicker.C:
				// Check if file was rotated/truncated
				info, err := os.Stat(w.logFile)
				if err != nil {
					// File might be temporarily unavailable, retry later
					continue
				}

				newInode := getInode(info)
				newSize := info.Size()

				needReopen := false

				// File was replaced (inode changed)
				if newInode != currentInode && newInode != 0 {
					needReopen = true
				}

				// File was truncated (size shrunk)
				if newSize < currentSize {
					needReopen = true
				}

				if needReopen {
					if err := openFile(); err == nil {
						// After reopen, seek to beginning to read new content
						file.Seek(0, 0)
						currentSize = 0
						if info, err := file.Stat(); err == nil {
							currentInode = getInode(info)
						}
						reader = bufio.NewReader(file)
					}
				}
			default:
				line, err := reader.ReadString('\n')
				if err != nil {
					// No new lines, wait a bit
					time.Sleep(100 * time.Millisecond)
					continue
				}

				// Update current size after reading
				if pos, err := file.Seek(0, 1); err == nil {
					currentSize = pos
				}

				if event := processLine(line); event != nil {
					select {
					case events <- *event:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	<-ctx.Done()
	return nil
}

// GetRecentLogins returns recent login records using 'last' command
func GetRecentLogins() ([]string, error) {
	data, err := os.ReadFile("/var/log/wtmp")
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}

// platformLogFiles returns log files for Linux
func platformLogFiles() []string {
	logFile := "/var/log/auth.log" // Debian/Ubuntu
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		logFile = "/var/log/secure" // RHEL/CentOS
	}
	return []string{logFile}
}
