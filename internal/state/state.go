// Package state manages the persistent state of the Tux pet.
package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// State represents the current state of the Tux pet.
// All values are clamped between 0 and 100.
type State struct {
	Hunger  int       `json:"hunger"`  // 0 = starving, 100 = full (fullness level)
	Mood    int       `json:"mood"`    // 0 = miserable, 100 = ecstatic
	Energy  int       `json:"energy"`  // 0 = exhausted, 100 = full of energy
	LastUpdate time.Time `json:"last_update"`
}

// DefaultState returns the initial state for a new pet.
func DefaultState() State {
	return State{
		Hunger:    80,  // Starts fairly full
		Mood:      80,
		Energy:    80,
		LastUpdate: time.Now(),
	}
}

// NewState loads the state from the config directory.
// If no state file exists, it returns the default state.
func NewState(dataDir string) (*State, error) {
	statePath := filepath.Join(dataDir, "state.json")

	// Try to load existing state
	data, err := os.ReadFile(statePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// No state file exists, return default
			s := DefaultState()
			return &s, nil
		}
		return nil, err
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

// Save persists the current state to disk.
func (s *State) Save(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	statePath := filepath.Join(dataDir, "state.json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0644)
}

// clamp ensures a value is between 0 and 100.
func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// Update applies time-based changes to the pet's state.
// This should be called periodically (every 10 minutes).
func (s *State) Update() {
	// Hunger decreases over time (gets hungry)
	s.Hunger = clamp(s.Hunger - 5)
	s.Energy = clamp(s.Energy - 3)

	// Mood changes based on hunger (fullness)
	if s.Hunger < 20 {
		s.Mood = clamp(s.Mood - 10)
	} else if s.Hunger > 60 {
		s.Mood = clamp(s.Mood + 5)
	}

	s.LastUpdate = time.Now()
}

// Feed feeds the pet, increasing hunger (fullness) and improving mood.
func (s *State) Feed() {
	s.Hunger = clamp(s.Hunger + 30)
	s.Mood = clamp(s.Mood + 5)
	s.LastUpdate = time.Now()
}

// Play plays with the pet, consuming energy but improving mood.
func (s *State) Play() {
	s.Energy = clamp(s.Energy - 10)
	s.Mood = clamp(s.Mood + 15)
	s.LastUpdate = time.Now()
}

// Sleep lets the pet rest, restoring energy.
func (s *State) Sleep() {
	s.Energy = clamp(s.Energy + 40)
	s.LastUpdate = time.Now()
}

// GetMoodLabel returns a human-readable mood description.
func (s *State) GetMoodLabel() string {
	if s.Mood >= 80 {
		return "Happy"
	}
	if s.Mood >= 50 {
		return "Neutral"
	}
	if s.Mood >= 30 {
		return "Sad"
	}
	return "Angry"
}

// IsAngry returns true if the pet is angry (very hungry or low mood).
func (s *State) IsAngry() bool {
	return s.Hunger < 20 || s.Mood < 30
}
