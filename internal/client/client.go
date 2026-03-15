// Package client implements the CLI client for communicating with the Tux daemon.
package client

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Command represents a command to send to the daemon.
type Command struct {
	Action string `json:"action"`
}

// Response represents the daemon's response.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	State   *State `json:"state,omitempty"`
}

// State represents the pet's state as returned by the daemon.
type State struct {
	Hunger     int    `json:"hunger"`
	Mood       int    `json:"mood"`
	Energy     int    `json:"energy"`
	MoodLabel  string `json:"mood_label"`
	LastUpdate string `json:"last_update"`
}

// DefaultDataDir returns the default data directory for Tux.
func DefaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/tux"
	}
	return filepath.Join(home, ".local", "share", "tux")
}

// SendCommand sends a command to the daemon and returns the response.
func SendCommand(action string) (*Response, error) {
	dataDir := DefaultDataDir()
	socketPath := filepath.Join(dataDir, "socket")

	// Connect to daemon socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w (is tuxd running?)", err)
	}
	defer conn.Close()

	// Set connection timeout
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send command
	cmd := Command{Action: action}
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	if _, err := fmt.Fprintln(conn, string(data)); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	var response Response
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &response, nil
}

// FormatTime formats a timestamp string for display.
func FormatTime(s string) string {
	if s == "" {
		return "unknown"
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}

	return t.Format("2006-01-02 15:04:05")
}
