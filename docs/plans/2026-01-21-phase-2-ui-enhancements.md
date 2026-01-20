# Phase 2: UI Enhancements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extend the task management TUI with Kanban view, filtering, sorting, and help modal

**Architecture:** Add new view modes to the existing Bubble Tea application. Filter/sort logic lives in domain layer. Each view is a separate model that can be swapped via view mode switching. Modal views (filter, help) overlay the current view.

**Tech Stack:** Go 1.24+, Bubble Tea, Lipgloss, modernc.org/sqlite

---

## Task 1: View Mode Infrastructure

**Files:**
- Modify: `internal/app/app.go`
- Modify: `internal/app/messages.go`

**Step 1: Add view mode constants**

File: `internal/app/app.go` (modify viewMode constants)

```go
type viewMode int

const (
	viewModeList viewMode = iota
	viewModeCreate
	viewModeKanban
	viewModeFilter
	viewModeHelp
)
```

**Step 2: Add view toggle key handler**

File: `internal/app/app.go` (add to list mode key handlers)

```go
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
```

**Step 3: Add previousMode field to Model**

File: `internal/app/app.go` (modify Model struct)

```go
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
```

**Step 4: Build to verify compilation**

Run: `go build ./cmd/task`

Expected: Build succeeds

**Step 5: Commit**

```bash
git add internal/app/
git commit -m "feat: add view mode infrastructure

Add view mode constants for kanban, filter, and help views.
Add previousMode field for modal overlay support.
Prepare key handlers for view switching.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 2: Help Modal

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add help view rendering**

File: `internal/app/app.go` (add method)

```go
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
```

**Step 2: Add help mode key handler**

File: `internal/app/app.go` (add to Update method after create mode check)

```go
		// Handle help mode
		if m.mode == viewModeHelp {
			switch msg.String() {
			case "esc", "?", "f1":
				m.mode = m.previousMode
			}
			return m, nil
		}
```

**Step 3: Update View method to render help**

File: `internal/app/app.go` (modify View method)

```go
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
```

**Step 4: Add placeholder viewKanban method**

File: `internal/app/app.go` (add method)

```go
func (m *Model) viewKanban() string {
	return "Kanban view (coming soon)\n\nPress 'v' to switch to list view, '?' for help, 'q' to quit.\n"
}
```

**Step 5: Build and test**

Run: `go build ./cmd/task && ./bin/task`

Test: Press '?' to see help modal, press 'Esc' to close

**Step 6: Commit**

```bash
git add internal/app/app.go
git commit -m "feat: add help modal

Implement context-aware help modal:
- Different help text for list vs kanban view
- Toggle with ? or F1 key
- Close with Esc or ? key
- Modal overlay pattern with previousMode tracking

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 3: Kanban View Model

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add kanban cursor fields to Model**

File: `internal/app/app.go` (modify Model struct)

```go
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
	// Kanban view state
	kanbanColumn  int // 0=New, 1=Working, 2=Completed
	kanbanCursors [3]int // Cursor position within each column
}
```

**Step 2: Add helper to get tasks by status**

File: `internal/app/app.go` (add method)

```go
func (m *Model) tasksByStatus(status domain.TaskStatus) []*domain.Task {
	var result []*domain.Task
	for _, task := range m.tasks {
		if task.Status == status {
			result = append(result, task)
		}
	}
	return result
}
```

**Step 3: Implement viewKanban method**

File: `internal/app/app.go` (replace placeholder)

