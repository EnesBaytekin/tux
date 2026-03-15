// Package ascii provides ASCII art rendering of the Tux pet.
package ascii

import (
	"fmt"
	"strings"
)

// getPenguinEye returns the eye character based on state.
func getPenguinEye(hunger, mood, energy int) string {
	// Yatarken her zaman kapalı göz
	if energy < 30 {
		return "-"
	}

	// Çok açsa şaşkın
	if hunger < 20 {
		return "o"
	}

	// Çok toksa hevesli
	if hunger >= 95 {
		return "*"
	}

	// Mood'a göre göz
	if mood >= 80 {
		return "^"
	}
	if mood >= 50 {
		return "."
	}
	if mood >= 30 {
		return "-"
	}
	return ">"
}

// Render returns the ASCII art based on current state.
func Render(hunger, mood, energy int) string {
	eye := getPenguinEye(hunger, mood, energy)

	// Energy düşükse yatan penguen
	if energy < 30 {
		lines := []string{
			"        ___",
			"      ,'   '-.__",
			"     /  --' )  " + eye + ")=-",
			"  --'--'-------'",
		}
		return strings.Join(lines, "\n")
	}

	// Çok mutluysa ve enerjisi yüksekse kanatlarını açan penguen
	if mood >= 80 {
		lines := []string{
			"    --.   __",
			"   (   \\.' " + eye + ")=-",
			"    `.  '-.-",
			"      ;-  |\\",
			"      |   |'",
			"    _,:__/_",
		}
		return strings.Join(lines, "\n")
	}

	// Normal ayakta penguen
	lines := []string{
		"          __",
		"        -' " + eye + ")=-",
		"       /.-.'",
		"      //  |\\",
		"      ||  |'",
		"    _,;(_/_",
	}
	return strings.Join(lines, "\n")
}

// Display renders the pet with its name.
func Display(name string, hunger, moodValue, energy int) string {
	var sb strings.Builder
	sb.WriteString(Render(hunger, moodValue, energy))
	sb.WriteString("\n")

	// Center the name
	padding := (13 - len(name)) / 2
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}
	sb.WriteString(name)
	sb.WriteString("\n")

	return sb.String()
}

// DisplayWithStats renders the pet with all status information.
func DisplayWithStats(name string, hunger, moodValue, energy int) string {
	moodLabel := getMoodLabel(moodValue)
	stateLabel := getStateLabel(hunger, moodValue, energy)
	var sb strings.Builder

	sb.WriteString(Display(name, hunger, moodValue, energy))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Mood: %s\n", moodLabel))
	sb.WriteString(fmt.Sprintf("Hunger: %d\n", hunger))
	sb.WriteString(fmt.Sprintf("Energy: %d\n", energy))
	if stateLabel != "" {
		sb.WriteString(fmt.Sprintf("%s\n", stateLabel))
	}

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

// getStateLabel returns additional state info.
func getStateLabel(hunger, mood, energy int) string {
	if energy < 30 {
		return "Sleeping..."
	}
	if hunger < 20 {
		return "Very hungry!"
	}
	if hunger >= 90 {
		return "Too full!"
	}
	return ""
}
