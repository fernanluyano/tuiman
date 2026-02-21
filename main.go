package main

import (
	"fmt"
	"os"

	"github.com/fberlanga/tuiman/internal/ui"
)

func main() {
	app := ui.NewApp()
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
