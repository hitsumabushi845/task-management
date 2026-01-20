package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hitsumabushi845/task-management/internal/domain"
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

	// Verify both tasks are present (order may vary due to same timestamp)
	titles := map[string]bool{}
	for _, task := range tasks {
		titles[task.Title] = true
	}
	if !titles["Task 1"] {
		t.Errorf("List() missing Task 1")
	}
	if !titles["Task 2"] {
		t.Errorf("List() missing Task 2")
	}
}

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
