package app

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hitsumabushi845/task-management/internal/domain"
	"github.com/hitsumabushi845/task-management/internal/ui/styles"
)

type viewMode int

const (
	viewModeList viewMode = iota
	viewModeCreate
	viewModeKanban
	viewModeFilter
	viewModeHelp
)

// Model is the root application model
type Model struct {
	repo          domain.TaskRepository
	tasks         []*domain.Task
	cursor        int
	width         int
	height        int
	err           error
	mode          viewMode
	previousMode  viewMode
	inputTitle    string
	inputPriority domain.Priority
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

// createTask creates a new task
func (m *Model) createTask(title string, priority domain.Priority) tea.Cmd {
	return func() tea.Msg {
		task := &domain.Task{
			Title:    title,
			Status:   domain.TaskStatusNew,
			Priority: priority,
		}

		err := m.repo.Create(context.Background(), task)
		if err != nil {
			return errMsg{err: err}
		}

		return taskCreatedMsg{task: task}
	}
}

// deleteTask deletes the selected task
func (m *Model) deleteTask(id int64) tea.Cmd {
	return func() tea.Msg {
		err := m.repo.Delete(context.Background(), id)
		if err != nil {
			return errMsg{err: err}
		}
		return taskDeletedMsg{id: id}
	}
}

// toggleTaskStatus toggles the task status: new -> working -> completed -> new
func (m *Model) toggleTaskStatus(task *domain.Task) tea.Cmd {
	return func() tea.Msg {
		// Update status and timestamps
		now := time.Now()
		switch task.Status {
		case domain.TaskStatusNew:
			task.Status = domain.TaskStatusWorking
			task.StartedAt = &now
		case domain.TaskStatusWorking:
			task.Status = domain.TaskStatusCompleted
			task.CompletedAt = &now
		case domain.TaskStatusCompleted:
			task.Status = domain.TaskStatusNew
			task.StartedAt = nil
			task.CompletedAt = nil
		}

		err := m.repo.Update(context.Background(), task)
		if err != nil {
			return errMsg{err: err}
		}

		return taskUpdatedMsg{task: task}
	}
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle create mode separately
		if m.mode == viewModeCreate {
			return m.updateCreateMode(msg)
		}

		// Handle help mode
		if m.mode == viewModeHelp {
			switch msg.String() {
			case "esc", "?", "f1":
				m.mode = m.previousMode
			}
			return m, nil
		}

		// List mode key handlers
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "j", "down":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "n":
			// Enter create mode
			m.mode = viewModeCreate
			m.inputTitle = ""
			m.inputPriority = domain.PriorityMedium

		case "d":
			// Delete selected task
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
				task := m.tasks[m.cursor]
				return m, m.deleteTask(task.ID)
			}

		case " ":
			// Toggle task status
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
				task := m.tasks[m.cursor]
				return m, m.toggleTaskStatus(task)
			}

		case "v":
			// Toggle between list and kanban view
			if m.mode == viewModeList {
				m.mode = viewModeKanban
			} else if m.mode == viewModeKanban {
				m.mode = viewModeList
			}

		case "?", "f1":
			// Show help modal
			m.previousMode = m.mode
			m.mode = viewModeHelp
		}

	case taskListLoadedMsg:
		m.tasks = msg.tasks
		if m.cursor >= len(m.tasks) {
			m.cursor = len(m.tasks) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}

	case taskCreatedMsg:
		// Task created, reload list
		return m, m.loadTasks()

	case taskDeletedMsg:
		// Task deleted, reload list
		return m, m.loadTasks()

	case taskUpdatedMsg:
		// Task updated, reload list
		return m, m.loadTasks()

	case errMsg:
		m.err = msg.err

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// updateCreateMode handles input in create mode
func (m *Model) updateCreateMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel creation
		m.mode = viewModeList

	case "enter":
		// Create task
		if m.inputTitle != "" {
			m.mode = viewModeList
			return m, m.createTask(m.inputTitle, m.inputPriority)
		}

	case "backspace":
		if len(m.inputTitle) > 0 {
			m.inputTitle = m.inputTitle[:len(m.inputTitle)-1]
		}

	case "tab":
		// Cycle priority
		switch m.inputPriority {
		case domain.PriorityLow:
			m.inputPriority = domain.PriorityMedium
		case domain.PriorityMedium:
			m.inputPriority = domain.PriorityHigh
		case domain.PriorityHigh:
			m.inputPriority = domain.PriorityLow
		}

	default:
		// Add character to title
		if len(msg.String()) == 1 {
			m.inputTitle += msg.String()
		} else if msg.Type == tea.KeySpace {
			m.inputTitle += " "
		} else if msg.Type == tea.KeyRunes {
			m.inputTitle += string(msg.Runes)
		}
	}

	return m, nil
}

