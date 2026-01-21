# Task Editing Feature Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add ability to edit existing tasks with title, description, priority, and due date fields via a modal interface

**Architecture:** Add new `viewModeEdit` view mode with field-based navigation. Edit state stored in Model struct with temporary values that get committed on save. TDD approach for validation logic.

**Tech Stack:** Go 1.24+, Bubble Tea, Lipgloss, modernc.org/sqlite

---

## Task 1: Add Edit State to Model

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add viewModeEdit constant**

File: `internal/app/app.go` (modify viewMode constants)

```go
const (
	viewModeList viewMode = iota
	viewModeCreate
	viewModeKanban
	viewModeFilter
	viewModeHelp
	viewModeEdit  // New
)
```

**Step 2: Add edit state fields to Model struct**

File: `internal/app/app.go` (add after sort state in Model struct)

```go
	// Edit state
	editTask      *domain.Task  // Reference to task being edited
	editCursor    int           // 0=title, 1=desc, 2=priority, 3=date, 4=save, 5=cancel
	editingField  bool          // Currently typing in a field
	editTitle     string        // Edited title value
	editDesc      string        // Edited description value
	editPriority  domain.Priority
	editDueDate   string        // String for input, parsed on save
	editError     string        // Validation error message
```

**Step 3: Build to verify compilation**

Run: `go build ./cmd/task`

Expected: Build succeeds

**Step 4: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: add edit state to model

Add viewModeEdit constant and edit state fields:
- editTask: reference to task being edited
- editCursor: field navigation (0-5)
- editingField: whether currently typing
- edit field values for title, desc, priority, date
- editError: validation error display

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Add startEditMode Helper

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add startEditMode method**

File: `internal/app/app.go` (add method)

```go
// startEditMode initializes edit mode with the given task
func (m *Model) startEditMode(task *domain.Task) {
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
	m.editError = ""
	m.mode = viewModeEdit
}
```

**Step 2: Build to verify compilation**

Run: `go build ./cmd/task`

Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: add startEditMode helper

Initialize edit mode with task data:
- Copy current values to edit fields
- Format due date as YYYY-MM-DD string
- Reset cursor and error state

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Add Edit Key Handler to List Mode

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add 'e' key handler to list mode**

File: `internal/app/app.go` (add to list mode key handlers in Update method, after the 'd' case)

```go
		case "e":
			// Edit selected task
			filteredTasks := m.filter.Apply(m.tasks)
			sortedTasks := m.taskSort.Apply(filteredTasks)
			if len(sortedTasks) > 0 && m.cursor < len(sortedTasks) {
				task := sortedTasks[m.cursor]
				m.startEditMode(task)
			}
```

**Step 2: Build and test manually**

Run: `go build ./cmd/task && ./bin/task`

Test: Create a task, select it, press 'e' - should enter edit mode (no view yet, but no crash)

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: add edit key handler to list mode

Press 'e' on selected task to enter edit mode.
Uses filtered and sorted task list for correct selection.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Add Edit Key Handler to Kanban Mode

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add 'e' key handler to kanban mode**

File: `internal/app/app.go` (add to updateKanbanMode method, after the 'd' case)

```go
	case "e":
		// Edit selected task
		col := m.kanbanColumn
		if len(columns[col]) > 0 && m.kanbanCursors[col] < len(columns[col]) {
			task := columns[col][m.kanbanCursors[col]]
			m.startEditMode(task)
		}
```

**Step 2: Build and test manually**

Run: `go build ./cmd/task && ./bin/task`

Test: Press 'v' for kanban, select a task, press 'e' - should enter edit mode

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: add edit key handler to kanban mode

Press 'e' on selected task in kanban view to enter edit mode.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Implement Edit View Rendering

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add viewEdit method**

