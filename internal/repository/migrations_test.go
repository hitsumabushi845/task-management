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
