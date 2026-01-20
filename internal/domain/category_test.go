package domain

import (
	"strings"
	"testing"
)

func TestCategory_Validate(t *testing.T) {
	tests := []struct {
		name     string
		category *Category
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid category",
			category: &Category{
				Name:  "Work",
				Color: "red",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			category: &Category{
				Name:  "",
				Color: "#FF0000",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "empty color",
			category: &Category{
				Name:  "Work",
				Color: "",
			},
			wantErr: true,
			errMsg:  "color is required",
		},
		{
			name: "name too long",
			category: &Category{
				Name:  strings.Repeat("a", 51),
				Color: "blue",
			},
			wantErr: true,
			errMsg:  "name must be 50 characters or less",
		},
		{
			name: "name exactly 50 characters",
			category: &Category{
				Name:  strings.Repeat("a", 50),
				Color: "blue",
			},
			wantErr: false,
		},
		{
			name: "invalid color format",
			category: &Category{
				Name:  "Work",
				Color: "#FF0000",
			},
			wantErr: true,
			errMsg:  "invalid color",
		},
		{
			name: "valid predefined color blue",
			category: &Category{
				Name:  "Work",
				Color: "blue",
			},
			wantErr: false,
		},
		{
			name: "valid predefined color green",
			category: &Category{
				Name:  "Work",
				Color: "green",
			},
			wantErr: false,
		},
		{
			name: "valid predefined color red",
			category: &Category{
				Name:  "Work",
				Color: "red",
			},
			wantErr: false,
		},
		{
			name: "valid predefined color yellow",
			category: &Category{
				Name:  "Work",
				Color: "yellow",
			},
			wantErr: false,
		},
		{
			name: "valid predefined color purple",
			category: &Category{
				Name:  "Work",
				Color: "purple",
			},
			wantErr: false,
		},
		{
			name: "valid predefined color cyan",
			category: &Category{
				Name:  "Work",
				Color: "cyan",
			},
			wantErr: false,
		},
		{
			name: "valid predefined color magenta",
			category: &Category{
				Name:  "Work",
				Color: "magenta",
			},
			wantErr: false,
		},
		{
			name: "valid predefined color white",
			category: &Category{
				Name:  "Work",
				Color: "white",
			},
			wantErr: false,
		},
		{
			name: "valid predefined color black",
			category: &Category{
				Name:  "Work",
				Color: "black",
			},
			wantErr: false,
		},
		{
			name: "invalid custom color",
			category: &Category{
				Name:  "Work",
				Color: "orange",
			},
			wantErr: true,
			errMsg:  "invalid color",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.category.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Category.Validate() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Category.Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Category.Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