File: `internal/app/app.go` (add method)

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
		{"Priority", m.editPriority.String()},
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
		}
		if m.editCursor == i && m.editingField && i != 2 {
			value += "█"
		}
		if value == "" && i != 2 {
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
	if m.editCursor == 4 {
		saveCursor = "> "
	}
	if m.editCursor == 5 {
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
	s += "│ [j/k]Move [Enter]Edit [Esc]Cancel      │\n"
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
```

**Step 2: Add viewEdit to View method**

File: `internal/app/app.go` (add to View method, after filter mode check)

```go
	// Edit mode view
	if m.mode == viewModeEdit {
		return m.viewEdit()
	}
```

**Step 3: Build and test manually**

Run: `go build ./cmd/task && ./bin/task`

Test: Create a task, press 'e' - should see edit modal with task data

**Step 4: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: implement edit view rendering

Add edit modal display:
- Shows title, description, priority, due date fields
- Cursor indicator for selected field
- Priority displayed as selector (H M L)
- Save and Cancel buttons
- Error message area
- Help text at bottom

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: Implement Edit Mode Navigation

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add updateEditMode method**

File: `internal/app/app.go` (add method)

```go
// updateEditMode handles input in edit mode
func (m *Model) updateEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If editing a text field, handle text input
	if m.editingField {
		return m.updateEditFieldInput(msg)
	}

	// Navigation mode
	switch msg.String() {
	case "j", "down":
		if m.editCursor < 5 {
			m.editCursor++
		}

	case "k", "up":
		if m.editCursor > 0 {
			m.editCursor--
		}

	case "enter":
		switch m.editCursor {
		case 0, 1, 3:
			// Start editing text field (title, description, due date)
			m.editingField = true
		case 2:
			// Priority - cycle on Enter
			m.cyclePriority()
		case 4:
			// Save button
			return m.saveEditedTask()
		case 5:
			// Cancel button
			m.mode = viewModeList
			m.editError = ""
		}

	case "tab":
		// Cycle priority when on priority field
		if m.editCursor == 2 {
			m.cyclePriority()
		}

	case "esc":
		// Cancel edit
		m.mode = viewModeList
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
```

**Step 2: Add edit mode handling to Update method**

File: `internal/app/app.go` (add to Update method, after create mode check)

```go
		// Handle edit mode
		if m.mode == viewModeEdit {
			return m.updateEditMode(msg)
		}
```

**Step 3: Build and test manually**

Run: `go build ./cmd/task && ./bin/task`

Test:
- Enter edit mode, use j/k to navigate between fields
- Press Tab on priority to cycle
- Press Esc to cancel

**Step 4: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: implement edit mode navigation

Add field navigation and actions:
- j/k for moving between fields
- Enter to start editing or activate buttons
- Tab to cycle priority
- Esc to cancel and return to previous view

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 7: Implement Edit Field Text Input

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add updateEditFieldInput method**

File: `internal/app/app.go` (add method)

```go
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
		case 3: // Due Date
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
			case 3: // Due Date
				// Only allow date-like characters
				if len(m.editDueDate) < 10 {
					m.editDueDate += char
				}
			}
		}
	}

	return m, nil
}
```

**Step 2: Build and test manually**

Run: `go build ./cmd/task && ./bin/task`

Test:
- Enter edit mode, press Enter on Title field
- Type some text, see cursor
- Press Backspace to delete
- Press Enter or Esc to confirm

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: implement edit field text input

Add text editing for title, description, and due date:
- Backspace to delete (Unicode-safe)
- Enter/Esc to confirm and exit editing
- Character input with proper handling
- Due date limited to 10 characters (YYYY-MM-DD)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 8: Implement Save Logic with Validation

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add saveEditedTask method**

File: `internal/app/app.go` (add method)

```go
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

	// Validate using domain validation
	if err := m.editTask.Validate(); err != nil {
		m.editError = err.Error()
		return m, nil
	}

	// Save to repository
	m.mode = viewModeList
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
```

**Step 2: Build and test manually**

Run: `go build ./cmd/task && ./bin/task`

Test:
- Edit a task, change title, save - should update
- Clear title, try to save - should show error
- Enter invalid date format, try to save - should show error

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: implement save logic with validation

Add task save functionality:
- Validate title is not empty
- Validate due date format (YYYY-MM-DD)
- Run domain validation
- Display error messages inline
- Save to repository on success

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 9: Update Help Modal with Edit Key

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Update viewHelp method**

File: `internal/app/app.go` (modify viewHelp method)

Update the list view help text to include the 'e' key:

```go
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
```

Also update kanban view help text:

```go
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
```

**Step 2: Build and verify**

Run: `go build ./cmd/task && ./bin/task`

Test: Press '?' in list and kanban views, verify 'e' key is shown

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: update help modal with edit key

Add 'e' key documentation to both list and kanban view help:
- Listed under Task Actions section
- Consistent with other key bindings

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 10: Update Status Bars

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Update list view status bar**

File: `internal/app/app.go` (modify viewList status bar)

```go
	// Status bar
	helpText := "[n]New [e]Edit [d]Delete [Space]Status [f]Filter [s]Sort [v]Kanban [?]Help [q]Quit"
	s += styles.StatusBar.Render(helpText) + "\n"
