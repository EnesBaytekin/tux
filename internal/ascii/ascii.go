// Package ascii provides ASCII art rendering of the Tux pet.
package ascii

import (
	"fmt"
	"strings"
)

// Art represents the ASCII art for different moods.
var art = map[string]string{
	"Happy": `
   /\_/\
  ( ^.^ )
  /  >  \`,
	"Neutral": `
   /\_/\
  ( o.o )
  /  >  \`,
	"Sad": `
   /\_/\
  ( -.- )
  /  >  \`,
	"Angry": `
   /\_/\
  ( >.< )
  /  >  \`,
}

// Render returns the ASCII art for the given mood.
func Render(mood string) string {
	if a, ok := art[mood]; ok {
		return a
	}
	return art["Neutral"]
}

// Display renders the pet with its name and status.
func Display(name, mood string) string {
	var sb strings.Builder
	sb.WriteString(Render(mood))
	sb.WriteString("\n")

	// Center the name
	padding := (8 - len(name)) / 2
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}
	sb.WriteString(name)
	sb.WriteString("\n")

	return sb.String()
}

// DisplayWithStats renders the pet with all status information.
func DisplayWithStats(name string, hunger, mood, energy int) string {
	moodLabel := getMoodLabel(mood)
	var sb strings.Builder

	sb.WriteString(Display(name, moodLabel))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Mood: %s\n", moodLabel))
	sb.WriteString(fmt.Sprintf("Hunger: %d\n", hunger))
	sb.WriteString(fmt.Sprintf("Energy: %d\n", energy))

	return sb.String()
}

// getMoodLabel converts a numeric mood value to a label.
func getMoodLabel(mood int) string {
	if mood >= 80 {
		return "Happy"
	}
	if mood >= 50 {
		return "Neutral"
	}
	if mood >= 30 {
		return "Sad"
	}
	return "Angry"
}
