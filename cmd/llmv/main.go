package main

import (
	"flag"
	"github.com/aj-seven/llmverse/internal/config"
	"github.com/aj-seven/llmverse/internal/history"
	aihub "github.com/aj-seven/llmverse/internal/providers/ollama"
	"github.com/aj-seven/llmverse/internal/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var host string
	flag.StringVar(&host, "host", "", "Ollama host address")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadOrNew(host)
	if err != nil {
		exit(err)
	}

	// Initialize history storage
	fileStorage, err := history.NewFileStorage(cfg)
	if err != nil {
		exit(err)
	}

	// Initialize history manager
	historyManager := history.NewManager(fileStorage)
	defer historyManager.Close()

	// Load available models
	models, err := aihub.GetModelsDetailed(cfg)
	if err != nil {
		exit(err)
	}
	if len(models) == 0 {
		exitString("no Ollama models found (is Ollama running?)")
	}

	// Create application model
	m := ui.New(models, historyManager, cfg)

	// Start Bubble Tea program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if err := p.Start(); err != nil {
		exit(err)
	}
}

// Helper functions

func exit(err error) {
	os.Stderr.WriteString(err.Error() + "\n")
	os.Exit(1)
}

func exitString(msg string) {
	os.Stderr.WriteString(msg + "\n")
	os.Exit(1)
}
