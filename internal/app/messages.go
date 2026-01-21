package app

import "github.com/hitsumabushi845/task-management/internal/domain"

// Message types for Bubble Tea updates

type taskListLoadedMsg struct {
	tasks []*domain.Task
}

type taskCreatedMsg struct {
	task *domain.Task
}

type taskUpdatedMsg struct {
	task *domain.Task
}

type taskDeletedMsg struct {
	id int64
}

type errMsg struct {
	err error
}

// categoriesLoadedMsg is sent when categories are loaded
type categoriesLoadedMsg struct {
	categories []*domain.Category
}
