package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/xsddz/whozere/internal/config"
	"github.com/xsddz/whozere/internal/notifier"
	"github.com/xsddz/whozere/internal/watcher"
)

const version = "0.1.0"

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	testNotify := flag.Bool("test", false, "Send a test notification and exit")
	since := flag.Duration("since", 0, "Check login events from this duration ago (e.g., 1h, 30m)")
	integrity := flag.Bool("integrity", true, "Enable log integrity monitoring (detect tampering)")
	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("whozere v%s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		fmt.Println("Who's here? - Login detection & notification tool")
		return
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// Create notifiers
	var notifiers []notifier.Notifier
	for _, nc := range cfg.Notifiers {
		if !nc.Enabled {
			continue
		}
		n, err := notifier.New(nc)
		if err != nil {
			log.Printf("Warning: failed to create notifier %s: %v", nc.Name, err)
			continue
		}
		notifiers = append(notifiers, n)
		log.Printf("Notifier enabled: %s", n.Name())
	}

	if len(notifiers) == 0 {
		log.Fatal("No notifiers available")
	}

	// Test mode: send a test notification
	if *testNotify {
		hostname, _ := os.Hostname()
		testEvent := notifier.LoginEvent{
			Username:  os.Getenv("USER"),
			Hostname:  hostname,
			Terminal:  "test",
			Timestamp: time.Now(),
			OS:        runtime.GOOS,
		}

		log.Println("Sending test notification...")
		for _, n := range notifiers {
			if err := n.Send(testEvent); err != nil {
				log.Printf("Failed to send test to %s: %v", n.Name(), err)
			} else {
				log.Printf("Test notification sent to %s", n.Name())
			}
		}
		return
	}

	// Create watcher
	w, err := watcher.New()
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	log.Printf("Using %s watcher", w.Name())

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Create event channel
	events := make(chan notifier.LoginEvent, 10)

	// Start watcher with options
	watchOpts := watcher.Options{Since: *since}
	go func() {
		if err := w.WatchWithOptions(ctx, events, watchOpts); err != nil && ctx.Err() == nil {
			log.Printf("Watcher error: %v", err)
		}
	}()

	// Start log integrity monitor if enabled and applicable
	if *integrity {
		logFiles := watcher.PlatformLogFiles()
		if len(logFiles) > 0 {
			monitor := watcher.NewLogIntegrityMonitor(logFiles, watcher.DefaultLogIntegrityOptions())
			go func() {
				if err := monitor.Start(ctx, events); err != nil && ctx.Err() == nil {
					log.Printf("Log integrity monitor error: %v", err)
				}
			}()
			log.Printf("Log integrity monitor started for: %v", logFiles)
		}
	}

	if *since > 0 {
		log.Printf("whozere v%s started, checking logins from %v ago and watching for new ones...", version, *since)
	} else {
		log.Printf("whozere v%s started, watching for logins...", version)
	}

	// Process events
	for {
		select {
		case event := <-events:
			log.Printf("Login detected: %s@%s (%s)", event.Username, event.Hostname, event.Terminal)
			for _, n := range notifiers {
				go func(n notifier.Notifier) {
					if err := n.Send(event); err != nil {
						log.Printf("Failed to send notification via %s: %v", n.Name(), err)
					}
				}(n)
			}
		case <-ctx.Done():
			log.Println("Shutdown complete")
			return
		}
	}
}
