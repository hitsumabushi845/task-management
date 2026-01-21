package app

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

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
	viewModeEdit
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
	inputTitle       string
	inputPriority    domain.Priority
	inputCategoryIdx int // Index into categories slice, -1 for no category
	// Kanban view state
	kanbanColumn  int    // 0=New, 1=Working, 2=Completed
	kanbanCursors [3]int // Cursor position within each column
	// Filter state
	filter       domain.Filter
	filterCursor int
	// Sort state
	taskSort     domain.Sort
	sortMenuOpen bool
	// Edit state
	editTask        *domain.Task    // Reference to task being edited
	editCursor      int             // 0=title, 1=desc, 2=priority, 3=category, 4=date, 5=save, 6=cancel
	editingField    bool            // Currently typing in a field
	editTitle       string          // Edited title value
	editDesc        string          // Edited description value
	editPriority    domain.Priority
	editCategoryIdx int    // Index into categories slice, -1 for no category
	editDueDate     string // String for input, parsed on save
	editError       string // Validation error message
	// Category state
	categories []*domain.Category // All available categories
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
	return tea.Batch(m.loadTasks(), m.loadCategories())
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

// loadCategories loads all categories from the repository
func (m *Model) loadCategories() tea.Cmd {
	return func() tea.Msg {
		categories, err := m.repo.GetCategories(context.Background())
		if err != nil {
			return errMsg{err: err}
		}
		return categoriesLoadedMsg{categories: categories}
	}
}

