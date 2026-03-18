// Tux - A terminal pet penguin.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/imns/tux/internal/ascii"
	"github.com/imns/tux/internal/game"
	"github.com/imns/tux/internal/state"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Tux - A terminal pet penguin\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  tux [command]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  (none)   Show pet's current state\n")
		fmt.Fprintf(os.Stderr, "  feed     Feed the pet\n")
		fmt.Fprintf(os.Stderr, "  play     Play mini game with the pet\n")
		fmt.Fprintf(os.Stderr, "  sleep    Let the pet sleep\n")
		fmt.Fprintf(os.Stderr, "  status   Show pet's stats\n")
		fmt.Fprintf(os.Stderr, "  rename   Rename the pet\n")
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
		fmt.Printf("%s has been fed!\n", s.Name)
	case "play":
		// Start minigame
		g := game.NewGame()
		if err := g.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting game: %v\n", err)
			os.Exit(1)
		}

		// Check if player played enough (minimum score)
		const minScore = 100
		score := g.GetScore()
		if score >= minScore {
			// Player played enough, apply play effects
			s.Play()
			fmt.Printf("\n%s had fun playing! (Score: %d)\n", s.Name, score)
		} else {
			// Didn't play enough, no effect
			fmt.Printf("\nGame over! (Score: %d) - Play a bit longer next time!\n", score)
		}
	case "sleep":
		s.Sleep()
		fmt.Printf("%s is resting.\n", s.Name)
	case "status":
		// Just show state
	case "rename":
		renamePet(s, dataDir)
		return // Don't display state after rename
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
	fmt.Println(ascii.DisplayWithStats(s.Name, s.Hunger, s.Mood, s.Energy))
}

// renamePet handles the rename command with confirmation.
func renamePet(s *state.State, dataDir string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Current name: %s\n", s.Name)
	fmt.Print("Enter new name (leave empty to cancel): ")

	newName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	newName = strings.TrimSpace(newName)
	if newName == "" {
		fmt.Println("Rename cancelled.")
		return
	}

	// Check if same name
	if strings.EqualFold(newName, s.Name) {
		fmt.Printf("That's already the pet's name!\n")
		return
	}

	// Confirm by re-typing
	fmt.Printf("Confirm: Type '%s' again to confirm: ", newName)
	confirm, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	confirm = strings.TrimSpace(confirm)
	if !strings.EqualFold(confirm, newName) {
		fmt.Println("Names don't match. Rename cancelled.")
		return
	}

	// Update name
	s.Name = newName
	if err := s.Save(dataDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving state: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nPet renamed to '%s'!\n", newName)
	fmt.Println(ascii.DisplayWithStats(s.Name, s.Hunger, s.Mood, s.Energy))
}
