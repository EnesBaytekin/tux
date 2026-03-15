// Tux - A terminal pet penguin.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imns/tux/internal/ascii"
	"github.com/imns/tux/internal/state"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Tux - A terminal pet penguin\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  tux [command]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  (none)   Show Tux's current state\n")
		fmt.Fprintf(os.Stderr, "  feed     Feed Tux\n")
		fmt.Fprintf(os.Stderr, "  play     Play with Tux\n")
		fmt.Fprintf(os.Stderr, "  sleep    Let Tux sleep\n")
		fmt.Fprintf(os.Stderr, "  status   Show Tux's stats\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()

	var action string
	if len(args) == 0 {
		action = "status"
	} else {
		action = args[0]
	}

	// Get data directory
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}
	dataDir := filepath.Join(home, ".local", "share", "tux")

	// Load state
	s, err := state.NewState(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
		os.Exit(1)
	}

	// Update state based on time passed since last interaction
	s.UpdateSinceLast()

	// Execute action
	switch action {
	case "feed":
		s.Feed()
		fmt.Println("Tux has been fed!")
	case "play":
		s.Play()
		fmt.Println("Tux had fun playing!")
	case "sleep", "status":
		// Just show state, sleep is automatic based on energy
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", action)
		os.Exit(1)
	}

	// Save state
	if err := s.Save(dataDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving state: %v\n", err)
		os.Exit(1)
	}

	// Display current state
	fmt.Println(ascii.DisplayWithStats("Tux", s.Hunger, s.Mood, s.Energy))
}