```go
func (m *Model) viewKanban() string {
	newTasks := m.tasksByStatus(domain.TaskStatusNew)
	workingTasks := m.tasksByStatus(domain.TaskStatusWorking)
	completedTasks := m.tasksByStatus(domain.TaskStatusCompleted)

	// Column width
	colWidth := 25

	// Header
	s := fmt.Sprintf("┌─ New (%d) %s┬─ Working (%d) %s┬─ Completed (%d) %s┐\n",
		len(newTasks), strings.Repeat("─", colWidth-10-len(fmt.Sprintf("%d", len(newTasks)))),
		len(workingTasks), strings.Repeat("─", colWidth-13-len(fmt.Sprintf("%d", len(workingTasks)))),
		len(completedTasks), strings.Repeat("─", colWidth-15-len(fmt.Sprintf("%d", len(completedTasks)))),
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
	helpText := "[h/l]列移動 [j/k]上下 [Enter]次へ [v]リスト [?]ヘルプ [q]終了"
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
		priorityText = "高"
	case domain.PriorityMedium:
		priorityStyle = styles.PriorityMedium
		priorityText = "中"
	case domain.PriorityLow:
		priorityStyle = styles.PriorityLow
		priorityText = "低"
	}

	// Truncate title if needed
	title := task.Title
	maxTitleLen := width - 6 // "[優] " + padding
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen-2] + ".."
	}

	cell := fmt.Sprintf("[%s] %s", priorityStyle.Render(priorityText), title)

	// Pad to width
	cellLen := len(fmt.Sprintf("[%s] %s", priorityText, title))
	if cellLen < width {
		cell += strings.Repeat(" ", width-cellLen)
	}

	if isSelected {
		cell = styles.Selected.Render(cell)
	}

	return cell
}
```

**Step 4: Add strings import**

File: `internal/app/app.go` (add to imports)

```go
import (
	"context"
	"fmt"
	"strings"
	"time"
	// ... rest of imports
)
```

**Step 5: Build to verify compilation**

Run: `go build ./cmd/task`

Expected: Build succeeds

**Step 6: Commit**

```bash
git add internal/app/app.go
git commit -m "feat: add kanban view rendering

Implement kanban board display:
- Three columns for New, Working, Completed
- Task count in column headers
- Priority indicators and truncated titles
- Selected task highlighting
- Responsive column layout

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 4: Kanban Navigation

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add kanban key handlers**

File: `internal/app/app.go` (add to Update method after help mode check, before list mode handlers)

```go
		// Handle kanban mode
		if m.mode == viewModeKanban {
			return m.updateKanbanMode(msg)
		}
```

**Step 2: Implement updateKanbanMode method**

File: `internal/app/app.go` (add method)

```go
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

	case "d":
		col := m.kanbanColumn
		if len(columns[col]) > 0 && m.kanbanCursors[col] < len(columns[col]) {
			task := columns[col][m.kanbanCursors[col]]
			return m, m.deleteTask(task.ID)
		}

	case "v":
		m.mode = viewModeList

	case "?", "f1":
		m.previousMode = m.mode
		m.mode = viewModeHelp
	}

	return m, nil
}
```

**Step 3: Add advanceTaskStatus method**

File: `internal/app/app.go` (add method)

```go
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
```

**Step 4: Update list mode 'v' key handler**

File: `internal/app/app.go` (add to list mode key handlers)

```go
		case "v":
			m.mode = viewModeKanban
```

**Step 5: Build and test**

Run: `go build ./cmd/task && ./bin/task`

Test:
1. Press 'v' to switch to kanban view
2. Use h/l to move between columns
3. Use j/k to navigate within columns
4. Press Enter to advance task status
5. Press 'v' to return to list view

**Step 6: Commit**

```bash
git add internal/app/app.go
git commit -m "feat: add kanban navigation and task advancement

Implement kanban view interactions:
- h/l for column navigation
- j/k for vertical navigation within column
- Enter to advance task to next status
- n for new task, d for delete
- v to toggle back to list view
- ? for help modal

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 5: Filter Model

**Files:**
- Create: `internal/domain/filter.go`
- Create: `internal/domain/filter_test.go`

**Step 1: Write failing test for Filter**

File: `internal/domain/filter_test.go`

