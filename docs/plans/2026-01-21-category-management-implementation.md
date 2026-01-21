# Category Management Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enable users to assign categories to tasks with visual display and filtering support

**Architecture:** Extend the existing Model to load categories on init, add category selection to create/edit forms, display category in task list/kanban views, and integrate with existing filter system. Repository layer already has CreateCategory and GetCategories methods.

**Tech Stack:** Go 1.24+, Bubble Tea, Lipgloss, modernc.org/sqlite

---

## Task 1: Add Category State to Model

**Files:**
- Modify: `internal/app/app.go`
- Modify: `internal/app/messages.go`

**Step 1: Add category state fields to Model struct**

File: `internal/app/app.go` (add after edit state in Model struct)

```go
	// Category state
	categories []*domain.Category // All available categories
```

**Step 2: Add categoriesLoadedMsg to messages.go**

File: `internal/app/messages.go` (add message type)

```go
// categoriesLoadedMsg is sent when categories are loaded
type categoriesLoadedMsg struct {
	categories []*domain.Category
}
```

**Step 3: Add loadCategories command**

File: `internal/app/app.go` (add method after loadTasks)

```go
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
```

**Step 4: Update Init to load categories**

File: `internal/app/app.go` (modify Init method)

```go
// Init initializes the application
func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.loadTasks(), m.loadCategories())
}
```

**Step 5: Handle categoriesLoadedMsg in Update**

File: `internal/app/app.go` (add case in Update method, after taskListLoadedMsg)

```go
	case categoriesLoadedMsg:
		m.categories = msg.categories
```

**Step 6: Build to verify**

Run: `go build ./cmd/task`

**Step 7: Commit**

```bash
git add internal/app/
git commit -m "$(cat <<'EOF'
feat: add category state to model

- Add categories slice to Model
- Add loadCategories command
- Load categories on Init alongside tasks
- Handle categoriesLoadedMsg in Update

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Add Category to Create Form

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add category input field to Model**

File: `internal/app/app.go` (add to Model struct, after inputPriority)

```go
	inputCategoryIdx int // Index into categories slice, -1 for no category
```

**Step 2: Initialize category in 'n' key handler**

File: `internal/app/app.go` (modify the 'n' case in list mode)

```go
		case "n":
			// Enter create mode
			m.mode = viewModeCreate
			m.inputTitle = ""
			m.inputPriority = domain.PriorityMedium
			m.inputCategoryIdx = -1 // No category selected by default
```

Also update the same in kanban mode's 'n' handler.

**Step 3: Update viewCreate to show category selector**

File: `internal/app/app.go` (modify viewCreate method)

```go
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
```

**Step 4: Add Shift+Tab handler for category cycling**

File: `internal/app/app.go` (modify updateCreateMode)

```go
	case "shift+tab":
		// Cycle category
		if len(m.categories) > 0 {
			m.inputCategoryIdx++
			if m.inputCategoryIdx >= len(m.categories) {
				m.inputCategoryIdx = -1 // Back to "None"
			}
		}
```

**Step 5: Update createTask to include category**

File: `internal/app/app.go` (modify createTask method)

```go
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
```

**Step 6: Update the call to createTask**

File: `internal/app/app.go` (modify the Enter handler in updateCreateMode)

```go
	case "enter":
		// Create task
		if m.inputTitle != "" {
			m.mode = viewModeList
			return m, m.createTask(m.inputTitle, m.inputPriority, m.inputCategoryIdx)
		}
```

**Step 7: Build and test**

Run: `go build ./cmd/task && ./bin/task`

Test: Press 'n', use Shift+Tab to cycle categories, create task

**Step 8: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: add category selection to create form

- Add inputCategoryIdx to Model
- Show category selector in create view
- Shift+Tab cycles through categories
- Create task with selected category

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Add Category to Edit Form

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add editCategoryIdx to Model**

File: `internal/app/app.go` (add to edit state in Model struct)

```go
	editCategoryIdx int // Index into categories slice, -1 for no category
