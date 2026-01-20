# Phase 1: Core Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement core task management functionality with SQLite storage and basic TUI list view

**Architecture:** Domain-driven design with repository pattern for data access. Bubble Tea framework for TUI. Test-first development with comprehensive unit tests for domain and repository layers.

**Tech Stack:** Go 1.21+, modernc.org/sqlite (pure Go), charmbracelet/bubbletea

---

## Task 1: Project Structure Setup

**Files:**
- Create: `cmd/task/main.go`
- Create: `internal/domain/task.go`
- Create: `internal/domain/category.go`
- Create: `internal/domain/repository.go`
- Create: `internal/repository/sqlite.go`
- Create: `internal/repository/migrations.go`
- Create: `internal/app/app.go`
- Create: `internal/app/messages.go`
- Create: `internal/ui/views/tasks_list.go`
- Create: `internal/ui/styles/styles.go`
- Create: `Makefile`

**Step 1: Create directory structure**

```bash
mkdir -p cmd/task
mkdir -p internal/domain
mkdir -p internal/repository
mkdir -p internal/app
mkdir -p internal/ui/views
mkdir -p internal/ui/styles
mkdir -p internal/ui/components
```

**Step 2: Create Makefile**

File: `Makefile`

```makefile
.PHONY: build test run clean install

build:
	go build -o bin/task ./cmd/task

install:
	go install ./cmd/task

test:
	go test -v -race -cover ./...

run:
	go run ./cmd/task

clean:
	rm -rf bin/
```

**Step 3: Commit**

```bash
git add -A
git commit -m "chore: setup project structure

Create directory layout following domain-driven design:
- cmd/task: entry point
- internal/domain: business logic and models
- internal/repository: data access layer
- internal/app: application orchestration
- internal/ui: TUI components and views

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 2: Domain Models

**Files:**
- Create: `internal/domain/task.go`
- Create: `internal/domain/task_test.go`
- Create: `internal/domain/category.go`
- Create: `internal/domain/category_test.go`

**Step 1: Write failing test for TaskStatus**

File: `internal/domain/task_test.go`

```go
package domain

import "testing"

func TestTaskStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   string
	}{
		{"new status", TaskStatusNew, "new"},
		{"working status", TaskStatusWorking, "working"},
		{"completed status", TaskStatusCompleted, "completed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("TaskStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   bool
	}{
		{"new is valid", TaskStatusNew, true},
		{"working is valid", TaskStatusWorking, true},
		{"completed is valid", TaskStatusCompleted, true},
		{"invalid status", TaskStatus("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("TaskStatus.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/domain -v`

Expected: Compilation error - types not defined

**Step 3: Write minimal TaskStatus implementation**

File: `internal/domain/task.go`

```go
package domain

import "time"

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusNew       TaskStatus = "new"
	TaskStatusWorking   TaskStatus = "working"
	TaskStatusCompleted TaskStatus = "completed"
)

// String returns the string representation of TaskStatus
func (s TaskStatus) String() string {
	return string(s)
}

// IsValid checks if the TaskStatus is valid
func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusNew, TaskStatusWorking, TaskStatusCompleted:
		return true
	default:
		return false
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/domain -v`

Expected: PASS

**Step 5: Write failing test for Priority**

File: `internal/domain/task_test.go` (append)

```go
func TestPriority_String(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		want     string
	}{
		{"low priority", PriorityLow, "low"},
		{"medium priority", PriorityMedium, "medium"},
		{"high priority", PriorityHigh, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priority.String(); got != tt.want {
				t.Errorf("Priority.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriority_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		want     bool
	}{
		{"low is valid", PriorityLow, true},
		{"medium is valid", PriorityMedium, true},
		{"high is valid", PriorityHigh, true},
		{"invalid priority", Priority("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priority.IsValid(); got != tt.want {
				t.Errorf("Priority.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**Step 6: Run test to verify it fails**

Run: `go test ./internal/domain -v`

Expected: Compilation error - Priority type not defined

**Step 7: Write minimal Priority implementation**

File: `internal/domain/task.go` (append)

```go
// Priority represents the priority level of a task
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// String returns the string representation of Priority
func (p Priority) String() string {
	return string(p)
}

// IsValid checks if the Priority is valid
func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}
```

**Step 8: Run test to verify it passes**

Run: `go test ./internal/domain -v`

Expected: PASS

**Step 9: Write failing test for Task validation**

File: `internal/domain/task_test.go` (append)

```go
func TestTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid task",
			task: Task{
				Title:     "Test task",
				Status:    TaskStatusNew,
				Priority:  PriorityMedium,
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty title",
			task: Task{
				Title:     "",
				Status:    TaskStatusNew,
				Priority:  PriorityMedium,
				CreatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "title too long",
			task: Task{
				Title:     string(make([]byte, 201)),
				Status:    TaskStatusNew,
				Priority:  PriorityMedium,
				CreatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "title must be 200 characters or less",
		},
		{
			name: "invalid status",
			task: Task{
				Title:     "Test",
				Status:    TaskStatus("invalid"),
				Priority:  PriorityMedium,
				CreatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name: "invalid priority",
			task: Task{
				Title:     "Test",
				Status:    TaskStatusNew,
				Priority:  Priority("invalid"),
				CreatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "invalid priority",
		},
		{
			name: "description too long",
			task: Task{
				Title:       "Test",
				Description: string(make([]byte, 1001)),
				Status:      TaskStatusNew,
				Priority:    PriorityMedium,
				CreatedAt:   time.Now(),
			},
			wantErr: true,
			errMsg:  "description must be 1000 characters or less",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Task.Validate() error = nil, want error containing %q", tt.errMsg)
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Task.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Task.Validate() error = %v, want nil", err)
				}
			}
		})
	}
}
```

**Step 10: Run test to verify it fails**

Run: `go test ./internal/domain -v`

Expected: Compilation error - Task type and Validate method not defined

**Step 11: Write minimal Task struct and Validate implementation**

File: `internal/domain/task.go` (append)

```go
import (
	"errors"
	"time"
)

// Task represents a task in the system
type Task struct {
	ID          int64
	Title       string
	Description string
	Status      TaskStatus
	Priority    Priority
	CategoryID  *int64
	DueDate     *time.Time
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}

// Validate checks if the task is valid
func (t *Task) Validate() error {
	if t.Title == "" {
		return errors.New("title is required")
	}
	if len(t.Title) > 200 {
		return errors.New("title must be 200 characters or less")
	}
	if len(t.Description) > 1000 {
		return errors.New("description must be 1000 characters or less")
	}
	if !t.Status.IsValid() {
		return errors.New("invalid status")
	}
	if !t.Priority.IsValid() {
		return errors.New("invalid priority")
	}
	return nil
}
```

**Step 12: Run test to verify it passes**

Run: `go test ./internal/domain -v`

Expected: PASS

**Step 13: Write failing test for Category validation**

File: `internal/domain/category_test.go`

```go
package domain

import "testing"

func TestCategory_Validate(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid category",
			category: Category{
				Name:  "Work",
				Color: "blue",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			category: Category{
				Name:  "",
				Color: "blue",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "empty color",
			category: Category{
				Name:  "Work",
				Color: "",
			},
			wantErr: true,
			errMsg:  "color is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.category.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Category.Validate() error = nil, want error containing %q", tt.errMsg)
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Category.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Category.Validate() error = %v, want nil", err)
				}
			}
		})
	}
}
```

**Step 14: Run test to verify it fails**

Run: `go test ./internal/domain -v`

Expected: Compilation error - Category type not defined

**Step 15: Write minimal Category implementation**

File: `internal/domain/category.go`

```go
package domain

import (
	"errors"
	"time"
)

// Category represents a task category
type Category struct {
	ID        int64
	Name      string
	Color     string
	CreatedAt time.Time
}

// Validate checks if the category is valid
func (c *Category) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Color == "" {
		return errors.New("color is required")
	}
	return nil
}
```

**Step 16: Run test to verify it passes**

Run: `go test ./internal/domain -v`

Expected: PASS

**Step 17: Commit**

```bash
git add internal/domain/
git commit -m "feat: add domain models with validation

Add Task and Category domain models with:
- TaskStatus enum (new, working, completed)
- Priority enum (low, medium, high)
- Validation methods for all models
- Comprehensive unit tests

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 3: Repository Interface

**Files:**
- Create: `internal/domain/repository.go`

**Step 1: Write repository interface**

File: `internal/domain/repository.go`

```go
package domain

import "context"

// TaskRepository defines the interface for task storage operations
type TaskRepository interface {
	// Create creates a new task
	Create(ctx context.Context, task *Task) error

	// Update updates an existing task
	Update(ctx context.Context, task *Task) error

	// Delete deletes a task by ID
	Delete(ctx context.Context, id int64) error

	// GetByID retrieves a task by ID
	GetByID(ctx context.Context, id int64) (*Task, error)

	// List retrieves all tasks
	List(ctx context.Context) ([]*Task, error)

	// CreateCategory creates a new category
	CreateCategory(ctx context.Context, category *Category) error

	// GetCategories retrieves all categories
	GetCategories(ctx context.Context) ([]*Category, error)

	// Close closes the repository connection
	Close() error
}
```

**Step 2: Commit**

```bash
git add internal/domain/repository.go
git commit -m "feat: add repository interface

Define TaskRepository interface for data access operations:
- CRUD operations for tasks
- Category management
- Context support for cancellation

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 4: SQLite Repository - Migrations

**Files:**
- Create: `internal/repository/migrations.go`
- Create: `internal/repository/migrations_test.go`

**Step 1: Add SQLite dependency**

Run: `go get modernc.org/sqlite`

**Step 2: Write failing test for migrations**

File: `internal/repository/migrations_test.go`

```go
package repository

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestRunMigrations(t *testing.T) {
	// Use in-memory database for testing
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("runMigrations() error = %v", err)
	}

	// Verify categories table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='categories'").Scan(&tableName)
	if err != nil {
		t.Errorf("categories table not found: %v", err)
	}
	if tableName != "categories" {
		t.Errorf("expected table name 'categories', got %q", tableName)
	}

	// Verify tasks table exists
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='tasks'").Scan(&tableName)
	if err != nil {
		t.Errorf("tasks table not found: %v", err)
	}
	if tableName != "tasks" {
		t.Errorf("expected table name 'tasks', got %q", tableName)
	}
}

func TestRunMigrations_CreatesDefaultCategories(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		t.Fatalf("runMigrations() error = %v", err)
	}

	// Count categories
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count categories: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 default categories, got %d", count)
	}
}
```

**Step 3: Run test to verify it fails**

Run: `go test ./internal/repository -v`

Expected: Compilation error - runMigrations function not defined

**Step 4: Write minimal migrations implementation**

File: `internal/repository/migrations.go`

```go
package repository

import (
	"database/sql"
	"time"
)

// runMigrations executes database migrations
func runMigrations(db *sql.DB) error {
	// Create categories table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			color TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create tasks table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL CHECK(status IN ('new', 'working', 'completed')),
			priority TEXT NOT NULL CHECK(priority IN ('low', 'medium', 'high')),
			category_id INTEGER,
			due_date DATETIME,
			created_at DATETIME NOT NULL,
			started_at DATETIME,
			completed_at DATETIME,
			FOREIGN KEY (category_id) REFERENCES categories(id)
		)
	`)
	if err != nil {
		return err
	}

	// Insert default categories if none exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		now := time.Now().Format(time.RFC3339)
		defaultCategories := []struct {
			name  string
			color string
		}{
			{"仕事", "blue"},
			{"個人", "green"},
			{"その他", "yellow"},
		}

		for _, cat := range defaultCategories {
			_, err = db.Exec(
				"INSERT INTO categories (name, color, created_at) VALUES (?, ?, ?)",
				cat.name, cat.color, now,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
```

**Step 5: Run test to verify it passes**

Run: `go test ./internal/repository -v`

Expected: PASS

**Step 6: Commit**

```bash
git add internal/repository/migrations.go internal/repository/migrations_test.go go.mod go.sum
git commit -m "feat: add database migrations

Implement SQLite schema migrations:
- Create categories and tasks tables
- Add constraints and foreign keys
- Insert default categories
- Test with in-memory database

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 5: SQLite Repository - Implementation

**Files:**
- Create: `internal/repository/sqlite.go`
- Create: `internal/repository/sqlite_test.go`

**Step 1: Write failing test for NewSQLiteRepository**

File: `internal/repository/sqlite_test.go`

```go
package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewSQLiteRepository(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create repository
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	defer repo.Close()

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("database file was not created at %s", dbPath)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/repository -v -run TestNewSQLiteRepository`

Expected: Compilation error - NewSQLiteRepository not defined

**Step 3: Write minimal NewSQLiteRepository implementation**

File: `internal/repository/sqlite.go`

```go
package repository

import (
	"context"
	"database/sql"
	"path/filepath"
	"os"

	"github.com/hitsumabushi845/task-management/internal/domain"
	_ "modernc.org/sqlite"
)

// SQLiteRepository implements TaskRepository using SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteRepository{db: db}, nil
}

// Close closes the database connection
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/repository -v -run TestNewSQLiteRepository`

Expected: PASS

**Step 5: Write failing test for Create**

File: `internal/repository/sqlite_test.go` (append)

```go
func TestSQLiteRepository_Create(t *testing.T) {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	task := &domain.Task{
		Title:       "Test task",
		Description: "Test description",
		Status:      domain.TaskStatusNew,
		Priority:    domain.PriorityMedium,
	}

	err = repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify ID was set
	if task.ID == 0 {
		t.Errorf("Create() did not set task ID")
	}

	// Verify CreatedAt was set
	if task.CreatedAt.IsZero() {
		t.Errorf("Create() did not set CreatedAt")
	}
}
```

**Step 6: Run test to verify it fails**

Run: `go test ./internal/repository -v -run TestSQLiteRepository_Create`

Expected: Compilation error - Create method not defined

**Step 7: Write minimal Create implementation**

File: `internal/repository/sqlite.go` (append)

```go
import "time"

// Create creates a new task
func (r *SQLiteRepository) Create(ctx context.Context, task *domain.Task) error {
	if err := task.Validate(); err != nil {
		return err
	}

	now := time.Now()
	task.CreatedAt = now

	result, err := r.db.ExecContext(ctx,
		`INSERT INTO tasks (title, description, status, priority, category_id, due_date, created_at, started_at, completed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.CategoryID,
		task.DueDate,
		task.CreatedAt.Format(time.RFC3339),
		formatTimePtr(task.StartedAt),
		formatTimePtr(task.CompletedAt),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	task.ID = id
	return nil
}

// Helper function to format *time.Time for SQL
func formatTimePtr(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339)
}
```

**Step 8: Run test to verify it passes**

Run: `go test ./internal/repository -v -run TestSQLiteRepository_Create`

Expected: PASS

**Step 9: Write failing test for List**

File: `internal/repository/sqlite_test.go` (append)

```go
func TestSQLiteRepository_List(t *testing.T) {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// Create test tasks
	task1 := &domain.Task{
		Title:    "Task 1",
		Status:   domain.TaskStatusNew,
		Priority: domain.PriorityHigh,
	}
	task2 := &domain.Task{
		Title:    "Task 2",
		Status:   domain.TaskStatusWorking,
		Priority: domain.PriorityLow,
	}

	if err := repo.Create(ctx, task1); err != nil {
		t.Fatalf("Create(task1) error = %v", err)
	}
	if err := repo.Create(ctx, task2); err != nil {
		t.Fatalf("Create(task2) error = %v", err)
	}

	// List all tasks
	tasks, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("List() returned %d tasks, want 2", len(tasks))
	}

	// Verify task data
	if tasks[0].Title != "Task 1" {
		t.Errorf("tasks[0].Title = %q, want %q", tasks[0].Title, "Task 1")
	}
	if tasks[1].Title != "Task 2" {
		t.Errorf("tasks[1].Title = %q, want %q", tasks[1].Title, "Task 2")
	}
}
```

**Step 10: Run test to verify it fails**

Run: `go test ./internal/repository -v -run TestSQLiteRepository_List`

Expected: Compilation error - List method not defined

**Step 11: Write minimal List implementation**

File: `internal/repository/sqlite.go` (append)

```go
// List retrieves all tasks
func (r *SQLiteRepository) List(ctx context.Context) ([]*domain.Task, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, description, status, priority, category_id, due_date, created_at, started_at, completed_at
		 FROM tasks
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		var createdAt, startedAt, completedAt, dueDate sql.NullString
		var categoryID sql.NullInt64

		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&categoryID,
			&dueDate,
			&createdAt,
			&startedAt,
			&completedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse timestamps
		if createdAt.Valid {
			t, _ := time.Parse(time.RFC3339, createdAt.String)
			task.CreatedAt = t
		}
		if startedAt.Valid {
			t, _ := time.Parse(time.RFC3339, startedAt.String)
			task.StartedAt = &t
		}
		if completedAt.Valid {
			t, _ := time.Parse(time.RFC3339, completedAt.String)
			task.CompletedAt = &t
		}
		if dueDate.Valid {
			t, _ := time.Parse(time.RFC3339, dueDate.String)
			task.DueDate = &t
		}
		if categoryID.Valid {
			id := categoryID.Int64
			task.CategoryID = &id
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
```

**Step 12: Run test to verify it passes**

Run: `go test ./internal/repository -v -run TestSQLiteRepository_List`

Expected: PASS

**Step 13: Write failing test for Update**

File: `internal/repository/sqlite_test.go` (append)

```go
func TestSQLiteRepository_Update(t *testing.T) {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// Create a task
	task := &domain.Task{
		Title:    "Original title",
		Status:   domain.TaskStatusNew,
		Priority: domain.PriorityMedium,
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update the task
	task.Title = "Updated title"
	task.Status = domain.TaskStatusWorking
	now := time.Now()
	task.StartedAt = &now

	if err := repo.Update(ctx, task); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Retrieve and verify
	updated, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if updated.Title != "Updated title" {
		t.Errorf("Title = %q, want %q", updated.Title, "Updated title")
	}
	if updated.Status != domain.TaskStatusWorking {
		t.Errorf("Status = %q, want %q", updated.Status, domain.TaskStatusWorking)
	}
	if updated.StartedAt == nil {
		t.Errorf("StartedAt is nil, want non-nil")
	}
}
```

**Step 14: Run test to verify it fails**

Run: `go test ./internal/repository -v -run TestSQLiteRepository_Update`

Expected: Compilation error - Update and GetByID methods not defined

**Step 15: Write minimal Update and GetByID implementations**

File: `internal/repository/sqlite.go` (append)

```go
// Update updates an existing task
func (r *SQLiteRepository) Update(ctx context.Context, task *domain.Task) error {
	if err := task.Validate(); err != nil {
		return err
	}

	_, err := r.db.ExecContext(ctx,
		`UPDATE tasks
		 SET title = ?, description = ?, status = ?, priority = ?, category_id = ?,
		     due_date = ?, started_at = ?, completed_at = ?
		 WHERE id = ?`,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.CategoryID,
		formatTimePtr(task.DueDate),
		formatTimePtr(task.StartedAt),
		formatTimePtr(task.CompletedAt),
		task.ID,
	)
	return err
}

// GetByID retrieves a task by ID
func (r *SQLiteRepository) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	task := &domain.Task{}
	var createdAt, startedAt, completedAt, dueDate sql.NullString
	var categoryID sql.NullInt64

	err := r.db.QueryRowContext(ctx,
		`SELECT id, title, description, status, priority, category_id, due_date, created_at, started_at, completed_at
		 FROM tasks
		 WHERE id = ?`,
		id,
	).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&categoryID,
		&dueDate,
		&createdAt,
		&startedAt,
		&completedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse timestamps
	if createdAt.Valid {
		t, _ := time.Parse(time.RFC3339, createdAt.String)
		task.CreatedAt = t
	}
	if startedAt.Valid {
		t, _ := time.Parse(time.RFC3339, startedAt.String)
		task.StartedAt = &t
	}
	if completedAt.Valid {
		t, _ := time.Parse(time.RFC3339, completedAt.String)
		task.CompletedAt = &t
	}
	if dueDate.Valid {
		t, _ := time.Parse(time.RFC3339, dueDate.String)
		task.DueDate = &t
	}
	if categoryID.Valid {
		id := categoryID.Int64
		task.CategoryID = &id
	}

	return task, nil
}
```

**Step 16: Run test to verify it passes**

Run: `go test ./internal/repository -v -run TestSQLiteRepository_Update`

Expected: PASS

**Step 17: Write failing test for Delete**

File: `internal/repository/sqlite_test.go` (append)

```go
func TestSQLiteRepository_Delete(t *testing.T) {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// Create a task
	task := &domain.Task{
		Title:    "To be deleted",
		Status:   domain.TaskStatusNew,
		Priority: domain.PriorityLow,
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Delete the task
	if err := repo.Delete(ctx, task.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's deleted
	_, err = repo.GetByID(ctx, task.ID)
	if err == nil {
		t.Errorf("GetByID() after Delete should return error, got nil")
	}
}
```

**Step 18: Run test to verify it fails**

Run: `go test ./internal/repository -v -run TestSQLiteRepository_Delete`

Expected: Compilation error - Delete method not defined

**Step 19: Write minimal Delete implementation**

File: `internal/repository/sqlite.go` (append)

```go
// Delete deletes a task by ID
func (r *SQLiteRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	return err
}
```

**Step 20: Run test to verify it passes**

Run: `go test ./internal/repository -v -run TestSQLiteRepository_Delete`

Expected: PASS

**Step 21: Write failing tests for category operations**

File: `internal/repository/sqlite_test.go` (append)

```go
func TestSQLiteRepository_CreateCategory(t *testing.T) {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	category := &domain.Category{
		Name:  "Test Category",
		Color: "red",
	}

	err = repo.CreateCategory(ctx, category)
	if err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}

	if category.ID == 0 {
		t.Errorf("CreateCategory() did not set category ID")
	}

	if category.CreatedAt.IsZero() {
		t.Errorf("CreateCategory() did not set CreatedAt")
	}
}

func TestSQLiteRepository_GetCategories(t *testing.T) {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// Should have 3 default categories
	categories, err := repo.GetCategories(ctx)
	if err != nil {
		t.Fatalf("GetCategories() error = %v", err)
	}

	if len(categories) != 3 {
		t.Errorf("GetCategories() returned %d categories, want 3", len(categories))
	}

	// Create a new category
	newCat := &domain.Category{
		Name:  "Custom",
		Color: "purple",
	}
	if err := repo.CreateCategory(ctx, newCat); err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}

	// Should now have 4 categories
	categories, err = repo.GetCategories(ctx)
	if err != nil {
		t.Fatalf("GetCategories() error = %v", err)
	}

	if len(categories) != 4 {
		t.Errorf("GetCategories() after create returned %d categories, want 4", len(categories))
	}
}
```

**Step 22: Run test to verify it fails**

Run: `go test ./internal/repository -v -run "TestSQLiteRepository_(CreateCategory|GetCategories)"`

Expected: Compilation error - methods not defined

**Step 23: Write minimal category methods implementations**

File: `internal/repository/sqlite.go` (append)

```go
// CreateCategory creates a new category
func (r *SQLiteRepository) CreateCategory(ctx context.Context, category *domain.Category) error {
	if err := category.Validate(); err != nil {
		return err
	}

	now := time.Now()
	category.CreatedAt = now

	result, err := r.db.ExecContext(ctx,
		"INSERT INTO categories (name, color, created_at) VALUES (?, ?, ?)",
		category.Name,
		category.Color,
		category.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	category.ID = id
	return nil
}

// GetCategories retrieves all categories
func (r *SQLiteRepository) GetCategories(ctx context.Context) ([]*domain.Category, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, color, created_at FROM categories ORDER BY name",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		category := &domain.Category{}
		var createdAt string

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Color,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse timestamp
		t, _ := time.Parse(time.RFC3339, createdAt)
		category.CreatedAt = t

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
```

**Step 24: Run all repository tests**

Run: `go test ./internal/repository -v`

Expected: All tests PASS

**Step 25: Commit**

```bash
git add internal/repository/
git commit -m "feat: implement SQLite repository

Add complete SQLite repository implementation:
- CRUD operations for tasks
- Category management
- Comprehensive test coverage with in-memory database
- Proper timestamp handling and nullable fields

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 6: Basic Bubble Tea Application

**Files:**
- Create: `internal/app/app.go`
- Create: `internal/app/messages.go`
- Create: `cmd/task/main.go`

**Step 1: Add Bubble Tea dependencies**

Run: `go get github.com/charmbracelet/bubbletea github.com/charmbracelet/lipgloss`

**Step 2: Write app messages**

File: `internal/app/messages.go`

```go
package app

import "github.com/hitsumabushi845/task-management/internal/domain"

// Message types for Bubble Tea updates

type taskListLoadedMsg struct {
	tasks []*domain.Task
}

type taskCreatedMsg struct {
	task *domain.Task
}

type taskUpdatedMsg struct {
	task *domain.Task
}

type taskDeletedMsg struct {
	id int64
}

type errMsg struct {
	err error
}
```

**Step 3: Write minimal app model**

File: `internal/app/app.go`

```go
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
```

**Step 4: Write main entry point**

File: `cmd/task/main.go`

```go
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
```

**Step 5: Test build and run**

Run: `go mod tidy && go build -o bin/task ./cmd/task`

Expected: Build succeeds

**Step 6: Test run (manual)**

Run: `./bin/task`

Expected: Application starts, shows "No tasks yet", can quit with 'q'

**Step 7: Commit**

```bash
git add internal/app/ cmd/task/ go.mod go.sum
git commit -m "feat: add basic Bubble Tea application

Create minimal TUI application:
- Application model with repository integration
- Message types for async operations
- Basic view rendering
- Task loading on startup
- Graceful quit with 'q'

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 7: List View with Navigation

**Files:**
- Modify: `internal/app/app.go`
- Create: `internal/ui/styles/styles.go`

**Step 1: Create styles**

File: `internal/ui/styles/styles.go`

```go
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
```

**Step 2: Update app model with cursor navigation**

File: `internal/app/app.go` (modify Update method)

```go
// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
		}

	case taskListLoadedMsg:
		m.tasks = msg.tasks
		if m.cursor >= len(m.tasks) {
			m.cursor = len(m.tasks) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}

	case errMsg:
		m.err = msg.err

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}
```

**Step 3: Update view with styled task list**

File: `internal/app/app.go` (modify View method)

```go
import (
	"fmt"
	"github.com/hitsumabushi845/task-management/internal/domain"
	"github.com/hitsumabushi845/task-management/internal/ui/styles"
)

// View renders the application
func (m *Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n\nPress q to quit.\n"
	}

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
	helpText := "[↑/k]上 [↓/j]下 [q]終了"
	s += styles.StatusBar.Render(helpText) + "\n"

	return s
}
```

**Step 4: Test build and run**

Run: `go mod tidy && make build && ./bin/task`

Expected: Can navigate task list with j/k or arrow keys, selected item highlighted

**Step 5: Commit**

```bash
git add internal/app/app.go internal/ui/styles/ go.mod go.sum Makefile
git commit -m "feat: add task list navigation with styling

Implement cursor-based navigation:
- j/k and arrow keys for up/down movement
- Styled task display with status icons and priority colors
- Selected item highlighting
- Status bar with keyboard shortcuts

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 8: Task Creation

**Files:**
- Modify: `internal/app/app.go`
- Modify: `internal/app/messages.go`

**Step 1: Add input mode to model**

File: `internal/app/app.go` (modify Model struct)

```go
type viewMode int

const (
	viewModeList viewMode = iota
	viewModeCreate
)

type Model struct {
	repo      domain.TaskRepository
	tasks     []*domain.Task
	cursor    int
	width     int
	height    int
	err       error
	mode      viewMode
	inputTitle string
	inputPriority domain.Priority
}
```

**Step 2: Update messages**

File: `internal/app/messages.go` (add)

```go
type createTaskMsg struct{}
```

**Step 3: Add task creation command**

File: `internal/app/app.go` (add method)

```go
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
```

**Step 4: Update Update method for creation mode**

File: `internal/app/app.go` (modify Update method)

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle create mode separately
		if m.mode == viewModeCreate {
			return m.updateCreateMode(msg)
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
		}
	}

	return m, nil
}
```

**Step 5: Update View method for create mode**

File: `internal/app/app.go` (modify View method)

```go
func (m *Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n\nPress q to quit.\n"
	}

	// Create mode view
	if m.mode == viewModeCreate {
		return m.viewCreate()
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
	helpText := "[n]新規 [↑/k]上 [↓/j]下 [q]終了"
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
```

**Step 6: Test build and run**

Run: `make build && ./bin/task`

Expected:
- Press 'n' to enter create mode
- Type task title
- Press Tab to change priority
- Press Enter to create
- Task appears in list

**Step 7: Commit**

```bash
git add internal/app/
git commit -m "feat: add task creation functionality

Implement interactive task creation:
- Press 'n' to enter create mode
- Text input for task title
- Tab to cycle through priority levels
- Enter to create, Esc to cancel
- Auto-reload task list after creation

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 9: Task Deletion

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add delete command**

File: `internal/app/app.go` (add method)

```go
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
```

**Step 2: Add delete key handler**

File: `internal/app/app.go` (modify Update method, list mode case)

```go
case "d":
	// Delete selected task
	if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
		task := m.tasks[m.cursor]
		return m, m.deleteTask(task.ID)
	}
```

**Step 3: Handle taskDeletedMsg**

File: `internal/app/app.go` (modify Update method, add case)

```go
case taskDeletedMsg:
	// Task deleted, reload list
	return m, m.loadTasks()
```

**Step 4: Update status bar in list view**

File: `internal/app/app.go` (modify viewList method)

```go
// Status bar
helpText := "[n]新規 [d]削除 [↑/k]上 [↓/j]下 [q]終了"
s += styles.StatusBar.Render(helpText) + "\n"
```

**Step 5: Test build and run**

Run: `make build && ./bin/task`

Expected:
- Create some tasks
- Navigate with j/k
- Press 'd' to delete selected task
- Task is removed from list

**Step 6: Commit**

```bash
git add internal/app/app.go
git commit -m "feat: add task deletion functionality

Implement task deletion:
- Press 'd' to delete selected task
- Auto-reload list after deletion
- Update status bar help text

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 10: Status Toggle

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add status toggle command**

File: `internal/app/app.go` (add method)

```go
import "time"

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
```

**Step 2: Add space key handler**

File: `internal/app/app.go` (modify Update method, list mode case)

```go
case " ", "space":
	// Toggle task status
	if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
		task := m.tasks[m.cursor]
		return m, m.toggleTaskStatus(task)
	}
```

**Step 3: Handle taskUpdatedMsg**

File: `internal/app/app.go` (modify Update method, add case)

```go
case taskUpdatedMsg:
	// Task updated, reload list
	return m, m.loadTasks()
```

**Step 4: Update status bar**

File: `internal/app/app.go` (modify viewList method)

```go
// Status bar
helpText := "[n]新規 [d]削除 [Space]ステータス [↑/k]上 [↓/j]下 [q]終了"
s += styles.StatusBar.Render(helpText) + "\n"
```

**Step 5: Test build and run**

Run: `make build && ./bin/task`

Expected:
- Create a task (status: new, icon ○)
- Press Space - status changes to working (icon ●)
- Press Space - status changes to completed (icon ✓)
- Press Space - status cycles back to new (icon ○)

**Step 6: Commit**

```bash
git add internal/app/app.go
git commit -m "feat: add task status toggle

Implement status cycling with Space key:
- new -> working (set StartedAt)
- working -> completed (set CompletedAt)
- completed -> new (clear timestamps)
- Visual feedback with status icons

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 11: Final Integration Test

**Step 1: Run all tests**

Run: `make test`

Expected: All tests pass

**Step 2: Build application**

Run: `make build`

Expected: Binary created at `bin/task`

**Step 3: Manual integration test**

Run: `./bin/task`

Test flow:
1. Start app - should show empty list
2. Press 'n' - enter create mode
3. Type "タスク1" - text appears
4. Press Tab - priority cycles
5. Press Enter - task created, back to list
6. Repeat to create 3 tasks
7. Use j/k to navigate
8. Press Space on a task - status toggles
9. Press 'd' on a task - task deleted
10. Press 'q' - app quits cleanly

**Step 4: Check database**

Run: `sqlite3 ~/.task-management/tasks.db "SELECT * FROM tasks;"`

Expected: See tasks in database

**Step 5: Final commit**

```bash
git add -A
git commit -m "test: verify Phase 1 complete integration

Phase 1 core features complete:
✅ Project structure with domain-driven design
✅ SQLite repository with migrations
✅ Task and category domain models
✅ Bubble Tea TUI application
✅ Task list view with navigation
✅ Task creation with priority selection
✅ Task deletion
✅ Status toggle (new/working/completed)

All unit tests passing.
Manual integration tests verified.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Summary

Phase 1 implementation is complete when:

- ✅ All unit tests pass (`make test`)
- ✅ Application builds successfully (`make build`)
- ✅ Manual testing confirms:
  - Task creation works
  - Task list displays with proper styling
  - Navigation works (j/k keys)
  - Task deletion works (d key)
  - Status toggle works (Space key)
  - App quits cleanly (q key)
- ✅ Database persists tasks correctly
- ✅ All changes committed to git

**Next Steps:** Phase 2 will add UI enhancements (Kanban view, filters, help modal).
