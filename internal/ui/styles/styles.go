package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Task status colors
	StatusNew       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	StatusWorking   = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	StatusCompleted = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))

	// Priority colors
	PriorityHigh   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	PriorityMedium = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	PriorityLow    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// UI elements
	Selected = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	Normal   = lipgloss.NewStyle()

	// Status bar
	StatusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)
)