```

**Step 2: Update startEditMode to initialize category**

File: `internal/app/app.go` (modify startEditMode)

```go
// startEditMode initializes edit mode with the given task
func (m *Model) startEditMode(task *domain.Task) {
	m.previousMode = m.mode
	m.editTask = task
	m.editCursor = 0
	m.editingField = false
	m.editTitle = task.Title
	m.editDesc = task.Description
	m.editPriority = task.Priority
	if task.DueDate != nil {
		m.editDueDate = task.DueDate.Format("2006-01-02")
	} else {
		m.editDueDate = ""
	}
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
	m.editError = ""
	m.mode = viewModeEdit
}
```

**Step 3: Update viewEdit to show category field**

File: `internal/app/app.go` (modify viewEdit - add category as 5th field, shift save/cancel to 5/6)

Update the fields slice and adjust cursor positions:
- 0: Title
- 1: Description
- 2: Priority
- 3: Category (NEW)
- 4: Due Date
- 5: Save button
- 6: Cancel button

```go
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
		{"Priority", ""}, // Rendered specially
		{"Category", ""}, // Rendered specially
		{"Due Date", m.editDueDate},
	}

	for i, field := range fields {
		cursor := "  "
		if m.editCursor == i {
			cursor = "> "
		}

		value := field.value
		if i == 2 {
			// Priority - show as selector
			value = m.renderPrioritySelector()
		} else if i == 3 {
			// Category - show as selector
			value = m.renderCategorySelector()
		} else {
			if m.editCursor == i && m.editingField {
				value += "█"
			}
			if value == "" {
				value = "(empty)"
			}
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

	// Save and Cancel buttons (now at positions 5 and 6)
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

func (m *Model) renderCategorySelector() string {
	if m.editCategoryIdx < 0 {
		return "[None]"
	}
	if m.editCategoryIdx < len(m.categories) {
		return "[" + m.categories[m.editCategoryIdx].Name + "]"
	}
	return "[None]"
}
```

**Step 4: Update updateEditMode cursor bounds and handlers**

File: `internal/app/app.go` (modify updateEditMode)

```go
// updateEditMode handles input in edit mode
func (m *Model) updateEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editingField {
		return m.updateEditFieldInput(msg)
	}

	switch msg.String() {
	case "j", "down":
		if m.editCursor < 6 { // Now max is 6 (cancel button)
			m.editCursor++
		}

	case "k", "up":
		if m.editCursor > 0 {
			m.editCursor--
		}

	case "enter":
		switch m.editCursor {
		case 0, 1, 4: // Title, Description, Due Date (note: due date moved to 4)
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
		m.mode = m.previousMode
		m.editError = ""
	}

	return m, nil
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
```

**Step 5: Update saveEditedTask to save category**

File: `internal/app/app.go` (modify saveEditedTask)

Add before the domain validation:

```go
	// Update category
	if m.editCategoryIdx >= 0 && m.editCategoryIdx < len(m.categories) {
		catID := m.categories[m.editCategoryIdx].ID
		m.editTask.CategoryID = &catID
	} else {
		m.editTask.CategoryID = nil
	}
```

**Step 6: Build and test**

Run: `go build ./cmd/task && ./bin/task`

Test: Edit a task, change category, save, verify category is saved

**Step 7: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: add category field to edit form

- Add editCategoryIdx to Model
- Show category selector in edit view
- Tab/Enter cycles category when on category field
- Save task with updated category
- Adjust edit cursor positions for new field

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Display Category in List View

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add helper to get category name by ID**

File: `internal/app/app.go` (add method)

```go
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
```

**Step 2: Update viewList to show category**

File: `internal/app/app.go` (modify viewList, in the task rendering loop)

Update the line building to include category:

```go
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
```

**Step 3: Build and test**

Run: `go build ./cmd/task && ./bin/task`

Test: Create task with category, verify category shows as @CategoryName in list

**Step 4: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: display category in list view

Show category as @CategoryName after task title in list view.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Display Category in Kanban View

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Update renderKanbanCell to show category**

File: `internal/app/app.go` (modify renderKanbanCell)

```go
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
		// Truncate category name if needed
		if len(catName) > 8 {
			catName = catName[:7] + "."
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
```

**Step 2: Build and test**

Run: `go build ./cmd/task && ./bin/task`

Test: View kanban, verify category shows on tasks

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: display category in kanban view

Show category as @CategoryName in kanban cells.
Truncate long category names to fit in cell.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: Integrate Category with Filter

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Update viewFilter to show category filter section**

The domain.Filter already supports Categories []int64. We need to add UI for it.

File: `internal/app/app.go` (modify viewFilter to add category section)

Add after the date range section, before the search text:

```go
	s += "│                                        │\n"

	// Category checkboxes (cursor 11 + i for each category)
	s += "│ Category:                              │\n"
	categoryStartCursor := 11
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
		padding := 38 - len(line)
		if padding < 0 {
			padding = 0
		}
		s += fmt.Sprintf("│ %s%s │\n", line, strings.Repeat(" ", padding))
	}
```

**Step 2: Add hasFilterCategory helper**

File: `internal/app/app.go` (add method)

```go
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
```

**Step 3: Update filter cursor bounds and handlers**

The cursor positions need to be adjusted:
- 0-2: Status (3 items)
- 3-5: Priority (3 items)
- 6-10: Date range (5 items)
- 11 to 11+len(categories)-1: Categories
- Next: Search text
- Next: Clear button

File: `internal/app/app.go` (update updateFilterMode)

Update the maxCursor calculation and add category toggle handling:

```go
func (m *Model) updateFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

	// ... rest of the handlers stay the same but update cursor references
```

**Step 4: Update viewFilter search and clear cursor positions**

Update the search text and clear button rendering to use dynamic cursor positions.

**Step 5: Build and test**

Run: `go build ./cmd/task && ./bin/task`

Test: Open filter, toggle category filters, apply, verify filtering works

**Step 6: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: add category filter support

- Add category checkboxes to filter view
- Toggle categories with Space key
- Filter tasks by selected categories

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 7: Update Help Modal

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Update create mode help in status bar**

Already done in Task 2 (Shift+Tab for category)

**Step 2: Update edit mode help**

File: `internal/app/app.go` (already updated in Task 3)

**Step 3: Build and verify**

Run: `go build ./cmd/task && ./bin/task`

**Step 4: Commit (if any changes)**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
docs: update help text for category features

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 8: Run All Tests

**Step 1: Run test suite**

Run: `go test ./... -v`

Expected: All tests pass

**Step 2: Fix any test failures**

If tests fail, fix and re-run.

**Step 3: Commit any fixes**

```bash
git add -A
git commit -m "$(cat <<'EOF'
fix: address test failures

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 9: Manual Integration Test

**Step 1: Build application**

Run: `make build`

**Step 2: Run manual test checklist**

Run: `./bin/task`

Test:
- [ ] Create task with category (Shift+Tab to select)
- [ ] Edit task, change category
- [ ] List view shows @CategoryName
- [ ] Kanban view shows @CategoryName
- [ ] Filter by category works
- [ ] Clear filter removes category filter
- [ ] Tasks without category show no @

**Step 3: Final commit**

```bash
git add -A
git commit -m "$(cat <<'EOF'
test: verify category management feature complete

Category management implementation complete:
- Categories loaded on app init
- Create form: Shift+Tab to select category
- Edit form: category field with Tab/Enter cycling
- List view: shows @CategoryName
- Kanban view: shows @CategoryName
- Filter: category checkboxes

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Summary

Category Management feature is complete when:

- All unit tests pass (`go test ./... -v`)
- Application builds successfully (`make build`)
- Manual testing confirms:
  - Can assign category when creating task
  - Can change category when editing task
  - Category displays in list view as @Name
  - Category displays in kanban view as @Name
  - Can filter tasks by category
- All changes committed to git

**Dependencies:** This plan assumes Phase 1, Phase 2, and Task Editing are already implemented.

**Estimated tasks:** 9 tasks with bite-sized steps.
