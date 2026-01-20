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

	if len(c.Name) > 50 {
		return errors.New("name must be 50 characters or less")
	}

	if c.Color == "" {
		return errors.New("color is required")
	}

	if !isValidColor(c.Color) {
		return errors.New("invalid color: must be one of blue, green, red, yellow, purple, cyan, magenta, white, black")
	}

	return nil
}

// isValidColor checks if the color is one of the predefined colors
func isValidColor(color string) bool {
	validColors := map[string]bool{
		"blue":    true,
		"green":   true,
		"red":     true,
		"yellow":  true,
		"purple":  true,
		"cyan":    true,
		"magenta": true,
		"white":   true,
		"black":   true,
	}
	return validColors[color]
}
