// Package daemon implements the background daemon for Tux.
package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/imns/tux/internal/state"
)

const (
	updateInterval = 10 * time.Minute
	socketPath     = "socket"
)

// Command represents a command sent from the CLI.
type Command struct {
	Action string `json:"action"`
}

// Response represents the daemon's response to a command.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	State   *State `json:"state,omitempty"`
}

// State is a serializable version of the pet state for responses.
type State struct {
	Hunger     int    `json:"hunger"`
	Mood       int    `json:"mood"`
	Energy     int    `json:"energy"`
	MoodLabel  string `json:"mood_label"`
	LastUpdate string `json:"last_update"`
}

// Daemon runs the background pet daemon.
type Daemon struct {
	state    *state.State
	dataDir  string
	socket   string
	mu       sync.Mutex
	stopCh   chan struct{}
}

// New creates a new daemon instance.
func New(dataDir string) (*Daemon, error) {
	s, err := state.NewState(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	socketPath := filepath.Join(dataDir, socketPath)

	return &Daemon{
		state:  s,
		dataDir: dataDir,
		socket: socketPath,
		stopCh: make(chan struct{}),
	}, nil
}

// Run starts the daemon.
func (d *Daemon) Run() error {
	// Start the update ticker
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	// Start the socket server
	if err := d.startSocketServer(); err != nil {
		return fmt.Errorf("failed to start socket server: %w", err)
	}

	log.Printf("Tux daemon started, socket: %s", d.socket)

	// Main loop
	for {
		select {
		case <-ticker.C:
			d.updateState()
		case <-d.stopCh:
			log.Println("Tux daemon stopping")
			return nil
		}
	}
}

// Stop gracefully stops the daemon.
func (d *Daemon) Stop() {
	close(d.stopCh)
}

// startSocketServer creates and listens on the Unix domain socket.
func (d *Daemon) startSocketServer() error {
	// Remove existing socket if present
	os.Remove(d.socket)

	// Create socket directory if needed
	if err := os.MkdirAll(filepath.Dir(d.socket), 0755); err != nil {
		return err
	}

	listener, err := net.Listen("unix", d.socket)
	if err != nil {
		return err
	}

	// Set socket permissions
	if err := os.Chmod(d.socket, 0660); err != nil {
		listener.Close()
		return err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-d.stopCh:
					return
				default:
					log.Printf("Accept error: %v", err)
					continue
				}
			}

			go d.handleConnection(conn)
		}
	}()

	return nil
}

// handleConnection handles a single client connection.
func (d *Daemon) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return
	}

	var cmd Command
	if err := json.Unmarshal(scanner.Bytes(), &cmd); err != nil {
		d.sendError(conn, "invalid command format")
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	resp := d.handleCommand(cmd.Action)
	d.sendResponse(conn, resp)
}

// handleCommand processes a command and returns a response.
func (d *Daemon) handleCommand(action string) Response {
	switch action {
	case "feed":
		d.state.Feed()
		if err := d.state.Save(d.dataDir); err != nil {
			return Response{
				Success: false,
				Message: fmt.Sprintf("failed to save state: %v", err),
			}
		}
		return Response{
			Success: true,
			Message: "Tux has been fed!",
			State:   d.serializeState(),
		}

	case "play":
		d.state.Play()
		if err := d.state.Save(d.dataDir); err != nil {
			return Response{
				Success: false,
				Message: fmt.Sprintf("failed to save state: %v", err),
			}
		}
		return Response{
			Success: true,
			Message: "Tux had fun playing!",
			State:   d.serializeState(),
		}

	case "sleep":
		d.state.Sleep()
		if err := d.state.Save(d.dataDir); err != nil {
			return Response{
				Success: false,
				Message: fmt.Sprintf("failed to save state: %v", err),
			}
		}
		return Response{
			Success: true,
			Message: "Tux is resting peacefully.",
			State:   d.serializeState(),
		}

	case "status":
		return Response{
			Success: true,
			State:   d.serializeState(),
		}

	default:
		return Response{
			Success: false,
			Message: fmt.Sprintf("unknown command: %s", action),
		}
	}
}

// updateState applies periodic state updates.
func (d *Daemon) updateState() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.state.Update()
	if err := d.state.Save(d.dataDir); err != nil {
		log.Printf("Failed to save state: %v", err)
	}
}

// serializeState converts the state to a JSON-serializable format.
func (d *Daemon) serializeState() *State {
	return &State{
		Hunger:     d.state.Hunger,
		Mood:       d.state.Mood,
		Energy:     d.state.Energy,
		MoodLabel:  d.state.GetMoodLabel(),
		LastUpdate: d.state.LastUpdate.Format(time.RFC3339),
	}
}

// sendResponse sends a JSON response to the client.
func (d *Daemon) sendResponse(conn net.Conn, resp Response) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	fmt.Fprintln(conn, string(data))
}

// sendError sends an error response to the client.
func (d *Daemon) sendError(conn net.Conn, msg string) {
	resp := Response{
		Success: false,
		Message: msg,
	}
	d.sendResponse(conn, resp)
}
