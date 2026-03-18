package game

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"golang.org/x/term"
)

// Game constants
const (
	GameWidth  = 53
	GameHeight = 17
)

// Game state
type Game struct {
	penguinPos       int
	icebergs         []Iceberg
	score            int
	running          bool
	oldState         *term.State
	inputChan        chan rune
	cancelChan       chan struct{}
	passedIcebergs   map[int]bool // Track which icebergs we've passed
	bonusMessage     string
	bonusTimer       float64
	icebergCollision map[int]bool // Track if penguin was ever in wrong Y position during X collision
}

type Iceberg struct {
	x      int
	y      int
	art    []string
	width  int
	height int
}

// Iceberg art variations
var icebergArts = [][][]string{
	{
		{
			"    _",
			"  _/ |",
			" / )  \\",
			"(___)__)",
		},
	},
	{
		{
			"  _",
			" | \\",
			"/___)",
		},
	},
	{
		{
			"     _",
			"  __/ \\",
			" /  )  \\",
			"(____)__)",
		},
	},
}

// Penguin art
const penguinArt = `
  __
=(__)O>`

func NewGame() *Game {
	return &Game{
		penguinPos:        GameHeight / 2, // Start in middle
		icebergs:          make([]Iceberg, 0),
		score:             0,
		running:           true,
		inputChan:         make(chan rune, 10),
		cancelChan:        make(chan struct{}),
		passedIcebergs:    make(map[int]bool),
		bonusMessage:      "",
		bonusTimer:        0,
		icebergCollision:  make(map[int]bool),
	}
}

func (g *Game) Start() error {
	// Save terminal state and set raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	g.oldState = oldState

	// Restore terminal on exit
	defer g.restoreTerminal()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		g.Stop()
	}()

	// Start input reader
	go g.readInput()

	// Clear screen and hide cursor
	fmt.Print("\x1b[2J\x1b[?25l")

	// Add initial iceberg
	g.addIceberg()

	// Game loop
	defer close(g.inputChan) // Close input channel to stop readInput goroutine

	// Random iceberg spawn timer
	lastSpawn := time.Now()
	nextSpawnDelay := time.Duration(800+rand.Intn(1200)) * time.Millisecond

	// Fixed delta time for consistent 30 FPS
	const targetFPS = 30.0
	const fixedDeltaTime = 1.0 / targetFPS // ~0.033 seconds

	for g.running {
		now := time.Now()

		// Check if we should spawn icebergs
		if now.Sub(lastSpawn) >= nextSpawnDelay {
			// Always spawn at least one iceberg
			g.addIceberg()

			// 50% chance to spawn a second iceberg at different position
			if rand.Intn(2) == 0 {
				g.addIceberg()
			}

			lastSpawn = now
			nextSpawnDelay = time.Duration(800+rand.Intn(1200)) * time.Millisecond
		}

		// Always update and render
		g.update(fixedDeltaTime)
		g.render()

		// Sleep for consistent 30 FPS
		time.Sleep(33 * time.Millisecond)

		// Check for input (non-blocking)
		select {
		case key := <-g.inputChan:
			g.handleInput(key)
		case <-g.cancelChan:
			g.running = false
		default:
			// No input, continue
		}
	}

	g.renderGameOver()
	return nil
}

func (g *Game) Stop() {
	g.running = false
	close(g.cancelChan)
}

func (g *Game) restoreTerminal() {
	// Just show cursor and restore terminal state - NO CLEAR
	// (Game over screen has already been rendered)
	fmt.Print("\x1b[?25h")
	if g.oldState != nil {
		term.Restore(int(os.Stdin.Fd()), g.oldState)
	}
}

func (g *Game) readInput() {
	buf := make([]byte, 1)
	escapeSeq := false
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			return
		}

		ch := rune(buf[0])

		// Handle escape sequences (arrow keys, etc.)
		if ch == 27 { // ESC
			escapeSeq = true
			continue
		}
		if escapeSeq {
			// Skip the rest of the escape sequence
			if ch >= 60 && ch <= 90 { // Typically the character after ESC [
				escapeSeq = false
				continue
			}
			escapeSeq = false
		}

		// Check if channel is still open
		select {
		case g.inputChan <- ch:
		default:
			return // Channel closed, stop reading
		}
	}
}