// createTask creates a new task
func (m *Model) createTask(title string, priority domain.Priority, categoryIdx int) tea.Cmd {
	return func() tea.Msg {
		task := &domain.Task{
			Title:    title,
			Status:   domain.TaskStatusNew,
			Priority: priority,
		}

		// Set category if selected
		if categoryIdx >= 0 && categoryIdx < len(m.categories) {
			catID := m.categories[categoryIdx].ID
			task.CategoryID = &catID
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

func (m *Model) tasksByStatus(status domain.TaskStatus) []*domain.Task {
	// Apply filter first
	filtered := m.filter.Apply(m.tasks)
	// Apply sort
	sorted := m.taskSort.Apply(filtered)
	var result []*domain.Task
	for _, task := range sorted {
		if task.Status == status {
			result = append(result, task)
		}
	}
	return result
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

		// Handle filter mode
		if m.mode == viewModeFilter {
			return m.updateFilterMode(msg)
		}

		// Handle edit mode
		if m.mode == viewModeEdit {
			return m.updateEditMode(msg)
		}

		// Handle kanban mode
		if m.mode == viewModeKanban {
			return m.updateKanbanMode(msg)
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
			m.inputCategoryIdx = -1 // No category selected by default

		case "d":
			// Delete selected task
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
				task := m.tasks[m.cursor]
				return m, m.deleteTask(task.ID)
			}

		case "e":
			// Edit selected task
			filteredTasks := m.filter.Apply(m.tasks)
			sortedTasks := m.taskSort.Apply(filteredTasks)
			if len(sortedTasks) > 0 && m.cursor < len(sortedTasks) {
				task := sortedTasks[m.cursor]
				m.startEditMode(task)
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

		case "f":
			// Open filter modal
			m.previousMode = m.mode
			m.mode = viewModeFilter
			m.filterCursor = 0

		case "s":
			// Toggle sort menu
			m.sortMenuOpen = !m.sortMenuOpen

		case "1", "2", "3", "4", "5":
			if m.sortMenuOpen {
				idx := int(msg.String()[0] - '1')
				sortOptions := []domain.SortBy{
					domain.SortByCreatedAt,
					domain.SortByDueDate,
					domain.SortByPriority,
					domain.SortByStatus,
					domain.SortByTitle,
				}
				if idx >= 0 && idx < len(sortOptions) {
					if m.taskSort.By == sortOptions[idx] {
						m.taskSort.Ascending = !m.taskSort.Ascending
					} else {
						m.taskSort.By = sortOptions[idx]
						m.taskSort.Ascending = false
					}
				}
				m.sortMenuOpen = false
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

	case categoriesLoadedMsg:
		m.categories = msg.categories

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
			return m, m.createTask(m.inputTitle, m.inputPriority, m.inputCategoryIdx)
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

	case "shift+tab":
		// Cycle category
		if len(m.categories) > 0 {
			m.inputCategoryIdx++
			if m.inputCategoryIdx >= len(m.categories) {
				m.inputCategoryIdx = -1 // Back to "None"
			}
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

// updateKanbanMode handles input in kanban mode
func (m *Model) updateKanbanMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	columns := [3][]*domain.Task{
		m.tasksByStatus(domain.TaskStatusNew),
		m.tasksByStatus(domain.TaskStatusWorking),
		m.tasksByStatus(domain.TaskStatusCompleted),
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "h", "left":
		if m.kanbanColumn > 0 {
			m.kanbanColumn--
			// Adjust cursor if needed
			if m.kanbanCursors[m.kanbanColumn] >= len(columns[m.kanbanColumn]) {
				m.kanbanCursors[m.kanbanColumn] = len(columns[m.kanbanColumn]) - 1
			}
			if m.kanbanCursors[m.kanbanColumn] < 0 {
				m.kanbanCursors[m.kanbanColumn] = 0
			}
		}

	case "l", "right":
		if m.kanbanColumn < 2 {
			m.kanbanColumn++
			// Adjust cursor if needed
			if m.kanbanCursors[m.kanbanColumn] >= len(columns[m.kanbanColumn]) {
				m.kanbanCursors[m.kanbanColumn] = len(columns[m.kanbanColumn]) - 1
			}
			if m.kanbanCursors[m.kanbanColumn] < 0 {
				m.kanbanCursors[m.kanbanColumn] = 0
			}
		}

	case "j", "down":
		col := m.kanbanColumn
		if m.kanbanCursors[col] < len(columns[col])-1 {
			m.kanbanCursors[col]++
		}

	case "k", "up":
		col := m.kanbanColumn
		if m.kanbanCursors[col] > 0 {
			m.kanbanCursors[col]--
		}

	case "enter":
		// Move task to next status
		col := m.kanbanColumn
		if len(columns[col]) > 0 && m.kanbanCursors[col] < len(columns[col]) {
			task := columns[col][m.kanbanCursors[col]]
			return m, m.advanceTaskStatus(task)
		}

	case "n":
		m.mode = viewModeCreate
		m.inputTitle = ""
		m.inputPriority = domain.PriorityMedium
		m.inputCategoryIdx = -1 // No category selected by default

	case "d":
		col := m.kanbanColumn
		if len(columns[col]) > 0 && m.kanbanCursors[col] < len(columns[col]) {
			task := columns[col][m.kanbanCursors[col]]
			return m, m.deleteTask(task.ID)
		}

	case "e":
		// Edit selected task
		col := m.kanbanColumn
		if len(columns[col]) > 0 && m.kanbanCursors[col] < len(columns[col]) {
			task := columns[col][m.kanbanCursors[col]]
			m.startEditMode(task)
		}

	case "v":
		m.mode = viewModeList

	case "f":
		m.previousMode = m.mode
		m.mode = viewModeFilter
		m.filterCursor = 0

	case "s":
		// Toggle sort menu
		m.sortMenuOpen = !m.sortMenuOpen

	case "1", "2", "3", "4", "5":
		if m.sortMenuOpen {
			idx := int(msg.String()[0] - '1')
			sortOptions := []domain.SortBy{
				domain.SortByCreatedAt,
				domain.SortByDueDate,
				domain.SortByPriority,
				domain.SortByStatus,
				domain.SortByTitle,
			}
			if idx >= 0 && idx < len(sortOptions) {
				if m.taskSort.By == sortOptions[idx] {
					m.taskSort.Ascending = !m.taskSort.Ascending
				} else {
					m.taskSort.By = sortOptions[idx]
					m.taskSort.Ascending = false
				}
			}
			m.sortMenuOpen = false
		}

	case "?", "f1":
		m.previousMode = m.mode
		m.mode = viewModeHelp
	}

	return m, nil
}

// advanceTaskStatus moves task to next status (new -> working -> completed)
func (m *Model) advanceTaskStatus(task *domain.Task) tea.Cmd {
	return func() tea.Msg {
		now := time.Now()
		switch task.Status {
		case domain.TaskStatusNew:
			task.Status = domain.TaskStatusWorking
			task.StartedAt = &now
		case domain.TaskStatusWorking:
			task.Status = domain.TaskStatusCompleted
			task.CompletedAt = &now
		case domain.TaskStatusCompleted:
			// Already completed, no change
			return nil
		}

		err := m.repo.Update(context.Background(), task)
		if err != nil {
			return errMsg{err: err}
		}

		return taskUpdatedMsg{task: task}
	}
}

// startEditMode initializes edit mode with the given task
func (m *Model) startEditMode(task *domain.Task) {
	m.previousMode = m.mode // Save current mode to return to after edit
	m.editTask = task
	m.editCursor = 0
	m.editingField = false
	m.editTitle = task.Title
	m.editDesc = task.Description
	m.editPriority = task.Priority
	// Find category index
	m.editCategoryIdx = -1
	if task.CategoryID != nil {
		for i, cat := range m.categories {
			if cat.ID == *task.CategoryID {
				m.editCategoryIdx = i
				break
			}
		}
	}
	if task.DueDate != nil {
		m.editDueDate = task.DueDate.Format("2006-01-02")
	} else {
		m.editDueDate = ""
	}
	m.editError = ""
	m.mode = viewModeEdit
}

// updateEditMode handles input in edit mode
func (m *Model) updateEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If editing a text field, handle text input
	if m.editingField {
		return m.updateEditFieldInput(msg)
	}

	// Navigation mode
	switch msg.String() {
	case "j", "down":
		if m.editCursor < 6 {
			m.editCursor++
		}

	case "k", "up":
		if m.editCursor > 0 {
			m.editCursor--
		}

	case "enter":
		switch m.editCursor {
		case 0, 1, 4: // Title, Description, Due Date (due date moved to 4)
			// Start editing text field
			m.editingField = true
		case 2: // Priority
			m.cyclePriority()
		case 3: // Category
			m.cycleCategory()
		case 5: // Save button
			return m.saveEditedTask()
		case 6: // Cancel button
			m.mode = m.previousMode
			m.editError = ""
		}

	case "tab":
		// Cycle priority or category depending on cursor
		if m.editCursor == 2 {
			m.cyclePriority()
		} else if m.editCursor == 3 {
			m.cycleCategory()
		}

	case "esc":
		// Cancel edit
		m.mode = m.previousMode
		m.editError = ""
	}

	return m, nil
}

func (m *Model) cyclePriority() {
	switch m.editPriority {
	case domain.PriorityLow:
		m.editPriority = domain.PriorityMedium
	case domain.PriorityMedium:
		m.editPriority = domain.PriorityHigh
	case domain.PriorityHigh:
		m.editPriority = domain.PriorityLow
	}
}

func (m *Model) cycleCategory() {
	if len(m.categories) == 0 {
		return
	}
	m.editCategoryIdx++
	if m.editCategoryIdx >= len(m.categories) {
		m.editCategoryIdx = -1
	}
}

// updateEditFieldInput handles text input when editing a field
func (m *Model) updateEditFieldInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		// Stop editing this field
		m.editingField = false

	case "backspace":
		// Delete character
		switch m.editCursor {
		case 0: // Title
			if len(m.editTitle) > 0 {
				runes := []rune(m.editTitle)
				m.editTitle = string(runes[:len(runes)-1])
			}
		case 1: // Description
			if len(m.editDesc) > 0 {
				runes := []rune(m.editDesc)
				m.editDesc = string(runes[:len(runes)-1])
			}
		case 4: // Due Date (moved to position 4)
			if len(m.editDueDate) > 0 {
				m.editDueDate = m.editDueDate[:len(m.editDueDate)-1]
			}
		}

	default:
		// Add character
		var char string
		if len(msg.String()) == 1 {
			char = msg.String()
		} else if msg.Type == tea.KeySpace {
			char = " "
		} else if msg.Type == tea.KeyRunes {
			char = string(msg.Runes)
		}

		if char != "" {
			switch m.editCursor {
			case 0: // Title
				m.editTitle += char
			case 1: // Description
				m.editDesc += char
			case 4: // Due Date (moved to position 4)
				// Only allow date-like characters
				if len(m.editDueDate) < 10 {
					m.editDueDate += char
				}
			}
		}
	}

	return m, nil
}

// saveEditedTask validates and saves the edited task
func (m *Model) saveEditedTask() (tea.Model, tea.Cmd) {
	// Validate title
	if strings.TrimSpace(m.editTitle) == "" {
		m.editError = "Title is required"
		return m, nil
	}

	// Validate and parse due date
	var dueDate *time.Time
	if m.editDueDate != "" {
		parsed, err := time.Parse("2006-01-02", m.editDueDate)
		if err != nil {
			m.editError = "Invalid date format (use YYYY-MM-DD)"
			return m, nil
		}
		dueDate = &parsed
	}

	// Update task
	m.editTask.Title = strings.TrimSpace(m.editTitle)
	m.editTask.Description = m.editDesc
	m.editTask.Priority = m.editPriority
	m.editTask.DueDate = dueDate

	// Update category
	if m.editCategoryIdx >= 0 && m.editCategoryIdx < len(m.categories) {
		catID := m.categories[m.editCategoryIdx].ID
		m.editTask.CategoryID = &catID
	} else {
		m.editTask.CategoryID = nil
	}

	// Validate using domain validation
	if err := m.editTask.Validate(); err != nil {
		m.editError = err.Error()
		return m, nil
	}

	// Save to repository
	m.mode = m.previousMode
	m.editError = ""
	return m, m.updateTask(m.editTask)
}

// updateTask updates a task in the repository
func (m *Model) updateTask(task *domain.Task) tea.Cmd {
	return func() tea.Msg {
		err := m.repo.Update(context.Background(), task)
		if err != nil {
			return errMsg{err: err}
		}
		return taskUpdatedMsg{task: task}
	}
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

	// Filter mode view
	if m.mode == viewModeFilter {
		return m.viewFilter()
	}

	// Edit mode view
	if m.mode == viewModeEdit {
		return m.viewEdit()
	}

	// Kanban mode view
	if m.mode == viewModeKanban {
		return m.viewKanban()
	}

	// List mode view
	return m.viewList()
}

func (m *Model) viewList() string {
	s := "Task Management"
	if !m.filter.IsEmpty() {
		s += " (Filter Active)"
	}
	s += "\n\n"

	// Show sort menu if open
	if m.sortMenuOpen {
		s += m.viewSortMenu() + "\n"
	}

	// Apply filter to tasks
	filteredTasks := m.filter.Apply(m.tasks)
	// Apply sort after filter
	sortedTasks := m.taskSort.Apply(filteredTasks)

	if len(sortedTasks) == 0 {
		if len(m.tasks) == 0 {
			s += "No tasks yet. Press 'n' to create one.\n\n"
		} else {
			s += "No tasks match the filter.\n\n"
		}
	} else {
		for i, task := range sortedTasks {
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
				priorityText = "H"
			case domain.PriorityMedium:
				priorityStyle = styles.PriorityMedium
				priorityText = "M"
			case domain.PriorityLow:
				priorityStyle = styles.PriorityLow
				priorityText = "L"
			}

			// Build task line
			catName := m.getCategoryName(task)
			catDisplay := ""
			if catName != "" {
				catDisplay = " @" + catName
			}

			line := fmt.Sprintf("%s [%s] %s%s",
				statusStyle.Render(statusIcon),
				priorityStyle.Render(priorityText),
				task.Title,
				catDisplay,
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
	helpText := "[n]New [e]Edit [d]Delete [Space]Status [f]Filter [s]Sort [v]Kanban [?]Help [q]Quit"
	s += styles.StatusBar.Render(helpText) + "\n"

	return s
}

func (m *Model) viewCreate() string {
	s := "Create New Task\n\n"

	s += "Title: " + m.inputTitle + "█\n\n"

	// Priority selection
	s += "Priority (Tab to cycle): "
	switch m.inputPriority {
	case domain.PriorityHigh:
		s += styles.PriorityHigh.Render("[H]") + " M L"
	case domain.PriorityMedium:
		s += "H " + styles.PriorityMedium.Render("[M]") + " L"
	case domain.PriorityLow:
		s += "H M " + styles.PriorityLow.Render("[L]")
	}
	s += "\n\n"

	// Category selection
	s += "Category (Shift+Tab to cycle): "
	if m.inputCategoryIdx < 0 {
		s += "[None]"
	} else if m.inputCategoryIdx < len(m.categories) {
		cat := m.categories[m.inputCategoryIdx]
		s += "[" + cat.Name + "]"
	}
	s += "\n\n"

	helpText := "[Enter]Create [Esc]Cancel [Tab]Priority [Shift+Tab]Category"
	s += styles.StatusBar.Render(helpText) + "\n"

	return s
}

func (m *Model) viewKanban() string {
	newTasks := m.tasksByStatus(domain.TaskStatusNew)
	workingTasks := m.tasksByStatus(domain.TaskStatusWorking)
	completedTasks := m.tasksByStatus(domain.TaskStatusCompleted)

	// Filter indicator
	var s string
	if !m.filter.IsEmpty() {
		s = "(Filter Active)\n"
	}

	// Show sort menu if open
	if m.sortMenuOpen {
		s += m.viewSortMenu() + "\n"
	}

	// Column width
	colWidth := 25

	// Header - use max(0, ...) to prevent negative repeat counts
	newPadding := colWidth - 10 - len(fmt.Sprintf("%d", len(newTasks)))
	if newPadding < 0 {
		newPadding = 0
	}
	workingPadding := colWidth - 13 - len(fmt.Sprintf("%d", len(workingTasks)))
	if workingPadding < 0 {
		workingPadding = 0
	}
	completedPadding := colWidth - 15 - len(fmt.Sprintf("%d", len(completedTasks)))
	if completedPadding < 0 {
		completedPadding = 0
	}
	s += fmt.Sprintf("┌─ New (%d) %s┬─ Working (%d) %s┬─ Completed (%d) %s┐\n",
		len(newTasks), strings.Repeat("─", newPadding),
		len(workingTasks), strings.Repeat("─", workingPadding),
		len(completedTasks), strings.Repeat("─", completedPadding),
	)

	// Find max rows
	maxRows := len(newTasks)
	if len(workingTasks) > maxRows {
		maxRows = len(workingTasks)
	}
	if len(completedTasks) > maxRows {
		maxRows = len(completedTasks)
	}
	if maxRows == 0 {
		maxRows = 1
	}

	// Render rows
	for i := 0; i < maxRows; i++ {
		s += "│"
		s += m.renderKanbanCell(newTasks, i, 0, colWidth)
		s += "│"
		s += m.renderKanbanCell(workingTasks, i, 1, colWidth)
		s += "│"
		s += m.renderKanbanCell(completedTasks, i, 2, colWidth)
		s += "│\n"
	}

	// Footer
	s += fmt.Sprintf("└%s┴%s┴%s┘\n",
		strings.Repeat("─", colWidth),
		strings.Repeat("─", colWidth),
		strings.Repeat("─", colWidth),
	)

	// Status bar
	helpText := "[h/l]Column [j/k]Up/Down [Enter]Advance [e]Edit [f]Filter [s]Sort [v]List [?]Help [q]Quit"
	s += "\n" + styles.StatusBar.Render(helpText) + "\n"

	return s
}

func (m *Model) renderKanbanCell(tasks []*domain.Task, row, col, width int) string {
	if row >= len(tasks) {
		return strings.Repeat(" ", width)
	}

	task := tasks[row]
	isSelected := m.kanbanColumn == col && m.kanbanCursors[col] == row

	// Priority indicator
	var priorityText string
	var priorityStyle lipgloss.Style
	switch task.Priority {
	case domain.PriorityHigh:
		priorityStyle = styles.PriorityHigh
		priorityText = "H"
	case domain.PriorityMedium:
		priorityStyle = styles.PriorityMedium
		priorityText = "M"
	case domain.PriorityLow:
		priorityStyle = styles.PriorityLow
		priorityText = "L"
	}

	// Category
	catName := m.getCategoryName(task)
	catDisplay := ""
	if catName != "" {
		// Truncate category name if needed (use rune count for Unicode support)
		if utf8.RuneCountInString(catName) > 8 {
			runes := []rune(catName)
			catName = string(runes[:7]) + "."
		}
		catDisplay = "@" + catName
	}

	// Truncate title if needed
	title := task.Title
	// Account for priority [P] + space + category
	maxTitleLen := width - 5 - len(catDisplay)
	if maxTitleLen < 5 {
		maxTitleLen = 5
	}
	if utf8.RuneCountInString(title) > maxTitleLen {
		runes := []rune(title)
		title = string(runes[:maxTitleLen-2]) + ".."
	}

	cell := fmt.Sprintf("[%s] %s", priorityStyle.Render(priorityText), title)
	if catDisplay != "" {
		cell += " " + catDisplay
	}

	// Pad to width
	cellLen := 4 + len(title) + len(catDisplay)
	if catDisplay != "" {
		cellLen++ // space before category
	}
	if cellLen < width {
		cell += strings.Repeat(" ", width-cellLen)
	}

	if isSelected {
		cell = styles.Selected.Render(cell)
	}

	return cell
}

func (m *Model) viewHelp() string {
	var s string

	if m.previousMode == viewModeKanban {
		s = `┌─ Help - Kanban View ───────────────────┐
│                                        │
│ Navigation:                            │
│   h/←      : Move to left column       │
│   l/→      : Move to right column      │
│   j/↓      : Move down in column       │
│   k/↑      : Move up in column         │
│                                        │
│ Task Actions:                          │
│   Enter    : Advance to next status    │
│   e        : Edit task                 │
│   n        : Create new task           │
│   d        : Delete task               │
│                                        │
│ View:                                  │
│   v        : Switch to list view       │
│   f        : Filter settings           │
│   s        : Sort settings             │
│   ?/F1     : This help                 │
│   q        : Quit                      │
│                                        │
│         [Press Esc or ? to close]      │
└────────────────────────────────────────┘`
	} else {
		s = `┌─ Help - List View ─────────────────────┐
│                                        │
│ Navigation:                            │
│   j/↓      : Move down                 │
│   k/↑      : Move up                   │
│                                        │
│ Task Actions:                          │
│   Space    : Toggle status             │
│   e        : Edit task                 │
│   n        : Create new task           │
│   d        : Delete task               │
│                                        │
│ View:                                  │
│   v        : Switch to kanban view     │
│   f        : Filter settings           │
│   s        : Sort settings             │
│   ?/F1     : This help                 │
│   q        : Quit                      │
│                                        │
│         [Press Esc or ? to close]      │
└────────────────────────────────────────┘`
	}

	return s
}

func (m *Model) viewFilter() string {
	// Dynamic cursor positions
	categoryStartCursor := 11
	searchCursor := categoryStartCursor + len(m.categories)
	clearCursor := searchCursor + 1

	s := "┌─ Filter Settings ─────────────────────┐\n"
	s += "│                                        │\n"

	// Status checkboxes (cursor 0-2)
	s += "│ Status:                                │\n"
	statusLabels := []string{"New", "Working", "Completed"}
	statusValues := []domain.TaskStatus{domain.TaskStatusNew, domain.TaskStatusWorking, domain.TaskStatusCompleted}
	for i, label := range statusLabels {
		checked := m.hasFilterStatus(statusValues[i])
		checkbox := "[ ]"
		if checked {
			checkbox = "[x]"
		}
		cursor := "  "
		if m.filterCursor == i {
			cursor = "> "
		}
		line := fmt.Sprintf("%s%s %s", cursor, checkbox, label)
		padding := 38 - len(line)
		if padding < 0 {
			padding = 0
		}
		s += fmt.Sprintf("│ %s%s │\n", line, strings.Repeat(" ", padding))
	}

	s += "│                                        │\n"

	// Priority checkboxes (cursor 3-5)
	s += "│ Priority:                              │\n"
	priorityLabels := []string{"High", "Medium", "Low"}
	priorityValues := []domain.Priority{domain.PriorityHigh, domain.PriorityMedium, domain.PriorityLow}
	for i, label := range priorityLabels {
		checked := m.hasFilterPriority(priorityValues[i])
		checkbox := "[ ]"
		if checked {
			checkbox = "[x]"
		}
		cursor := "  "
		if m.filterCursor == 3+i {
			cursor = "> "
		}
		line := fmt.Sprintf("%s%s %s", cursor, checkbox, label)
		padding := 38 - len(line)
		if padding < 0 {
			padding = 0
		}
		s += fmt.Sprintf("│ %s%s │\n", line, strings.Repeat(" ", padding))
	}

	s += "│                                        │\n"

	// Date range radio buttons (cursor 6-10)
	s += "│ Due Date:                              │\n"
	dateLabels := []string{"All", "Today", "This Week", "Overdue", "No Due Date"}
	dateValues := []domain.DateRange{domain.DateRangeAll, domain.DateRangeToday, domain.DateRangeThisWeek, domain.DateRangeOverdue, domain.DateRangeNoDueDate}
	for i, label := range dateLabels {
		selected := m.filter.DateRange == dateValues[i]
		radio := "( )"
		if selected {
			radio = "(o)"
		}
		cursor := "  "
		if m.filterCursor == 6+i {
			cursor = "> "
		}
		line := fmt.Sprintf("%s%s %s", cursor, radio, label)
		// Account for multi-byte characters in padding calculation
		runeCount := len(cursor) + len(radio) + 1 + len([]rune(label))
		padding := 38 - runeCount
		if padding < 0 {
			padding = 0
		}
		s += fmt.Sprintf("│ %s%s │\n", line, strings.Repeat(" ", padding))
	}

	s += "│                                        │\n"

	// Category checkboxes (cursor 11 + i for each category)
	s += "│ Category:                              │\n"
	for i, cat := range m.categories {
		checked := m.hasFilterCategory(cat.ID)
		checkbox := "[ ]"
		if checked {
			checkbox = "[x]"
		}
		cursor := "  "
		if m.filterCursor == categoryStartCursor+i {
			cursor = "> "
		}
		line := fmt.Sprintf("%s%s %s", cursor, checkbox, cat.Name)
		// Use rune count for Unicode support
		runeCount := len(cursor) + len(checkbox) + 1 + utf8.RuneCountInString(cat.Name)
		padding := 38 - runeCount
		if padding < 0 {
			padding = 0
		}
		s += fmt.Sprintf("│ %s%s │\n", line, strings.Repeat(" ", padding))
	}

	s += "│                                        │\n"

	// Search text field (dynamic cursor position)
	searchCursorStr := "  "
	if m.filterCursor == searchCursor {
		searchCursorStr = "> "
	}
	searchLine := fmt.Sprintf("%sSearch: %s█", searchCursorStr, m.filter.SearchText)
	searchRuneCount := len(searchCursorStr) + len("Search: ") + len([]rune(m.filter.SearchText)) + 1
	searchPadding := 38 - searchRuneCount
	if searchPadding < 0 {
		searchPadding = 0
	}
	s += fmt.Sprintf("│ %s%s │\n", searchLine, strings.Repeat(" ", searchPadding))

	s += "│                                        │\n"

	// Clear button (dynamic cursor position)
	clearCursorStr := "  "
	if m.filterCursor == clearCursor {
		clearCursorStr = "> "
	}
	clearLine := fmt.Sprintf("%s[Clear]", clearCursorStr)
	clearRuneCount := len(clearCursorStr) + len("[Clear]")
	clearPadding := 38 - clearRuneCount
	if clearPadding < 0 {
		clearPadding = 0
	}
	s += fmt.Sprintf("│ %s%s │\n", clearLine, strings.Repeat(" ", clearPadding))

	s += "│                                        │\n"
	s += "│   [j/k]Move [Space]Select [Enter]Apply │\n"
	s += "│   [Esc]Cancel                          │\n"
	s += "└────────────────────────────────────────┘"

	return s
}

// updateFilterMode handles input in filter mode
func (m *Model) updateFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Dynamic cursor positions
	categoryStartCursor := 11
	searchCursor := categoryStartCursor + len(m.categories)
	clearCursor := searchCursor + 1
	maxCursor := clearCursor

	switch msg.String() {
	case "j", "down":
		if m.filterCursor < maxCursor {
			m.filterCursor++
		}

	case "k", "up":
		if m.filterCursor > 0 {
			m.filterCursor--
		}

	case " ":
		// Toggle selection based on cursor position
		switch {
		case m.filterCursor >= 0 && m.filterCursor <= 2:
			// Status toggle
			statusValues := []domain.TaskStatus{domain.TaskStatusNew, domain.TaskStatusWorking, domain.TaskStatusCompleted}
			m.toggleFilterStatus(statusValues[m.filterCursor])
		case m.filterCursor >= 3 && m.filterCursor <= 5:
			// Priority toggle
			priorityValues := []domain.Priority{domain.PriorityHigh, domain.PriorityMedium, domain.PriorityLow}
			m.toggleFilterPriority(priorityValues[m.filterCursor-3])
		case m.filterCursor >= 6 && m.filterCursor <= 10:
			// Date range selection (radio button)
			dateValues := []domain.DateRange{domain.DateRangeAll, domain.DateRangeToday, domain.DateRangeThisWeek, domain.DateRangeOverdue, domain.DateRangeNoDueDate}
			m.filter.DateRange = dateValues[m.filterCursor-6]
		case m.filterCursor >= categoryStartCursor && m.filterCursor < searchCursor:
			// Category toggle
			catIdx := m.filterCursor - categoryStartCursor
			if catIdx < len(m.categories) {
				m.toggleFilterCategory(m.categories[catIdx].ID)
			}
		case m.filterCursor == clearCursor:
			// Clear filter
			m.filter = domain.Filter{}
		}

	case "enter":
		// Apply filter and close
		m.mode = m.previousMode

	case "esc":
		// Cancel and close (keep current filter)
		m.mode = m.previousMode

	case "backspace":
		// Delete character from search text
		if m.filterCursor == searchCursor && len(m.filter.SearchText) > 0 {
			// Handle multi-byte characters properly
			runes := []rune(m.filter.SearchText)
			m.filter.SearchText = string(runes[:len(runes)-1])
		}

	default:
		// Character input for search text
		if m.filterCursor == searchCursor {
			if len(msg.String()) == 1 {
				m.filter.SearchText += msg.String()
			} else if msg.Type == tea.KeySpace {
				m.filter.SearchText += " "
			} else if msg.Type == tea.KeyRunes {
				m.filter.SearchText += string(msg.Runes)
			}
		}
	}

	return m, nil
}

// hasFilterStatus checks if a status is in the filter
func (m *Model) hasFilterStatus(status domain.TaskStatus) bool {
	for _, s := range m.filter.Statuses {
		if s == status {
			return true
		}
	}
	return false
}

// hasFilterPriority checks if a priority is in the filter
func (m *Model) hasFilterPriority(priority domain.Priority) bool {
	for _, p := range m.filter.Priorities {
		if p == priority {
			return true
		}
	}
	return false
}

// toggleFilterStatus toggles a status in the filter
func (m *Model) toggleFilterStatus(status domain.TaskStatus) {
	for i, s := range m.filter.Statuses {
		if s == status {
			// Remove status
			m.filter.Statuses = append(m.filter.Statuses[:i], m.filter.Statuses[i+1:]...)
			return
		}
	}
	// Add status
	m.filter.Statuses = append(m.filter.Statuses, status)
}

// toggleFilterPriority toggles a priority in the filter
func (m *Model) toggleFilterPriority(priority domain.Priority) {
	for i, p := range m.filter.Priorities {
		if p == priority {
			// Remove priority
			m.filter.Priorities = append(m.filter.Priorities[:i], m.filter.Priorities[i+1:]...)
			return
		}
	}
	// Add priority
	m.filter.Priorities = append(m.filter.Priorities, priority)
}

// hasFilterCategory checks if a category ID is in the filter
func (m *Model) hasFilterCategory(catID int64) bool {
	for _, id := range m.filter.Categories {
		if id == catID {
			return true
		}
	}
	return false
}

// toggleFilterCategory toggles a category in the filter
func (m *Model) toggleFilterCategory(catID int64) {
	for i, id := range m.filter.Categories {
		if id == catID {
			m.filter.Categories = append(m.filter.Categories[:i], m.filter.Categories[i+1:]...)
			return
		}
	}
	m.filter.Categories = append(m.filter.Categories, catID)
}

// viewSortMenu renders the sort menu overlay
func (m *Model) viewSortMenu() string {
	s := "┌─ Sort ──────────────┐\n"

	options := []struct {
		by    domain.SortBy
		label string
	}{
		{domain.SortByCreatedAt, "Created"},
		{domain.SortByDueDate, "Due Date"},
		{domain.SortByPriority, "Priority"},
		{domain.SortByStatus, "Status"},
		{domain.SortByTitle, "Title"},
	}

	for i, opt := range options {
		selected := " "
		order := ""
		if m.taskSort.By == opt.by {
			selected = "●"
			if m.taskSort.Ascending {
				order = " ↑"
			} else {
				order = " ↓"
			}
		}
		s += fmt.Sprintf("│ %d (%s) %-10s%s│\n", i+1, selected, opt.label, order)
	}

	s += "└─────────────────────┘\n"
	return s
}

func (m *Model) viewEdit() string {
	s := "┌─ Edit Task ───────────────────────────┐\n"
	s += "│                                        │\n"

	// Field labels with cursor indicator
	fields := []struct {
		label string
		value string
	}{
		{"Title", m.editTitle},
		{"Description", m.editDesc},
		{"Priority", ""},    // Rendered specially
		{"Category", ""},    // Rendered specially
		{"Due Date", m.editDueDate},
	}

	for i, field := range fields {
		cursor := "  "
		if m.editCursor == i {
			cursor = "> "
		}

		// Show value with cursor if editing this field
		value := field.value
		if i == 2 {
			// Priority - show as selector
			value = m.renderPrioritySelector()
		} else if i == 3 {
			// Category - show as selector
			value = m.renderCategorySelector()
		}
		if m.editCursor == i && m.editingField && i != 2 && i != 3 {
			value += "█"
		}
		if value == "" && i != 2 && i != 3 {
			value = "(empty)"
		}

		// Truncate long values for display
		maxValueLen := 24
		displayValue := value
		if len(displayValue) > maxValueLen {
			displayValue = displayValue[:maxValueLen-2] + ".."
		}

		line := fmt.Sprintf("%s%-12s %s", cursor, field.label+":", displayValue)
		padding := 38 - len(line)
		if padding < 0 {
			padding = 0
		}
		s += fmt.Sprintf("│ %s%s │\n", line, strings.Repeat(" ", padding))
	}

	s += "│                                        │\n"

	// Save and Cancel buttons
	saveCursor := "  "
	cancelCursor := "  "
	if m.editCursor == 5 {
		saveCursor = "> "
	}
	if m.editCursor == 6 {
		cancelCursor = "> "
	}
	s += fmt.Sprintf("│   %s[Save]  %s[Cancel]                 │\n", saveCursor, cancelCursor)

	// Error message if any
	if m.editError != "" {
		s += "│                                        │\n"
		errorLine := fmt.Sprintf("│ Error: %-30s │\n", m.editError)
		s += errorLine
	}

	s += "│                                        │\n"
	s += "│ [j/k]Move [Enter]Edit [Tab]Cycle [Esc] │\n"
	s += "└────────────────────────────────────────┘"

	return s
}

func (m *Model) renderPrioritySelector() string {
	switch m.editPriority {
	case domain.PriorityHigh:
		return styles.PriorityHigh.Render("[H]") + " M L"
	case domain.PriorityMedium:
		return "H " + styles.PriorityMedium.Render("[M]") + " L"
	case domain.PriorityLow:
		return "H M " + styles.PriorityLow.Render("[L]")
	default:
		return "H M L"
	}
}

func (m *Model) renderCategorySelector() string {
	if m.editCategoryIdx < 0 {
		return "[None]"
	}
	if m.editCategoryIdx < len(m.categories) {
		return "[" + m.categories[m.editCategoryIdx].Name + "]"
	}
	return "[None]"
}

// getCategoryName returns the category name for a task, or empty string if no category
func (m *Model) getCategoryName(task *domain.Task) string {
	if task.CategoryID == nil {
		return ""
	}
	for _, cat := range m.categories {
		if cat.ID == *task.CategoryID {
			return cat.Name
		}
	}
	return ""
}
