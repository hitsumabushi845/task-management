package domain

import (
	"testing"
	"time"
)

func TestFilter_Match(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	tomorrow := now.Add(24 * time.Hour)
	yesterday := now.Add(-24 * time.Hour)

	tests := []struct {
		name   string
		filter Filter
		task   Task
		want   bool
	}{
		{
			name:   "empty filter matches all",
			filter: Filter{},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: true,
		},
		{
			name: "status filter matches",
			filter: Filter{
				Statuses: []TaskStatus{TaskStatusNew},
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: true,
		},
		{
			name: "status filter excludes",
			filter: Filter{
				Statuses: []TaskStatus{TaskStatusWorking},
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: false,
		},
		{
			name: "priority filter matches",
			filter: Filter{
				Priorities: []Priority{PriorityHigh, PriorityMedium},
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: true,
		},
		{
			name: "priority filter excludes",
			filter: Filter{
				Priorities: []Priority{PriorityHigh},
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityLow,
			},
			want: false,
		},
		{
			name: "search text matches title",
			filter: Filter{
				SearchText: "test",
			},
			task: Task{
				Title:    "Test Task",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: true,
		},
		{
			name: "search text matches description",
			filter: Filter{
				SearchText: "important",
			},
			task: Task{
				Title:       "Task",
				Description: "This is important",
				Status:      TaskStatusNew,
				Priority:    PriorityMedium,
			},
			want: true,
		},
		{
			name: "search text no match",
			filter: Filter{
				SearchText: "xyz",
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: false,
		},
		{
			name: "overdue filter matches",
			filter: Filter{
				DateRange: DateRangeOverdue,
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
				DueDate:  &yesterday,
			},
			want: true,
		},
		{
			name: "overdue filter excludes future",
			filter: Filter{
				DateRange: DateRangeOverdue,
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
				DueDate:  &tomorrow,
			},
			want: false,
		},
		{
			name: "category filter matches",
			filter: Filter{
				Categories: []int64{1, 2},
			},
			task: Task{
				Title:      "Test",
				Status:     TaskStatusNew,
				Priority:   PriorityMedium,
				CategoryID: func() *int64 { v := int64(1); return &v }(),
			},
			want: true,
		},
		{
			name: "category filter excludes nil category",
			filter: Filter{
				Categories: []int64{1},
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: false,
		},
		{
			name: "today filter matches",
			filter: Filter{
				DateRange: DateRangeToday,
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
				DueDate:  &today,
			},
			want: true,
		},
		{
			name: "this week filter matches",
			filter: Filter{
				DateRange: DateRangeThisWeek,
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
				DueDate:  &tomorrow,
			},
			want: true,
		},
		{
			name: "no due date filter matches",
			filter: Filter{
				DateRange: DateRangeNoDueDate,
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			want: true,
		},
		{
			name: "no due date filter excludes tasks with due date",
			filter: Filter{
				DateRange: DateRangeNoDueDate,
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
				DueDate:  &tomorrow,
			},
			want: false,
		},
		{
			name: "combined filter matches",
			filter: Filter{
				Statuses:   []TaskStatus{TaskStatusNew},
				Priorities: []Priority{PriorityHigh},
				SearchText: "urgent",
			},
			task: Task{
				Title:    "Urgent Task",
				Status:   TaskStatusNew,
				Priority: PriorityHigh,
			},
			want: true,
		},
		{
			name: "combined filter partial match fails",
			filter: Filter{
				Statuses:   []TaskStatus{TaskStatusNew},
				Priorities: []Priority{PriorityHigh},
			},
			task: Task{
				Title:    "Test",
				Status:   TaskStatusNew,
				Priority: PriorityLow,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Match(&tt.task)
			if got != tt.want {
				t.Errorf("Filter.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_IsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		filter Filter
		want   bool
	}{
		{
			name:   "empty filter",
			filter: Filter{},
			want:   true,
		},
		{
			name: "filter with status",
			filter: Filter{
				Statuses: []TaskStatus{TaskStatusNew},
			},
			want: false,
		},
		{
			name: "filter with search text",
			filter: Filter{
				SearchText: "test",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.IsEmpty()
			if got != tt.want {
				t.Errorf("Filter.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_Apply(t *testing.T) {
	tasks := []*Task{
		{ID: 1, Title: "High Priority", Status: TaskStatusNew, Priority: PriorityHigh},
		{ID: 2, Title: "Medium Priority", Status: TaskStatusWorking, Priority: PriorityMedium},
		{ID: 3, Title: "Low Priority", Status: TaskStatusCompleted, Priority: PriorityLow},
	}

	tests := []struct {
		name    string
		filter  Filter
		wantIDs []int64
	}{
		{
			name:    "empty filter returns all",
			filter:  Filter{},
			wantIDs: []int64{1, 2, 3},
		},
		{
			name: "status filter",
			filter: Filter{
				Statuses: []TaskStatus{TaskStatusNew, TaskStatusWorking},
			},
			wantIDs: []int64{1, 2},
		},
		{
			name: "priority filter",
			filter: Filter{
				Priorities: []Priority{PriorityHigh},
			},
			wantIDs: []int64{1},
		},
		{
			name: "no matches returns empty",
			filter: Filter{
				SearchText: "nonexistent",
			},
			wantIDs: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Apply(tasks)
			if len(got) != len(tt.wantIDs) {
				t.Errorf("Filter.Apply() returned %d tasks, want %d", len(got), len(tt.wantIDs))
				return
			}
			for i, task := range got {
				if task.ID != tt.wantIDs[i] {
					t.Errorf("Filter.Apply()[%d].ID = %d, want %d", i, task.ID, tt.wantIDs[i])
				}
			}
		})
	}
}
