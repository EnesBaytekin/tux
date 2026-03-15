// Tux - A terminal pet penguin.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/imns/tux/internal/ascii"
	"github.com/imns/tux/internal/client"
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

	resp, err := client.SendCommand(action)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Message)
		os.Exit(1)
	}

	if resp.State != nil {
		display(resp.State)
	}

	if resp.Message != "" {
		fmt.Println(resp.Message)
	}
}

// display shows Tux with its current state.
func display(state *client.State) {
	fmt.Println(ascii.DisplayWithStats("Tux", state.Hunger, state.Mood, state.Energy))
}
