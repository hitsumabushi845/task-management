package domain

import (
	"errors"
	"strings"
	"time"
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	TaskStatusNew       TaskStatus = "new"
	TaskStatusWorking   TaskStatus = "working"
	TaskStatusCompleted TaskStatus = "completed"
)

// String returns the string representation of the task status
func (s TaskStatus) String() string {
	return string(s)
}

// IsValid checks if the task status is valid
func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusNew, TaskStatusWorking, TaskStatusCompleted:
		return true
	default:
		return false
	}
}

// Priority represents the importance level of a task
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// String returns the string representation of the priority
func (p Priority) String() string {
	return string(p)
}

// IsValid checks if the priority is valid
func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}

// Task represents a task in the task management system
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

// Validate checks if the task has valid data
func (t *Task) Validate() error {
	if strings.TrimSpace(t.Title) == "" {
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