func (g *Game) handleInput(key rune) {
	switch key {
	case 'w', 'W': // Up
		if g.penguinPos > -1 {
			g.penguinPos -= 1
		}
	case 's', 'S': // Down
		if g.penguinPos < GameHeight-2 {
			g.penguinPos += 1
		}
	case 'q', 'Q', 27: // Quit
		g.Stop()
	}
}

func (g *Game) update(deltaTime float64) {
	// Move icebergs left (speed is pixels per second, adjusted by deltaTime)
	// At 30 FPS, we want similar speed to before (1 pixel per 60ms tick)
	// Old: 1 pixel per 60ms = 16.67 pixels/second
	// New: Target 30 FPS, so per-frame movement should be similar
	icebergSpeed := 30.0 // pixels per second (slower, easier)
	pixelsToMove := icebergSpeed * deltaTime

	for i := range g.icebergs {
		g.icebergs[i].x -= int(pixelsToMove)
	}

	// Remove off-screen icebergs and add score
	for i := len(g.icebergs) - 1; i >= 0; i-- {
		if g.icebergs[i].x < -10 {
			g.icebergs = append(g.icebergs[:i], g.icebergs[i+1:]...)
			delete(g.passedIcebergs, i) // Clean up tracking
			g.score += 10
		}
	}

	// Check if penguin passed any icebergs (Y position: penguin bottom vs iceberg bottom)
	penguinBottomY := g.penguinPos + 2
	for i, iceberg := range g.icebergs {
		if !g.passedIcebergs[i] {
			penguinLeftX := 3
			penguinRightX := 11
			icebergLeftX := iceberg.x
			icebergRightX := iceberg.x + iceberg.width
			icebergBottomY := iceberg.y + iceberg.height
			icebergTopY := iceberg.y

			// Check if penguin is colliding horizontally with iceberg
			if penguinRightX > icebergLeftX && penguinLeftX < icebergRightX {
				// Check Y position - must be at iceberg top (+1) OR bottom during entire X collision
				correctY := (penguinBottomY == icebergBottomY || penguinBottomY == icebergBottomY+1 ||
					penguinBottomY == icebergTopY+1)

				// If Y is wrong during X collision, mark as invalid for bonus
				if !correctY {
					g.icebergCollision[i] = true // Mark that penguin was in wrong Y position
				}
				// Initialize tracking if not set (false means "so far so good")
				if !g.icebergCollision[i] {
					g.icebergCollision[i] = false
				}
			} else {
				// X collision ended - iceberg is now fully to the left of penguin
				if icebergRightX < penguinLeftX {
					// Check if we have tracking data for this iceberg
					if tracked, exists := g.icebergCollision[i]; exists && !tracked {
						// Penguin maintained correct Y throughout X collision - bonus!
						g.passedIcebergs[i] = true
						g.score += 20 // Bonus for passing an iceberg

						// Show bonus message
						messages := []string{
							"NICE!",
							"GREAT!",
							"SWEET!",
							"AWESOME!",
						}
						g.bonusMessage = messages[rand.Intn(len(messages))]
						g.bonusTimer = 0.5 // Show for 0.5 seconds
					}
					// Clean up tracking
					delete(g.icebergCollision, i)
				}
			}
		}
	}

	// Update bonus timer
	if g.bonusTimer > 0 {
		g.bonusTimer -= deltaTime
		if g.bonusTimer <= 0 {
			g.bonusMessage = ""
		}
	}

	// Check collision
	if g.checkCollision() {
		g.running = false
	}

	// Increment score for survival (10 points per second)
	g.score += int(10 * deltaTime)
}

func (g *Game) addIceberg() {
	// Pick random iceberg art
	artVariants := icebergArts[rand.Intn(len(icebergArts))]
	art := artVariants[0]

	// Spawn to the RIGHT of the screen, in a 16-pixel zone
	// Icebergs will then move left into the screen
	x := GameWidth + rand.Intn(16)
	y := rand.Intn(GameHeight-3-len(art)) + 2

	g.icebergs = append(g.icebergs, Iceberg{
		x:      x,
		y:      y,
		art:    art,
		width:  len(art[0]),
		height: len(art),
	})
}

