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
