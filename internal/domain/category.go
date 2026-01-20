package domain

import (
	"errors"
	"time"
)

// Category represents a category for organizing tasks
type Category struct {
	ID        int64
	Name      string
	Color     string
	CreatedAt time.Time
}

// Validate checks if the category has valid data
func (c *Category) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}

	if c.Color == "" {
		return errors.New("color is required")
	}

	return nil
}
