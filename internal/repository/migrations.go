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
