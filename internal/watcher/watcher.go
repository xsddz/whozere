package watcher

import (
	"context"
	"time"

	"github.com/xsddz/whozere/internal/notifier"
)

// Options configures watcher behavior
type Options struct {
	// Since specifies how far back to check for login events
	// Zero means only watch new events (no history)
	Since time.Duration
}

// Watcher is the interface for login detection
type Watcher interface {
	// Watch starts watching for login events
	// It should send login events to the provided channel
	// The watcher should stop when the context is cancelled
	Watch(ctx context.Context, events chan<- notifier.LoginEvent) error

	// WatchWithOptions starts watching with specific options
	WatchWithOptions(ctx context.Context, events chan<- notifier.LoginEvent, opts Options) error

	// Name returns the name of this watcher
	Name() string
}

// New creates a new watcher for the current platform
func New() (Watcher, error) {
	return newPlatformWatcher()
}

// PlatformLogFiles returns the log files monitored on the current platform
// Returns nil for platforms that don't use log files (e.g., macOS, Windows)
func PlatformLogFiles() []string {
	return platformLogFiles()
}
