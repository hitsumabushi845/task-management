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
				Color: "#FF0000",
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
