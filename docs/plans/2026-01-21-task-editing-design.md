# Task Editing Feature Design

**Created:** 2026-01-21
**Status:** Ready for implementation

## Overview

Add ability to edit existing tasks with all fields: title, description, priority, and due date. Uses a modal with field focus pattern.

## User Interaction

### Entry Point
- Press `e` in list view or kanban view on selected task
- Opens edit modal with current task data pre-filled

### Navigation
- `j/k` or `↓/↑`: Move between fields
- `Enter`: Start editing focused field (or activate Save/Cancel button)
- `Esc`: Cancel field edit, or exit modal if not editing a field

### Field Editing
- **Title**: Text input, Enter to confirm
- **Description**: Text input (single line for now), Enter to confirm
- **Priority**: Tab or arrow keys to cycle Low/Medium/High, Enter to confirm
- **Due Date**: Text input as YYYY-MM-DD, Enter to confirm, validated on save

## UI Layout

```
┌─ Edit Task ───────────────────────────┐
│                                        │
│ > Title:       Buy groceries           │
│   Description: Milk, eggs, bread       │
│   Priority:    [M] Medium              │
│   Due Date:    2026-01-25              │
│                                        │
│   [Save]  [Cancel]                     │
│                                        │
│ [j/k]Move [Enter]Edit [Esc]Cancel      │
└────────────────────────────────────────┘
```

When editing a field, show cursor:
```
│ > Title:       Buy groceries█          │
```

## Implementation Plan

### Task 1: Add Edit State to Model

**File:** `internal/app/app.go`

Add to Model struct:
```go
// Edit state
editTask      *domain.Task  // Copy of task being edited
editCursor    int           // 0=title, 1=desc, 2=priority, 3=date, 4=save, 5=cancel
editingField  bool          // Currently typing in a field
editTitle     string        // Edited title value
editDesc      string        // Edited description value
editPriority  domain.Priority
editDueDate   string        // String for input, parsed on save
```

Add view mode:
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

### Task 2: Add Edit Key Handler

**File:** `internal/app/app.go`

Add `e` key handler to list mode and kanban mode:
```go
case "e":
    // Get selected task
    if len(filteredTasks) > 0 && m.cursor < len(filteredTasks) {
        task := filteredTasks[m.cursor]
        m.startEditMode(task)
    }
```

Add helper method:
```go
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
    m.mode = viewModeEdit
}
```

### Task 3: Implement viewEdit Method

**File:** `internal/app/app.go`

Render the edit modal showing all fields with cursor indicator.

### Task 4: Implement updateEditMode Method

**File:** `internal/app/app.go`

Handle:
- `j/k`: Navigate between fields (when not editing)
- `Enter`: Start editing field or activate button
- `Esc`: Cancel edit or exit modal
- Character input when editing text fields
- `Tab` when editing priority to cycle options
- Validation for due date format

### Task 5: Implement Save Logic

**File:** `internal/app/app.go`

Add saveEditedTask method:
```go
func (m *Model) saveEditedTask() tea.Cmd {
    return func() tea.Msg {
        // Validate
        m.editTask.Title = m.editTitle
        m.editTask.Description = m.editDesc
        m.editTask.Priority = m.editPriority

        if m.editDueDate != "" {
            date, err := time.Parse("2006-01-02", m.editDueDate)
            if err != nil {
                return errMsg{err: errors.New("invalid date format")}
            }
            m.editTask.DueDate = &date
        } else {
            m.editTask.DueDate = nil
        }

        if err := m.editTask.Validate(); err != nil {
            return errMsg{err: err}
        }

        err := m.repo.Update(context.Background(), m.editTask)
        if err != nil {
            return errMsg{err: err}
        }

        return taskUpdatedMsg{task: m.editTask}
    }
}
```

### Task 6: Update Help Modal

**File:** `internal/app/app.go`

Add `e` key to help text for both list and kanban views.

### Task 7: Testing

- Manual test: edit each field type
- Verify validation errors display
- Verify cancel doesn't save changes
- Verify changes persist after save

## Acceptance Criteria

- [ ] Press `e` opens edit modal with task data
- [ ] Can navigate between fields with j/k
- [ ] Can edit title (text input)
- [ ] Can edit description (text input)
- [ ] Can edit priority (cycle with Tab)
- [ ] Can edit due date (YYYY-MM-DD format)
- [ ] Save updates task in database
- [ ] Cancel exits without saving
- [ ] Invalid date shows error
- [ ] Empty title shows error
- [ ] Help modal updated with `e` key

## Future Enhancements (Out of Scope)

- Multi-line description editing
- Date picker widget
- Category assignment (separate feature)
- Undo/redo
