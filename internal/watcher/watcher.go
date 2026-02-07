package watcher

import (
	"context"

	"github.com/xsddz/whozere/internal/notifier"
)

// Watcher is the interface for login detection
type Watcher interface {
	// Watch starts watching for login events
	// It should send login events to the provided channel
	// The watcher should stop when the context is cancelled
	Watch(ctx context.Context, events chan<- notifier.LoginEvent) error

	// Name returns the name of this watcher
	Name() string
}

// New creates a new watcher for the current platform
func New() (Watcher, error) {
	return newPlatformWatcher()
}
