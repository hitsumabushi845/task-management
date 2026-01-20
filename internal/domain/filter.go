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
