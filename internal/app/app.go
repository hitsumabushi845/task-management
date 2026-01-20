package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hitsumabushi845/task-management/internal/domain"
)

// Model is the root application model
type Model struct {
	repo   domain.TaskRepository
	tasks  []*domain.Task
	cursor int
	width  int
	height int
	err    error
}

// New creates a new application model
func New(repo domain.TaskRepository) *Model {
	return &Model{
		repo:  repo,
		tasks: []*domain.Task{},
	}
}

// Init initializes the application
func (m *Model) Init() tea.Cmd {
	return m.loadTasks()
}

// loadTasks loads all tasks from the repository
func (m *Model) loadTasks() tea.Cmd {
	return func() tea.Msg {
		tasks, err := m.repo.List(context.Background())
		if err != nil {
			return errMsg{err: err}
		}
		return taskListLoadedMsg{tasks: tasks}
	}
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case taskListLoadedMsg:
		m.tasks = msg.tasks

	case errMsg:
		m.err = msg.err

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the application
func (m *Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n\nPress q to quit.\n"
	}

	s := "Task Management\n\n"

	if len(m.tasks) == 0 {
		s += "No tasks yet.\n\n"
	} else {
		s += "Tasks:\n"
		for _, task := range m.tasks {
			s += "- " + task.Title + "\n"
		}
		s += "\n"
	}

	s += "Press q to quit.\n"

	return s
}
