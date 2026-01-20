package domain

import (
	"strings"
	"testing"
)

func TestTaskStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   string
	}{
		{
			name:   "new status",
			status: TaskStatusNew,
			want:   "new",
		},
		{
			name:   "working status",
			status: TaskStatusWorking,
			want:   "working",
		},
		{
			name:   "completed status",
			status: TaskStatusCompleted,
			want:   "completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("TaskStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   bool
	}{
		{
			name:   "new is valid",
			status: TaskStatusNew,
			want:   true,
		},
		{
			name:   "working is valid",
			status: TaskStatusWorking,
			want:   true,
		},
		{
			name:   "completed is valid",
			status: TaskStatusCompleted,
			want:   true,
		},
		{
			name:   "invalid status",
			status: TaskStatus("invalid"),
			want:   false,
		},
		{
			name:   "empty status",
			status: TaskStatus(""),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("TaskStatus.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriority_String(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		want     string
	}{
		{
			name:     "low priority",
			priority: PriorityLow,
			want:     "low",
		},
		{
			name:     "medium priority",
			priority: PriorityMedium,
			want:     "medium",
		},
		{
			name:     "high priority",
			priority: PriorityHigh,
			want:     "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priority.String(); got != tt.want {
				t.Errorf("Priority.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriority_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		want     bool
	}{
		{
			name:     "low is valid",
			priority: PriorityLow,
			want:     true,
		},
		{
			name:     "medium is valid",
			priority: PriorityMedium,
			want:     true,
		},
		{
			name:     "high is valid",
			priority: PriorityHigh,
			want:     true,
		},
		{
			name:     "invalid priority",
			priority: Priority("invalid"),
			want:     false,
		},
		{
			name:     "empty priority",
			priority: Priority(""),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priority.IsValid(); got != tt.want {
				t.Errorf("Priority.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    *Task
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid task",
			task: &Task{
				Title:    "Test Task",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			wantErr: false,
		},
		{
			name: "empty title",
			task: &Task{
				Title:    "",
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "title too long",
			task: &Task{
				Title:    strings.Repeat("a", 201),
				Status:   TaskStatusNew,
				Priority: PriorityMedium,
			},
			wantErr: true,
			errMsg:  "title must be 200 characters or less",
		},
		{
			name: "invalid status",
			task: &Task{
				Title:    "Test Task",
				Status:   TaskStatus("invalid"),
				Priority: PriorityMedium,
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name: "invalid priority",
			task: &Task{
				Title:    "Test Task",
				Status:   TaskStatusNew,
				Priority: Priority("invalid"),
			},
			wantErr: true,
			errMsg:  "invalid priority",
		},
		{
			name: "description too long",
			task: &Task{
				Title:       "Test Task",
				Description: strings.Repeat("a", 1001),
				Status:      TaskStatusNew,
				Priority:    PriorityMedium,
			},
			wantErr: true,
			errMsg:  "description must be 1000 characters or less",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Task.Validate() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Task.Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Task.Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
