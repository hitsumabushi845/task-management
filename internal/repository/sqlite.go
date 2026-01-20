package repository

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/hitsumabushi845/task-management/internal/domain"
	_ "modernc.org/sqlite"
)

// SQLiteRepository implements TaskRepository using SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	// Ensure parent directory exists (skip for in-memory database)
	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
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
		formatTimePtr(task.DueDate),
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

// Delete deletes a task by ID
func (r *SQLiteRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
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

// Helper function to format *time.Time for SQL
func formatTimePtr(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339)
}