```

**Step 2: Update kanban view status bar**

File: `internal/app/app.go` (modify viewKanban status bar)

```go
	// Status bar
	helpText := "[h/l]Column [j/k]Up/Down [Enter]Advance [e]Edit [f]Filter [s]Sort [v]List [?]Help [q]Quit"
	s += "\n" + styles.StatusBar.Render(helpText) + "\n"
```

**Step 3: Build and verify**

Run: `go build ./cmd/task && ./bin/task`

Test: Verify status bars show [e]Edit in both views

**Step 4: Commit**

```bash
git add internal/app/app.go
git commit -m "$(cat <<'EOF'
feat: update status bars with edit shortcut

Add [e]Edit to status bar help text in both list and kanban views.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 11: Run All Tests

**Step 1: Run test suite**

Run: `go test ./... -v`

Expected: All tests pass

**Step 2: Fix any test failures**

If tests fail, fix issues and re-run.

**Step 3: Commit any fixes**

```bash
git add -A
git commit -m "$(cat <<'EOF'
fix: address test failures

[Describe any fixes made]

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 12: Manual Integration Test

**Step 1: Build application**

Run: `make build`

Expected: Binary created at `bin/task`

**Step 2: Run manual test checklist**

Run: `./bin/task`

Test each acceptance criterion:

- [ ] Press `e` opens edit modal with task data
- [ ] Can navigate between fields with j/k
- [ ] Can edit title (text input)
- [ ] Can edit description (text input)
- [ ] Can edit priority (cycle with Tab/Enter)
- [ ] Can edit due date (YYYY-MM-DD format)
- [ ] Save updates task in database
- [ ] Cancel exits without saving
- [ ] Invalid date shows error
- [ ] Empty title shows error
- [ ] Help modal updated with `e` key

**Step 3: Final commit**

```bash
git add -A
git commit -m "$(cat <<'EOF'
test: verify task editing feature complete

Task editing feature implementation complete:
- Press 'e' to edit selected task in list/kanban view
- Edit modal with field navigation (j/k)
- Text editing for title, description, due date
- Priority cycling with Tab/Enter
- Save with validation (title required, date format)
- Error messages displayed inline
- Help modal and status bars updated

All acceptance criteria verified:
- Edit modal opens with task data
- Field navigation works
- All fields editable
- Save persists changes
- Cancel discards changes
- Validation errors display correctly

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Summary

Task Editing feature is complete when:

- All unit tests pass (`go test ./... -v`)
- Application builds successfully (`make build`)
- Manual testing confirms all acceptance criteria:
  - Press 'e' opens edit modal
  - Can navigate and edit all fields
  - Save updates database
  - Cancel exits without saving
  - Validation errors display correctly
  - Help modal and status bars updated
- All changes committed to git

**Dependencies:** This plan assumes Phase 1 and Phase 2 are already implemented.

**Estimated tasks:** 12 tasks, each with 2-5 minute steps.
