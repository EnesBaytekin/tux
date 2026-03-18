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
// Values are stored as float64 to preserve precision.
type State struct {
	Name      string     `json:"name"`      // Pet's name
	Hunger    float64    `json:"hunger"`    // 0 = starving, 100 = full (fullness level)
	Mood      float64    `json:"mood"`      // 0 = miserable, 100 = ecstatic
	Energy    float64    `json:"energy"`    // 0 = exhausted, 100 = full of energy
	LastUpdate time.Time `json:"last_update"`
}

// DefaultState returns the initial state for a new pet.
func DefaultState() State {
	return State{
		Name:       "Tux",  // Default name
		Hunger:     80.0,   // Starts fairly full
		Mood:       80.0,
		Energy:     80.0,
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

	// Ensure name is not empty (for backward compatibility)
	if s.Name == "" {
		s.Name = "Tux"
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

// clampFloat ensures a float64 value is between 0 and 100.
func clampFloat(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// clampInt ensures a value is between 0 and 100 (alias for clamp).
func clampInt(v int) int {
	return clamp(v)
}

// Update applies time-based changes to the pet's state.
// This applies one 10-minute cycle of changes.
func (s *State) update() {
	// Hunger decreases over time (gets hungry)
	s.Hunger = clampFloat(s.Hunger - 5.0)
	s.Energy = clampFloat(s.Energy - 3.0)

	// Mood changes based on hunger (fullness)
	if s.Hunger < 20 {
		s.Mood = clampFloat(s.Mood - 10.0)
	} else if s.Hunger > 60 {
		s.Mood = clampFloat(s.Mood + 5.0)
	}

	s.LastUpdate = time.Now()
}

// UpdateSinceLast calculates time passed and applies linear updates.
// This should be called before any interaction to ensure state is current.
// State changes happen per-minute (actually per-second, continuously).
func (s *State) UpdateSinceLast() {
	now := time.Now()
	elapsed := now.Sub(s.LastUpdate)

	// Calculate minutes passed (can be fractional, e.g., 0.5 minutes = 30 seconds)
	minutes := elapsed.Minutes()

	if minutes < 0.0001 {
		// Less than a millisecond passed, no update needed
		return
	}

	// Apply linear changes based on minutes passed
	// Per 10 minutes: hunger -5, energy -3
	// Per 1 minute: hunger -0.5, energy -0.3
	s.Hunger = clampFloat(s.Hunger - (minutes * 0.5))
	s.Energy = clampFloat(s.Energy - (minutes * 0.3))

	// Mood changes: slowly decreases over time, faster if hungry
	// Default: -0.05 per minute (very slow natural decrease)
	// If hungry (< 20): -1.0 per minute (much faster)
	// If full (>= 90): +0.5 per minute (improves when very full)
	if s.Hunger < 20 {
		s.Mood = clampFloat(s.Mood - (minutes * 1.0))
	} else if s.Hunger >= 90 {
		s.Mood = clampFloat(s.Mood + (minutes * 0.5))
	} else {
		s.Mood = clampFloat(s.Mood - (minutes * 0.05))
	}

	// Update timestamp to now
	s.LastUpdate = now
}

// Feed feeds the pet, increasing hunger (fullness) and improving mood.
func (s *State) Feed() {
	s.Hunger = clampFloat(s.Hunger + 10.0)
	s.Mood = clampFloat(s.Mood + 2.0)
	s.LastUpdate = time.Now()
}

// Play plays with the pet, consuming energy but improving mood.
func (s *State) Play() {
	s.Energy = clampFloat(s.Energy - 10.0)
	s.Mood = clampFloat(s.Mood + 15.0)
	s.LastUpdate = time.Now()
}

// Sleep lets the pet rest, restoring energy.
func (s *State) Sleep() {
	s.Energy = clampFloat(s.Energy + 15.0)
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