// View renders the application
func (m *Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n\nPress q to quit.\n"
	}

	// Help modal overlay
	if m.mode == viewModeHelp {
		return m.viewHelp()
	}

	// Create mode view
	if m.mode == viewModeCreate {
		return m.viewCreate()
	}

	// Kanban mode view
	if m.mode == viewModeKanban {
		return m.viewKanban()
	}

	// List mode view
	return m.viewList()
}

func (m *Model) viewList() string {
	s := "Task Management\n\n"

	if len(m.tasks) == 0 {
		s += "No tasks yet. Press 'n' to create one.\n\n"
	} else {
		for i, task := range m.tasks {
			// Status icon
			var statusIcon string
			var statusStyle lipgloss.Style
			switch task.Status {
			case domain.TaskStatusNew:
				statusIcon = "○"
				statusStyle = styles.StatusNew
			case domain.TaskStatusWorking:
				statusIcon = "●"
				statusStyle = styles.StatusWorking
			case domain.TaskStatusCompleted:
				statusIcon = "✓"
				statusStyle = styles.StatusCompleted
			}

			// Priority indicator
			var priorityStyle lipgloss.Style
			var priorityText string
			switch task.Priority {
			case domain.PriorityHigh:
				priorityStyle = styles.PriorityHigh
				priorityText = "高"
			case domain.PriorityMedium:
				priorityStyle = styles.PriorityMedium
				priorityText = "中"
			case domain.PriorityLow:
				priorityStyle = styles.PriorityLow
				priorityText = "低"
			}

			// Build task line
			line := fmt.Sprintf("%s [%s] %s",
				statusStyle.Render(statusIcon),
				priorityStyle.Render(priorityText),
				task.Title,
			)

			// Highlight selected
			if i == m.cursor {
				line = styles.Selected.Render("> " + line)
			} else {
				line = "  " + line
			}

			s += line + "\n"
		}
		s += "\n"
	}

	// Status bar
	helpText := "[n]新規 [d]削除 [Space]ステータス [↑/k]上 [↓/j]下 [q]終了"
	s += styles.StatusBar.Render(helpText) + "\n"

	return s
}

func (m *Model) viewCreate() string {
	s := "新規タスク作成\n\n"

	s += "タイトル: " + m.inputTitle + "█\n\n"

	// Priority selection
	s += "優先度 (Tabで切替): "
	switch m.inputPriority {
	case domain.PriorityHigh:
		s += styles.PriorityHigh.Render("[高]") + " 中 低"
	case domain.PriorityMedium:
		s += "高 " + styles.PriorityMedium.Render("[中]") + " 低"
	case domain.PriorityLow:
		s += "高 中 " + styles.PriorityLow.Render("[低]")
	}
	s += "\n\n"

	helpText := "[Enter]作成 [Esc]キャンセル [Tab]優先度"
	s += styles.StatusBar.Render(helpText) + "\n"

	return s
}

func (m *Model) viewKanban() string {
	return "Kanban view (coming soon)\n\nPress 'v' to switch to list view, '?' for help, 'q' to quit.\n"
}

func (m *Model) viewHelp() string {
	var s string

	if m.previousMode == viewModeKanban {
		s = `┌─ ヘルプ - カンバンビュー ───────────────┐
│                                        │
│ 移動:                                  │
│   h/←      : 左の列へ                  │
│   l/→      : 右の列へ                  │
│   j/↓      : 列内で下へ                │
│   k/↑      : 列内で上へ                │
│                                        │
│ タスク操作:                            │
│   Enter    : 次のステータスへ移動      │
│   e        : タスク編集                │
│   n        : 新規タスク作成            │
│   d        : タスク削除                │
│                                        │
│ 表示:                                  │
│   v        : リストビューへ切替        │
│   f        : フィルタ設定              │
│   s        : ソート設定                │
│   ?/F1     : このヘルプ                │
│   q        : 終了                      │
│                                        │
│         [Esc または ? で閉じる]        │
└────────────────────────────────────────┘`
	} else {
		s = `┌─ ヘルプ - リストビュー ─────────────────┐
│                                        │
│ 移動:                                  │
│   j/↓      : 下へ移動                  │
│   k/↑      : 上へ移動                  │
│                                        │
│ タスク操作:                            │
│   Space    : ステータス切替            │
│   n        : 新規タスク作成            │
│   d        : タスク削除                │
│                                        │
│ 表示:                                  │
│   v        : カンバンビューへ切替      │
│   f        : フィルタ設定              │
│   s        : ソート設定                │
│   ?/F1     : このヘルプ                │
│   q        : 終了                      │
│                                        │
│         [Esc または ? で閉じる]        │
└────────────────────────────────────────┘`
	}

	return s
}