```go
package domain

import (
	"testing"
	"time"
)

func TestFilter_Match(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	yesterday := now.Add(-24 * time.Hour)

	tests := []struct {
		name   string
		filter Filter
		task   Task
		want   bool
	}{
		{
			name:   "empty filter matches all",
			filter: Filter{},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: true,
		},
		{
			name: "status filter matches",
			filter: Filter{
				Statuses: []TaskStatus{TaskStatusNew},
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: true,
		},
		{
			name: "status filter excludes",
			filter: Filter{
				Statuses: []TaskStatus{TaskStatusWorking},
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: false,
		},
		{
			name: "priority filter matches",
			filter: Filter{
				Priorities: []Priority{PriorityHigh, PriorityMedium},
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: true,
		},
		{
			name: "priority filter excludes",
			filter: Filter{
				Priorities: []Priority{PriorityHigh},
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityLow,
			},
			want: false,
		},
		{
			name: "search text matches title",
			filter: Filter{
				SearchText: "test",
			},
			task: Task{
				Title:    "Test Task",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: true,
		},
		{
			name: "search text matches description",
			filter: Filter{
				SearchText: "important",
			},
			task: Task{
				Title:       "Task",
				Description: "This is important",
				Status:      TaskStatusNew,
				Priority:    PriorityMedium,
			},
			want: true,
		},
		{
			name: "search text no match",
			filter: Filter{
				SearchText: "xyz",
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: false,
		},
		{
			name: "overdue filter matches",
			filter: Filter{
				DateRange: DateRangeOverdue,
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
				DueDate:  &yesterday,
			},
			want: true,
		},
		{
			name: "overdue filter excludes future",
			filter: Filter{
				DateRange: DateRangeOverdue,
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
				DueDate:  &tomorrow,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Match(&tt.task)
			if got != tt.want {
				t.Errorf("Filter.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_IsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		filter Filter
		want   bool
	}{
		{
			name:   "empty filter",
			filter: Filter{},
			want:   true,
		},
		{
			name: "filter with status",
			filter: Filter{
				Statuses: []TaskStatus{TaskStatusNew},
			},
			want: false,
		},
		{
			name: "filter with search text",
			filter: Filter{
				SearchText: "test",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.IsEmpty()
			if got != tt.want {
				t.Errorf("Filter.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/domain -v -run TestFilter`

Expected: Compilation error - Filter type not defined

**Step 3: Write Filter implementation**

File: `internal/domain/filter.go`

```go
package domain

import (
	"strings"
	"time"
)

// DateRange represents a date range filter option
type DateRange int

const (
	DateRangeAll DateRange = iota
	DateRangeToday
	DateRangeThisWeek
	DateRangeOverdue
	DateRangeNoDueDate
)

// Filter represents task filtering criteria
type Filter struct {
	Statuses   []TaskStatus
	Priorities []Priority
	Categories []int64
	DateRange  DateRange
	SearchText string
}

// IsEmpty returns true if no filter criteria are set
func (f *Filter) IsEmpty() bool {
	return len(f.Statuses) == 0 &&
		len(f.Priorities) == 0 &&
		len(f.Categories) == 0 &&
		f.DateRange == DateRangeAll &&
		f.SearchText == ""
}

// Match returns true if the task matches all filter criteria
func (f *Filter) Match(task *Task) bool {
	// Empty filter matches everything
	if f.IsEmpty() {
		return true
	}

	// Check status
	if len(f.Statuses) > 0 {
		found := false
		for _, s := range f.Statuses {
			if task.Status == s {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check priority
	if len(f.Priorities) > 0 {
		found := false
		for _, p := range f.Priorities {
			if task.Priority == p {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check category
	if len(f.Categories) > 0 {
		if task.CategoryID == nil {
			return false
		}
		found := false
		for _, c := range f.Categories {
			if *task.CategoryID == c {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check date range
	if f.DateRange != DateRangeAll {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		switch f.DateRange {
		case DateRangeToday:
			if task.DueDate == nil {
				return false
			}
			dueDay := time.Date(task.DueDate.Year(), task.DueDate.Month(), task.DueDate.Day(), 0, 0, 0, 0, task.DueDate.Location())
			if !dueDay.Equal(today) {
				return false
			}
		case DateRangeThisWeek:
			if task.DueDate == nil {
				return false
			}
			weekEnd := today.AddDate(0, 0, 7)
			if task.DueDate.Before(today) || task.DueDate.After(weekEnd) {
				return false
			}
		case DateRangeOverdue:
			if task.DueDate == nil {
				return false
			}
			if !task.DueDate.Before(today) {
				return false
			}
		case DateRangeNoDueDate:
			if task.DueDate != nil {
				return false
			}
		}
	}

	// Check search text
	if f.SearchText != "" {
		searchLower := strings.ToLower(f.SearchText)
		titleLower := strings.ToLower(task.Title)
		descLower := strings.ToLower(task.Description)
		if !strings.Contains(titleLower, searchLower) && !strings.Contains(descLower, searchLower) {
			return false
		}
	}

	return true
}

// Apply filters a slice of tasks
func (f *Filter) Apply(tasks []*Task) []*Task {
	if f.IsEmpty() {
		return tasks
	}

	var result []*Task
	for _, task := range tasks {
		if f.Match(task) {
			result = append(result, task)
		}
	}
	return result
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/domain -v -run TestFilter`