func (g *Game) checkCollision() bool {
	// Penguin collision: only bottom row matters, ignore underscores
	penguinX := 3
	penguinY := g.penguinPos + 1 // Only use second row
	penguinLine := "=(__)>"

	// Create a set of occupied positions for penguin (excluding _)
	penguinOccupied := make(map[[2]int]bool)
	for dx, ch := range penguinLine {
		if ch != ' ' && ch != '_' {
			penguinOccupied[[2]int{penguinX + dx, penguinY}] = true
		}
	}

	// Check each iceberg
	for _, iceberg := range g.icebergs {
		// Create a set of occupied positions for this iceberg (excluding _)
		icebergOccupied := make(map[[2]int]bool)
		for dy, line := range iceberg.art {
			for dx, ch := range line {
				if ch != ' ' && ch != '_' {
					icebergOccupied[[2]int{iceberg.x + dx, iceberg.y + dy}] = true
				}
			}
		}

		// Check for any overlap
		for pos := range penguinOccupied {
			if icebergOccupied[pos] {
				return true
			}
		}
	}

	return false
}

func (g *Game) render() {
	// Clear screen and move cursor to top-left
	fmt.Print("\x1b[H")

	// Create game buffer as rune arrays
	buffer := make([][]rune, GameHeight)
	for i := range buffer {
		buffer[i] = make([]rune, GameWidth)
		// Fill with spaces
		for j := range buffer[i] {
			buffer[i][j] = ' '
		}
		// Add side borders
		buffer[i][0] = '|'
		buffer[i][GameWidth-1] = '|'
	}

	// Draw icebergs sorted by bottom Y position (so lower ones draw on top)
	sortedIcebergs := make([]Iceberg, len(g.icebergs))
	copy(sortedIcebergs, g.icebergs)
	sort.Slice(sortedIcebergs, func(i, j int) bool {
		// Sort by bottom Y position (y + height) so items lower on screen draw last
		bottomI := sortedIcebergs[i].y + sortedIcebergs[i].height
		bottomJ := sortedIcebergs[j].y + sortedIcebergs[j].height
		return bottomI < bottomJ // Draw top first, then bottom (on top)
	})

	for _, iceberg := range sortedIcebergs {
		for dy, line := range iceberg.art {
			y := iceberg.y + dy
			if y >= 0 && y < GameHeight {
				for dx, ch := range line {
					x := iceberg.x + dx
					if x >= 1 && x < GameWidth-1 && ch != ' ' {
						buffer[y][x] = ch
					}
				}
			}
		}
	}

	// Draw penguin last (always on top)
	penguinLines := []string{
		"  __",
		"=(__)>",
	}
	for i, line := range penguinLines {
		y := g.penguinPos + i
		if y >= 0 && y < GameHeight {
			for dx, ch := range line {
				x := 3 + dx
				if x >= 1 && x < GameWidth-1 && ch != ' ' {
					buffer[y][x] = ch
				}
			}
		}
	}

	// Create top and bottom borders
	topBorder := make([]rune, GameWidth)
	bottomBorder := make([]rune, GameWidth)
	for i := 0; i < GameWidth; i++ {
		if i == 0 {
			topBorder[i] = '.'
			bottomBorder[i] = '\''
		} else if i == GameWidth-1 {
			topBorder[i] = '.'
			bottomBorder[i] = '\''
		} else {
			topBorder[i] = '-'
			bottomBorder[i] = '-'
		}
	}

	// Add score to top border
	scoreText := fmt.Sprintf(" Score: %d ", g.score)
	centerStart := (GameWidth - len(scoreText)) / 2
	for i, ch := range scoreText {
		if centerStart+i < GameWidth-1 {
			topBorder[centerStart+i] = ch
		}
	}

	// Draw bonus message in the buffer (bottom right corner) if active
	if g.bonusMessage != "" {
		msg := fmt.Sprintf(">>> %s! <<<", g.bonusMessage)
		msgY := GameHeight - 3 // One row above bottom border
		msgX := GameWidth - len(msg) // Right aligned, touching the border
		for i, ch := range msg {
			x := msgX + i
			if x >= 1 && x < GameWidth-1 {
				buffer[msgY][x] = ch
			}
		}
	}

	// Render everything (with \r for line start in raw mode)
	fmt.Printf("%s\r\n", string(topBorder))
	for _, row := range buffer {
		fmt.Printf("%s\r\n", string(row))
	}
	fmt.Printf("%s\r\n", string(bottomBorder))

	// Instructions
	fmt.Printf("\r\n  W/S: Move  Q: Quit")
}

