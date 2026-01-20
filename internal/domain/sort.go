package domain

import "sort"

// SortBy represents the field to sort by
type SortBy int

const (
	SortByCreatedAt SortBy = iota
	SortByDueDate
	SortByPriority
	SortByStatus
	SortByTitle
)

func (s SortBy) String() string {
	switch s {
	case SortByCreatedAt:
		return "created_at"
	case SortByDueDate:
		return "due_date"
	case SortByPriority:
		return "priority"
	case SortByStatus:
		return "status"
	case SortByTitle:
		return "title"
	default:
		return "unknown"
	}
}

func (s SortBy) IsValid() bool {
	return s >= SortByCreatedAt && s <= SortByTitle
}

// Sort represents sorting criteria
type Sort struct {
	By        SortBy
	Ascending bool
}

// Apply sorts a slice of tasks
func (s *Sort) Apply(tasks []*Task) []*Task {
	result := make([]*Task, len(tasks))
	copy(result, tasks)

	sort.Slice(result, func(i, j int) bool {
		var less bool

		switch s.By {
		case SortByCreatedAt:
			less = result[i].CreatedAt.Before(result[j].CreatedAt)
		case SortByDueDate:
			// nil due dates go last
			if result[i].DueDate == nil && result[j].DueDate == nil {
				less = false
			} else if result[i].DueDate == nil {
				less = false
			} else if result[j].DueDate == nil {
				less = true
			} else {
				less = result[i].DueDate.Before(*result[j].DueDate)
			}
		case SortByPriority:
			// High > Medium > Low
			pi := priorityOrder(result[i].Priority)
			pj := priorityOrder(result[j].Priority)
			less = pi > pj
		case SortByStatus:
			si := statusOrder(result[i].Status)
			sj := statusOrder(result[j].Status)
			less = si < sj
		case SortByTitle:
			less = result[i].Title < result[j].Title
		}

		if s.Ascending {
			return less
		}
		return !less
	})

	return result
}

func priorityOrder(p Priority) int {
	switch p {
	case PriorityHigh:
		return 0
	case PriorityMedium:
		return 1
	case PriorityLow:
		return 2
	default:
		return 3
	}
}

func statusOrder(s TaskStatus) int {
	switch s {
	case TaskStatusNew:
		return 0
	case TaskStatusWorking:
		return 1
	case TaskStatusCompleted:
		return 2
	default:
		return 3
	}
}