Expected: PASS

**Step 5: Commit**

```bash
git add internal/domain/filter.go internal/domain/filter_test.go
git commit -m "feat: add filter model with matching logic

Implement task filtering:
- Filter by status (multiple selection)
- Filter by priority (multiple selection)
- Filter by category (multiple selection)
- Filter by date range (today, this week, overdue, no due date)
- Full-text search on title and description
- Comprehensive unit tests

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 6: Sort Model

**Files:**
- Create: `internal/domain/sort.go`
- Create: `internal/domain/sort_test.go`

**Step 1: Write failing test for Sort**

File: `internal/domain/sort_test.go`

```go
package domain

import (
	"testing"
	"time"
)

func TestSortTasks(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	tasks := []*Task{
		{ID: 1, Title: "B Task", Priority: PriorityMedium, CreatedAt: now, DueDate: &tomorrow},
		{ID: 2, Title: "A Task", Priority: PriorityHigh, CreatedAt: yesterday, DueDate: &now},
		{ID: 3, Title: "C Task", Priority: PriorityLow, CreatedAt: tomorrow, DueDate: &yesterday},
	}

	tests := []struct {
		name      string
		sortBy    SortBy
		ascending bool
		wantOrder []int64
	}{
		{
			name:      "sort by created desc",
			sortBy:    SortByCreatedAt,
			ascending: false,
			wantOrder: []int64{3, 1, 2},
		},
		{
			name:      "sort by created asc",
			sortBy:    SortByCreatedAt,
			ascending: true,
			wantOrder: []int64{2, 1, 3},
		},
		{
			name:      "sort by priority desc (high first)",
			sortBy:    SortByPriority,
			ascending: false,
			wantOrder: []int64{2, 1, 3},
		},
		{
			name:      "sort by due date asc",
			sortBy:    SortByDueDate,
			ascending: true,
			wantOrder: []int64{3, 2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid mutating original
			tasksCopy := make([]*Task, len(tasks))
			copy(tasksCopy, tasks)

			sort := Sort{By: tt.sortBy, Ascending: tt.ascending}
			result := sort.Apply(tasksCopy)

			for i, wantID := range tt.wantOrder {
				if result[i].ID != wantID {
					t.Errorf("position %d: got ID %d, want %d", i, result[i].ID, wantID)
				}
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/domain -v -run TestSortTasks`

Expected: Compilation error - SortBy type not defined

**Step 3: Write Sort implementation**

File: `internal/domain/sort.go`

```go
package domain

import "sort"

// SortBy represents the field to sort by
type SortBy int

const (
	SortByCreatedAt SortBy = iota
	SortByDueDate
	SortByPriority
	SortByStatus
	SortByTitle
)

// Sort represents sorting criteria
type Sort struct {
	By        SortBy
	Ascending bool
}

// Apply sorts a slice of tasks
func (s *Sort) Apply(tasks []*Task) []*Task {
	result := make([]*Task, len(tasks))
	copy(result, tasks)

	sort.Slice(result, func(i, j int) bool {
		var less bool

		switch s.By {
		case SortByCreatedAt:
			less = result[i].CreatedAt.Before(result[j].CreatedAt)
		case SortByDueDate:
			// nil due dates go last
			if result[i].DueDate == nil && result[j].DueDate == nil {
				less = false
			} else if result[i].DueDate == nil {
				less = false
			} else if result[j].DueDate == nil {
				less = true
			} else {
				less = result[i].DueDate.Before(*result[j].DueDate)
			}
		case SortByPriority:
			// High > Medium > Low
			pi := priorityOrder(result[i].Priority)
			pj := priorityOrder(result[j].Priority)
			less = pi < pj
		case SortByStatus:
			si := statusOrder(result[i].Status)
			sj := statusOrder(result[j].Status)
			less = si < sj
		case SortByTitle:
			less = result[i].Title < result[j].Title
		}

		if s.Ascending {
			return less
		}
		return !less
	})

	return result
}

func priorityOrder(p Priority) int {
	switch p {
	case PriorityHigh:
		return 0
	case PriorityMedium:
		return 1
	case PriorityLow:
		return 2
	default:
		return 3
	}
}

func statusOrder(s TaskStatus) int {
	switch s {
	case TaskStatusNew:
		return 0
	case TaskStatusWorking:
		return 1
	case TaskStatusCompleted:
		return 2
	default:
		return 3
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/domain -v -run TestSortTasks`

Expected: PASS

**Step 5: Commit**

```bash
git add internal/domain/sort.go internal/domain/sort_test.go
git commit -m "feat: add sort model

Implement task sorting:
- Sort by created date, due date, priority, status, title
- Ascending/descending order
- Nil due dates sorted last
- Priority ordered High > Medium > Low
- Unit tests for all sort options

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 7: Filter UI

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add filter state to Model**

File: `internal/app/app.go` (modify Model struct)

```go
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
	// Kanban view state
	kanbanColumn  int
	kanbanCursors [3]int
	// Filter state
	filter       domain.Filter
	filterCursor int
}
```

**Step 2: Add 'f' key handler to list and kanban modes**

File: `internal/app/app.go` (add to list mode key handlers)

```go
		case "f":
			m.previousMode = m.mode
			m.mode = viewModeFilter
			m.filterCursor = 0
```

Also add to updateKanbanMode:

```go
		case "f":
			m.previousMode = m.mode
			m.mode = viewModeFilter
			m.filterCursor = 0
```

**Step 3: Add viewFilter method**

File: `internal/app/app.go` (add method)

```go
func (m *Model) viewFilter() string {
	s := "┌─ フィルタ設定 ─────────────────┐\n"

	// Status filter
	s += "│ ステータス:                    │\n"
	statuses := []struct {
		status domain.TaskStatus
		label  string
	}{
		{domain.TaskStatusNew, "New"},
		{domain.TaskStatusWorking, "Working"},
		{domain.TaskStatusCompleted, "Completed"},
	}
	for i, st := range statuses {
		checked := " "
		for _, fs := range m.filter.Statuses {
			if fs == st.status {
				checked = "✓"
				break
			}
		}
		line := fmt.Sprintf("│  [%s] %-24s│\n", checked, st.label)
		if m.filterCursor == i {
			line = styles.Selected.Render(line)
		}
		s += line
	}

	s += "│                                │\n"

	// Priority filter
	s += "│ 優先度:                        │\n"
	priorities := []struct {
		priority domain.Priority
		label    string
	}{
		{domain.PriorityHigh, "高"},
		{domain.PriorityMedium, "中"},
		{domain.PriorityLow, "低"},
	}
	for i, pr := range priorities {
		checked := " "
		for _, fp := range m.filter.Priorities {
			if fp == pr.priority {
				checked = "✓"
				break
			}
		}
		line := fmt.Sprintf("│  [%s] %-24s│\n", checked, pr.label)
		if m.filterCursor == 3+i {
			line = styles.Selected.Render(line)
		}
		s += line
	}

	s += "│                                │\n"

	// Date range filter
	s += "│ 期限:                          │\n"
	dateRanges := []struct {
		dateRange domain.DateRange
		label     string
	}{
		{domain.DateRangeAll, "すべて"},
		{domain.DateRangeToday, "今日"},
		{domain.DateRangeThisWeek, "今週"},
		{domain.DateRangeOverdue, "期限切れ"},
		{domain.DateRangeNoDueDate, "期限なし"},
	}
	for i, dr := range dateRanges {
		selected := " "
		if m.filter.DateRange == dr.dateRange {
			selected = "●"
		}
		line := fmt.Sprintf("│  (%s) %-24s│\n", selected, dr.label)
		if m.filterCursor == 6+i {
			line = styles.Selected.Render(line)
		}
		s += line
	}

	s += "│                                │\n"

	// Search text
	searchLine := fmt.Sprintf("│ 検索: [%-21s] │\n", m.filter.SearchText+"_")
	if m.filterCursor == 11 {
		searchLine = styles.Selected.Render(searchLine)
	}
	s += searchLine

	s += "│                                │\n"

	// Actions
	clearLine := "│     [クリア]                   │\n"
	if m.filterCursor == 12 {
		clearLine = styles.Selected.Render(clearLine)
	}
	s += clearLine

	s += "└────────────────────────────────┘\n"

	helpText := "[j/k]移動 [Space]選択 [Enter]適用 [Esc]キャンセル"
	s += "\n" + styles.StatusBar.Render(helpText) + "\n"

	return s
}
```

**Step 4: Add updateFilterMode method**

File: `internal/app/app.go` (add method)

```go
func (m *Model) updateFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxCursor := 12 // 3 statuses + 3 priorities + 5 date ranges + search + clear

	switch msg.String() {
	case "esc":
		m.mode = m.previousMode

	case "enter":
		m.mode = m.previousMode

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
		case m.filterCursor < 3:
			// Status toggle
			status := []domain.TaskStatus{domain.TaskStatusNew, domain.TaskStatusWorking, domain.TaskStatusCompleted}[m.filterCursor]
			m.toggleFilterStatus(status)
		case m.filterCursor < 6:
			// Priority toggle
			priority := []domain.Priority{domain.PriorityHigh, domain.PriorityMedium, domain.PriorityLow}[m.filterCursor-3]
			m.toggleFilterPriority(priority)
		case m.filterCursor < 11:
			// Date range selection (radio)
			dateRange := []domain.DateRange{domain.DateRangeAll, domain.DateRangeToday, domain.DateRangeThisWeek, domain.DateRangeOverdue, domain.DateRangeNoDueDate}[m.filterCursor-6]
			m.filter.DateRange = dateRange
		case m.filterCursor == 12:
			// Clear filter
			m.filter = domain.Filter{}
		}

	case "backspace":
		if m.filterCursor == 11 && len(m.filter.SearchText) > 0 {
			m.filter.SearchText = m.filter.SearchText[:len(m.filter.SearchText)-1]
		}

	default:
		// Add character to search if on search field
		if m.filterCursor == 11 {
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

func (m *Model) toggleFilterStatus(status domain.TaskStatus) {
	for i, s := range m.filter.Statuses {
		if s == status {
			m.filter.Statuses = append(m.filter.Statuses[:i], m.filter.Statuses[i+1:]...)
			return
		}
	}
	m.filter.Statuses = append(m.filter.Statuses, status)
}

func (m *Model) toggleFilterPriority(priority domain.Priority) {
	for i, p := range m.filter.Priorities {
		if p == priority {
			m.filter.Priorities = append(m.filter.Priorities[:i], m.filter.Priorities[i+1:]...)
			return
		}
	}
	m.filter.Priorities = append(m.filter.Priorities, priority)
}
```

**Step 5: Update Update method to handle filter mode**

File: `internal/app/app.go` (add after help mode check)

```go
		// Handle filter mode
		if m.mode == viewModeFilter {
			return m.updateFilterMode(msg)
		}
```

**Step 6: Update View method**

File: `internal/app/app.go` (add to View method before kanban check)

```go
	// Filter modal
	if m.mode == viewModeFilter {
		return m.viewFilter()
	}
```

**Step 7: Apply filter to task lists**

File: `internal/app/app.go` (modify tasksByStatus)

```go
func (m *Model) tasksByStatus(status domain.TaskStatus) []*domain.Task {
	filtered := m.filter.Apply(m.tasks)
	var result []*domain.Task
	for _, task := range filtered {
		if task.Status == status {
			result = append(result, task)
		}
	}
	return result
}
```

Also update viewList to use filtered tasks:

```go
func (m *Model) viewList() string {
	filteredTasks := m.filter.Apply(m.tasks)

	s := "Task Management"
	if !m.filter.IsEmpty() {
		s += " (フィルタ適用中)"
	}
	s += "\n\n"

	if len(filteredTasks) == 0 {
		if len(m.tasks) == 0 {
			s += "No tasks yet. Press 'n' to create one.\n\n"
		} else {
			s += "フィルタに一致するタスクがありません。\n\n"
		}
	} else {
		for i, task := range filteredTasks {
			// ... existing rendering code, but use filteredTasks
```

**Step 8: Build and test**

Run: `go build ./cmd/task && ./bin/task`

Test:
1. Press 'f' to open filter
2. Navigate with j/k
3. Toggle options with Space
4. Press Enter to apply
5. Verify filtered results

**Step 9: Commit**

```bash
git add internal/app/app.go
git commit -m "feat: add filter UI

Implement filter modal:
- Multi-select for status and priority
- Radio buttons for date range
- Text search field
- Clear filter option
- Filter indicator in view header
- Filters apply to both list and kanban views

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 8: Sort UI

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add sort state to Model**

File: `internal/app/app.go` (modify Model struct, add after filter)

```go
	// Sort state
	taskSort     domain.Sort
	sortMenuOpen bool
```

**Step 2: Add 's' key handler**

File: `internal/app/app.go` (add to list mode key handlers)

```go
		case "s":
			m.sortMenuOpen = !m.sortMenuOpen
```

Also add to updateKanbanMode.

**Step 3: Add sort menu to views**

File: `internal/app/app.go` (add method)

```go
func (m *Model) viewSortMenu() string {
	s := "┌─ ソート ────────────┐\n"

	options := []struct {
		by    domain.SortBy
		label string
	}{
		{domain.SortByCreatedAt, "作成日"},
		{domain.SortByDueDate, "期限"},
		{domain.SortByPriority, "優先度"},
		{domain.SortByStatus, "ステータス"},
		{domain.SortByTitle, "タイトル"},
	}

	for _, opt := range options {
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
		s += fmt.Sprintf("│ (%s) %-14s%s│\n", selected, opt.label, order)
	}

	s += "└─────────────────────┘\n"
	return s
}
```

**Step 4: Integrate sort into views**

File: `internal/app/app.go` (modify viewList to show sort menu when open and apply sort)

```go
func (m *Model) viewList() string {
	filteredTasks := m.filter.Apply(m.tasks)
	sortedTasks := m.taskSort.Apply(filteredTasks)

	s := "Task Management"
	if !m.filter.IsEmpty() {
		s += " (フィルタ適用中)"
	}
	s += "\n\n"

	if m.sortMenuOpen {
		s += m.viewSortMenu() + "\n"
	}

	// ... rest uses sortedTasks instead of filteredTasks
```

**Step 5: Add sort key handling**

File: `internal/app/app.go` (add sort selection when menu is open)

In Update method for list mode, add number key handling when sortMenuOpen:

```go
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
```

**Step 6: Update status bar**

File: `internal/app/app.go` (modify viewList status bar)

```go
	// Status bar
	helpText := "[n]新規 [d]削除 [Space]ステータス [f]フィルタ [s]ソート [v]カンバン [?]ヘルプ [q]終了"
	s += styles.StatusBar.Render(helpText) + "\n"
```

**Step 7: Build and test**

Run: `go build ./cmd/task && ./bin/task`

Test:
1. Press 's' to open sort menu
2. Press 1-5 to select sort option
3. Press same number again to toggle direction
4. Verify sorted results

**Step 8: Commit**

```bash
git add internal/app/app.go
git commit -m "feat: add sort UI

Implement sort menu:
- Sort by created date, due date, priority, status, title
- Toggle ascending/descending with repeated selection
- Visual indicator for current sort option and direction
- Sort applies to both list and kanban views

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 9: Final Integration Test

**Step 1: Run all tests**

Run: `make test`

Expected: All tests pass

**Step 2: Build application**

Run: `make build`

Expected: Binary created at `bin/task`

**Step 3: Manual integration test**

Run: `./bin/task`

Test flow:
1. Create 5+ tasks with different priorities
2. Press 'v' to switch to kanban view
3. Use h/l/j/k to navigate
4. Press Enter to advance task status
5. Press 'v' to return to list view
6. Press 'f' to open filter
7. Select status filter, apply
8. Verify filtered results
9. Press 's' to sort by priority
10. Verify sorted results
11. Press '?' for help
12. Press 'q' to quit

**Step 4: Commit**

```bash
git add -A
git commit -m "test: verify Phase 2 complete integration

Phase 2 UI enhancements complete:
✅ View mode infrastructure
✅ Help modal (context-aware)
✅ Kanban view with navigation
✅ Task status advancement from kanban
✅ Filter model with matching logic
✅ Sort model with multiple options
✅ Filter UI modal
✅ Sort menu

All unit tests passing.
Manual integration tests verified.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Summary

Phase 2 implementation is complete when:

- ✅ All unit tests pass (`make test`)
- ✅ Application builds successfully (`make build`)
- ✅ Manual testing confirms:
  - View switching (list ↔ kanban) works
  - Kanban navigation and task advancement works
  - Filter modal works with all filter types
  - Sort menu works with direction toggle
  - Help modal shows context-aware content
- ✅ All changes committed to git

**Next Steps:** Phase 3 will add data enhancements (task editing, categories, due dates).
