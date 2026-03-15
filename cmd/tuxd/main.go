// Tuxd - The Tux pet daemon.
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/imns/tux/internal/client"
	"github.com/imns/tux/internal/daemon"
)

func main() {
	logFile := flag.String("log", "", "Log file path (default: stderr)")
	flag.Parse()

	// Setup logging
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// Create daemon
	d, err := daemon.New(client.DefaultDataDir())
	if err != nil {
		log.Fatalf("Failed to create daemon: %v", err)
	}

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start daemon in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- d.Run()
	}()

	// Wait for signal or error
	select {
	case <-sigCh:
		log.Println("Received shutdown signal")
		d.Stop()
		<-errCh // Wait for Run() to complete
	case err := <-errCh:
		if err != nil {
			log.Fatalf("Daemon error: %v", err)
		}
	}
}
