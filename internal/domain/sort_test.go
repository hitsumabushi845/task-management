package domain

import (
	"testing"
	"time"
)

func TestSortTasks(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	tasks := []*Task{
		{ID: 1, Title: "B Task", Priority: PriorityMedium, Status: TaskStatusWorking, CreatedAt: now, DueDate: &tomorrow},
		{ID: 2, Title: "A Task", Priority: PriorityHigh, Status: TaskStatusNew, CreatedAt: yesterday, DueDate: &now},
		{ID: 3, Title: "C Task", Priority: PriorityLow, Status: TaskStatusCompleted, CreatedAt: tomorrow, DueDate: &yesterday},
	}

	tests := []struct {
		name      string
		sortBy    SortBy
		ascending bool
		wantOrder []int64
	}{
		{
			name:      "sort by created desc",
			sortBy:    SortByCreatedAt,
			ascending: false,
			wantOrder: []int64{3, 1, 2},
		},
		{
			name:      "sort by created asc",
			sortBy:    SortByCreatedAt,
			ascending: true,
			wantOrder: []int64{2, 1, 3},
		},
		{
			name:      "sort by priority desc (high first)",
			sortBy:    SortByPriority,
			ascending: false,
			wantOrder: []int64{2, 1, 3},
		},
		{
			name:      "sort by due date asc",
			sortBy:    SortByDueDate,
			ascending: true,
			wantOrder: []int64{3, 2, 1},
		},
		{
			name:      "sort by status asc",
			sortBy:    SortByStatus,
			ascending: true,
			wantOrder: []int64{2, 1, 3}, // New, Working, Completed
		},
		{
			name:      "sort by status desc",
			sortBy:    SortByStatus,
			ascending: false,
			wantOrder: []int64{3, 1, 2}, // Completed, Working, New
		},
		{
			name:      "sort by title asc",
			sortBy:    SortByTitle,
			ascending: true,
			wantOrder: []int64{2, 1, 3}, // A Task, B Task, C Task
		},
		{
			name:      "sort by title desc",
			sortBy:    SortByTitle,
			ascending: false,
			wantOrder: []int64{3, 1, 2}, // C Task, B Task, A Task
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid mutating original
			tasksCopy := make([]*Task, len(tasks))
			copy(tasksCopy, tasks)

			// Record original order
			originalOrder := make([]int64, len(tasksCopy))
			for i, task := range tasksCopy {
				originalOrder[i] = task.ID
			}

			sort := Sort{By: tt.sortBy, Ascending: tt.ascending}
			result := sort.Apply(tasksCopy)

			// Verify sorted result
			for i, wantID := range tt.wantOrder {
				if result[i].ID != wantID {
					t.Errorf("position %d: got ID %d, want %d", i, result[i].ID, wantID)
				}
			}

			// Verify original slice is unchanged
			for i, wantID := range originalOrder {
				if tasksCopy[i].ID != wantID {
					t.Errorf("original slice modified at position %d: got ID %d, want %d", i, tasksCopy[i].ID, wantID)
				}
			}
		})
	}
}

func TestSortNilDueDateGoesLast(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	tasks := []*Task{
		{ID: 1, Title: "No Due Date", DueDate: nil},
		{ID: 2, Title: "Has Due Date", DueDate: &now},
		{ID: 3, Title: "Also No Due Date", DueDate: nil},
		{ID: 4, Title: "Earlier Due Date", DueDate: &yesterday},
	}

	// Record original order
	originalOrder := make([]int64, len(tasks))
	for i, task := range tasks {
		originalOrder[i] = task.ID
	}

	sort := Sort{By: SortByDueDate, Ascending: true}
	result := sort.Apply(tasks)

	// Tasks with due dates should come first, sorted by date
	// Then tasks with nil due dates should come last
	wantOrder := []int64{4, 2, 1, 3} // yesterday, now, nil, nil

	for i, wantID := range wantOrder {
		if result[i].ID != wantID {
			t.Errorf("position %d: got ID %d, want %d", i, result[i].ID, wantID)
		}
	}

	// Verify original slice is unchanged
	for i, wantID := range originalOrder {
		if tasks[i].ID != wantID {
			t.Errorf("original slice modified at position %d: got ID %d, want %d", i, tasks[i].ID, wantID)
		}
	}
}

func TestSortByString(t *testing.T) {
	tests := []struct {
		sortBy SortBy
		want   string
	}{
		{SortByCreatedAt, "created_at"},
		{SortByDueDate, "due_date"},
		{SortByPriority, "priority"},
		{SortByStatus, "status"},
		{SortByTitle, "title"},
		{SortBy(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.sortBy.String(); got != tt.want {
				t.Errorf("SortBy.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortByIsValid(t *testing.T) {
	tests := []struct {
		name   string
		sortBy SortBy
		want   bool
	}{
		{"SortByCreatedAt", SortByCreatedAt, true},
		{"SortByDueDate", SortByDueDate, true},
		{"SortByPriority", SortByPriority, true},
		{"SortByStatus", SortByStatus, true},
		{"SortByTitle", SortByTitle, true},
		{"invalid negative", SortBy(-1), false},
		{"invalid high", SortBy(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sortBy.IsValid(); got != tt.want {
				t.Errorf("SortBy.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
