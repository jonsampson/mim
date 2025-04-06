package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonsampson/mim/internal/infra"
	"github.com/jonsampson/mim/internal/tui"
)

func main() {
	// Open a log file for debugging
	logFile, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Set up the log package to write to the log file
	log.SetOutput(logFile)

	factory := infra.CollectorFactory{}
	collectors := factory.CreateCollectors()

	model, err := tui.InitialModel(collectors...)
	if err != nil {
		log.Printf("Error initializing model: %v\n", err)
		fmt.Printf("Error initializing model: %v\n", err)
		os.Exit(1)
	}

	// Initialize the Bubble Tea program (output remains on the terminal)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Printf("Alas, there's been an error: %v", err)
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
