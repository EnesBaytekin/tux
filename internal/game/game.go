package game

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
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
	penguinPos  int
	icebergs    []Iceberg
	score       int
	running     bool
	oldState    *term.State
	inputChan   chan rune
	cancelChan  chan struct{}
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
		penguinPos: 7,
		icebergs:   make([]Iceberg, 0),
		score:      0,
		running:    true,
		inputChan:  make(chan rune, 10),
		cancelChan: make(chan struct{}),
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
	ticker := time.NewTicker(60 * time.Millisecond)
	icebergTicker := time.NewTicker(1500 * time.Millisecond)

	defer ticker.Stop()
	defer icebergTicker.Stop()

	for g.running {
		select {
		case <-ticker.C:
			g.update()
			g.render()
		case <-icebergTicker.C:
			g.addIceberg()
		case key := <-g.inputChan:
			g.handleInput(key)
		case <-g.cancelChan:
			g.running = false
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
	// Clear screen and show cursor
	fmt.Print("\x1b[2J\x1b[H\x1b[?25h")
	if g.oldState != nil {
		term.Restore(int(os.Stdin.Fd()), g.oldState)
	}
}

func (g *Game) readInput() {
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			return
		}
		g.inputChan <- rune(buf[0])
	}
}

func (g *Game) handleInput(key rune) {
	switch key {
	case 'w', 'W', 65: // Up
		if g.penguinPos > 1 {
			g.penguinPos -= 1
		}
	case 's', 'S', 66: // Down
		if g.penguinPos < GameHeight-4 {
			g.penguinPos += 1
		}
	case 'q', 'Q', 27: // Quit
		g.Stop()
	}
}

func (g *Game) update() {
	// Move icebergs left
	for i := range g.icebergs {
		g.icebergs[i].x -= 1
	}

	// Remove off-screen icebergs and add score
	for i := len(g.icebergs) - 1; i >= 0; i-- {
		if g.icebergs[i].x < -10 {
			g.icebergs = append(g.icebergs[:i], g.icebergs[i+1:]...)
			g.score += 10
		}
	}

	// Check collision
	if g.checkCollision() {
		g.running = false
	}

	// Increment score for survival
	g.score += 1
}

func (g *Game) addIceberg() {
	// Pick random iceberg art
	artVariants := icebergArts[rand.Intn(len(icebergArts))]
	art := artVariants[0]

	// Random position on right side
	x := GameWidth - 1
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
	// Penguin actual occupied positions (non-space chars)
	penguinX := 3
	penguinY := g.penguinPos
	penguinLines := []string{
		"  __",
		"=(__)>",
	}

	// Create a set of occupied positions for penguin
	penguinOccupied := make(map[[2]int]bool)
	for dy, line := range penguinLines {
		for dx, ch := range line {
			if ch != ' ' {
				penguinOccupied[[2]int{penguinX + dx, penguinY + dy}] = true
			}
		}
	}

	// Check each iceberg
	for _, iceberg := range g.icebergs {
		// Create a set of occupied positions for this iceberg
		icebergOccupied := make(map[[2]int]bool)
		for dy, line := range iceberg.art {
			for dx, ch := range line {
				if ch != ' ' {
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

	// Draw icebergs
	for _, iceberg := range g.icebergs {
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

	// Draw penguin
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

	// Render everything (with \r for line start in raw mode)
	fmt.Printf("%s\r\n", string(topBorder))
	for _, row := range buffer {
		fmt.Printf("%s\r\n", string(row))
	}
	fmt.Printf("%s\r\n", string(bottomBorder))

	// Instructions
	fmt.Printf("\r\n  W/S or ↑/↓: Move  Q: Quit")
}

func (g *Game) renderGameOver() {
	// Clear screen and move cursor to top-left
	fmt.Print("\x1b[2J\x1b[H")

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

	// Draw simple "GAME OVER" text
	textLines := []string{
		"  G A M E   O V E R  ",
		"                     ",
		fmt.Sprintf("    Score: %-8d    ", g.score),
		"                     ",
		"        ___          ",
		"       /  .\\         ",
		"      |  O O |       ",
		"       \\  < /        ",
		"        \\___/        ",
		"         |||         ",
	}

	startY := (GameHeight - len(textLines)) / 2
	for i, line := range textLines {
		y := startY + i
		if y >= 0 && y < GameHeight {
			startX := (GameWidth - len(line)) / 2
			for dx, ch := range line {
				x := startX + dx
				if x >= 1 && x < GameWidth-1 {
					buffer[y][x] = ch
				}
			}
		}
	}

	// Draw "Press Q to exit" at bottom
	pressLine := "Press Q to exit"
	startX := (GameWidth - len(pressLine)) / 2
	for dx, ch := range pressLine {
		x := startX + dx
		if x >= 1 && x < GameWidth-1 {
			buffer[GameHeight-2][x] = ch
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

	// Wait for Q key to exit
	buf := make([]byte, 1)
	for {
		os.Stdin.Read(buf)
		if buf[0] == 'q' || buf[0] == 'Q' || buf[0] == 27 {
			break
		}
	}
}