func (g *Game) renderGameOver() {
	// Clear screen for game over display
	fmt.Print("\x1b[H")

	// Create game buffer as rune arrays
	buffer := make([][]rune, GameHeight)
	for i := range buffer {
		buffer[i] = make([]rune, GameWidth)
		// Fill with spaces
		for j := range buffer[i] {
			buffer[i][j] = ' '
		}
		// Add side borders
		buffer[i][0] = '|'
		buffer[i][GameWidth-1] = '|'
	}

	// Generate random state for penguin
	hunger := rand.Float64() * 100
	mood := rand.Float64() * 100
	energy := rand.Float64() * 100

	// Get penguin eye based on random state
	var eye string
	if energy < 30 {
		eye = "-" // Sleeping
	} else if hunger < 20 {
		eye = "o" // Very hungry
	} else if hunger >= 95 {
		eye = "*" // Excited
	} else if mood >= 80 {
		eye = "^" // Very happy
	} else if mood >= 50 {
		eye = "." // Neutral
	} else if mood >= 30 {
		eye = "-" // Sad
	} else {
		eye = ">" // Angry
	}

	// Get penguin art based on random energy
	var penguinArt []string
	if energy < 30 {
		// Lying down
		penguinArt = []string{
			"        ___",
			"      ,'   '-.__",
			"     /  --' )  " + eye + ")=-",
			"  --'--'-------'",
		}
	} else if mood >= 80 {
		// Wings spread (happy)
		penguinArt = []string{
			"    --.   __",
			"   (   \\.' " + eye + ")=-",
			"    `.  '-.-",
			"      ;-  |\\",
			"      |   |'",
			"    _,:__/_",
		}
	} else {
		// Standing
		penguinArt = []string{
			"          __",
			"        -' " + eye + ")=-",
			"       /.-.'",
			"      //  |\\",
			"      ||  |'",
			"    _,;(_/_",
		}
	}

	// Draw game over text and score
	textLines := []string{
		"  G A M E   O V E R  ",
		"                     ",
		fmt.Sprintf("    Score: %-8d    ", g.score),
		"                     ",
	}
	textLines = append(textLines, penguinArt...)

	startY := (GameHeight - len(textLines)) / 2

	// Find max line length for centering the whole block
	maxLen := 0
	for _, line := range textLines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}
	startX := (GameWidth - maxLen) / 2

	// Draw all lines with same starting X position
	for i, line := range textLines {
		y := startY + i
		if y >= 0 && y < GameHeight {
			for dx, ch := range line {
				x := startX + dx
				if x >= 1 && x < GameWidth-1 && ch != ' ' {
					buffer[y][x] = ch
				}
			}
		}
	}

	// Create top and bottom borders
	topBorder := make([]rune, GameWidth)
	bottomBorder := make([]rune, GameWidth)
	for i := 0; i < GameWidth; i++ {
		if i == 0 {
			topBorder[i] = '.'
			bottomBorder[i] = '\''
		} else if i == GameWidth-1 {
			topBorder[i] = '.'
			bottomBorder[i] = '\''
		} else {
			topBorder[i] = '-'
			bottomBorder[i] = '-'
		}
	}

	// Render everything with \r\n
	fmt.Printf("%s\r\n", string(topBorder))
	for _, row := range buffer {
		fmt.Printf("%s\r\n", string(row))
	}
	fmt.Printf("%s\r\n", string(bottomBorder))

	// Restore terminal (show cursor, exit raw mode) - NO CLEAR
	fmt.Print("\x1b[?25h")
	if g.oldState != nil {
		term.Restore(int(os.Stdin.Fd()), g.oldState)
	}
}
