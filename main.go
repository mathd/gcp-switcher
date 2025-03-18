package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mathieudupuis/gcp-switcher/internal"
	"github.com/mathieudupuis/gcp-switcher/ui"
)

const (
	logFilePath = "gcp-switcher.log"
)

var (
	debugMode bool
	logger    *log.Logger
)

// Init logger
func initLogger() {
	if !debugMode {
		logger = log.New(io.Discard, "", 0)
		return
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}

	logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	logger.Println("=== GCP Switcher Started ===")
}

func main() {
	// Parse command line flags
	flag.BoolVar(&debugMode, "debug", false, "Enable debug logging")
	flag.Parse()

	// Initialize logger
	initLogger()

	logger.Println("Starting GCP Switcher application")

	// Initialize styles
	styles := ui.NewStyles()

	// Create and start the program
	p := tea.NewProgram(internal.InitialModel(styles), tea.WithAltScreen())

	// Start the program
	if _, err := p.Run(); err != nil {
		logger.Printf("Error running program: %v", err)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
