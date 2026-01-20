package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hitsumabushi845/task-management/internal/app"
	"github.com/hitsumabushi845/task-management/internal/repository"
)

func main() {
	// Get user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create data directory
	dataDir := filepath.Join(home, ".task-management")
	dbPath := filepath.Join(dataDir, "tasks.db")

	// Create repository
	repo, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating repository: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Create and run application
	model := app.New(repo)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
