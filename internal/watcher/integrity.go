package watcher

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/xsddz/whozere/internal/notifier"
)

// LogIntegrityOptions configures log integrity monitoring
type LogIntegrityOptions struct {
	// Enabled enables log integrity monitoring
	Enabled bool
	// CheckInterval is how often to check log integrity
	CheckInterval time.Duration
	// FileSizeDropThreshold triggers alert if file size drops by this percentage (0-100)
	// Set to 0 to disable
	FileSizeDropThreshold int
	// DetectDeletion alerts if log file is deleted
	DetectDeletion bool
	// DetectInodeChange alerts if file inode changes (file replaced)
	DetectInodeChange bool
	// DetectPermissionChange alerts if file permissions change
	DetectPermissionChange bool
	// Note: "No new logs" detection is intentionally NOT included
	// because it would cause too many false positives on idle systems
}

// DefaultLogIntegrityOptions returns sensible defaults
func DefaultLogIntegrityOptions() LogIntegrityOptions {
	return LogIntegrityOptions{
		Enabled:                true,
		CheckInterval:          30 * time.Second,
		FileSizeDropThreshold:  50, // Alert if file shrinks by 50%+
		DetectDeletion:         true,
		DetectInodeChange:      true,
		DetectPermissionChange: true,
	}
}

// LogIntegrityMonitor monitors log files for tampering
type LogIntegrityMonitor struct {
	files  []string
	opts   LogIntegrityOptions
	states map[string]*logFileState
	mu     sync.RWMutex
}

type logFileState struct {
	path     string
	size     int64
	inode    uint64
	mode     os.FileMode
	lastSeen time.Time
}

// NewLogIntegrityMonitor creates a new log integrity monitor
func NewLogIntegrityMonitor(files []string, opts LogIntegrityOptions) *LogIntegrityMonitor {
	return &LogIntegrityMonitor{
		files:  files,
		opts:   opts,
		states: make(map[string]*logFileState),
	}
}

// Start begins monitoring log files for tampering
func (m *LogIntegrityMonitor) Start(ctx context.Context, alerts chan<- notifier.LoginEvent) error {
	if !m.opts.Enabled {
		return nil
	}

	// Initialize states for all files
	for _, file := range m.files {
		if state, err := m.getFileState(file); err == nil {
			m.states[file] = state
		}
	}

	ticker := time.NewTicker(m.opts.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			m.checkIntegrity(alerts)
		}
	}
}

func (m *LogIntegrityMonitor) getFileState(path string) (*logFileState, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	state := &logFileState{
		path:     path,
		size:     info.Size(),
		mode:     info.Mode(),
		lastSeen: time.Now(),
	}
	// Get inode (platform-specific)
	state.inode = getInode(info)

	return state, nil
}

func (m *LogIntegrityMonitor) checkIntegrity(alerts chan<- notifier.LoginEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hostname, _ := os.Hostname()

	for _, file := range m.files {
		oldState := m.states[file]
		newState, err := m.getFileState(file)

		// Check if file was deleted
		if err != nil {
			if os.IsNotExist(err) && m.opts.DetectDeletion && oldState != nil {
				alerts <- notifier.LoginEvent{
					Username:  "SECURITY",
					Hostname:  hostname,
					Terminal:  "log-integrity",
					Timestamp: time.Now(),
					OS:        getOS(),
					IP:        fmt.Sprintf("⚠️ Log file DELETED: %s", file),
				}
				delete(m.states, file)
			}
			continue
		}

		// First time seeing this file
		if oldState == nil {
			m.states[file] = newState
			continue
		}

		// Check for file size drop (truncation)
		if m.opts.FileSizeDropThreshold > 0 && oldState.size > 0 {
			if newState.size < oldState.size {
				dropPercent := float64(oldState.size-newState.size) / float64(oldState.size) * 100
				if dropPercent >= float64(m.opts.FileSizeDropThreshold) {
					alerts <- notifier.LoginEvent{
						Username:  "SECURITY",
						Hostname:  hostname,
						Terminal:  "log-integrity",
						Timestamp: time.Now(),
						OS:        getOS(),
						IP:        fmt.Sprintf("⚠️ Log file TRUNCATED: %s (%.0f%% smaller)", file, dropPercent),
					}
				}
			}
		}

		// Check for inode change (file replaced)
		if m.opts.DetectInodeChange && oldState.inode != newState.inode {
			alerts <- notifier.LoginEvent{
				Username:  "SECURITY",
				Hostname:  hostname,
				Terminal:  "log-integrity",
				Timestamp: time.Now(),
				OS:        getOS(),
				IP:        fmt.Sprintf("⚠️ Log file REPLACED: %s (inode changed)", file),
			}
		}

		// Check for permission change
		if m.opts.DetectPermissionChange && oldState.mode != newState.mode {
			alerts <- notifier.LoginEvent{
				Username:  "SECURITY",
				Hostname:  hostname,
				Terminal:  "log-integrity",
				Timestamp: time.Now(),
				OS:        getOS(),
				IP:        fmt.Sprintf("⚠️ Log file PERMISSIONS changed: %s (%v → %v)", file, oldState.mode, newState.mode),
			}
		}

		// Update state
		m.states[file] = newState
	}
}

func getOS() string {
	switch {
	case fileExists("/var/log/auth.log") || fileExists("/var/log/secure"):
		return "linux"
	default:
		return "unknown"
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
