package domain

import (
	"strings"
	"testing"
)

func TestNewCategory(t *testing.T) {
	tests := []struct {
        name      string  // test name (describes what you're testing)
        inputName string  // category name to test
        color     string  // color to test
        wantErr   bool    // should this fail?
    }{	
        // Valid cases
        {
            name:      "valid category",
            inputName: "Groceries ",
            color:     "#FF5733",
            wantErr:   false,
        },
		{
			name: "name with spaces trimmed",
			inputName: "    Rent",
			color: "#00FF00",
			wantErr: false,
		},
		// Invalid cases
		{
            name:      "name with only spaces",
            inputName: "   ",
            color:     "#FF5733",
            wantErr:   true,
        },
        {
            name:      "name too long",
            inputName: "This is a very long category name that exceeds fifty characters",
            color:     "#FF5733",
            wantErr:   true,
        },
        {
            name:      "invalid color format",
            inputName: "Groceries",
            color:     "not-a-color",
            wantErr:   true,
        },
        {
            name:      "color missing hash",
            inputName: "Groceries",
            color:     "FF5733",
            wantErr:   true,
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCategory(tt.inputName, tt.color)

            // Check if error matches expectation
            if (err != nil) != tt.wantErr {
                t.Errorf("NewCategory() error = %v (wantErr: %v)", err, tt.wantErr)
                return
            }

            // If we expected success, verify the result
            if !tt.wantErr {
                if got.Name != strings.TrimSpace(tt.inputName) {
                    t.Errorf("Name = %v, want %v", got.Name, strings.TrimSpace(tt.inputName))
                }
                if got.Color != tt.color {
                    t.Errorf("Color = %v, want %v", got.Color, tt.color)
                }
            }
		})
	}
}

func TestNewCategory_GenerateUniqueIDs(t *testing.T) {
    cat1, err1 := NewCategory("Category1", "#FF5733")
    cat2, err2 := NewCategory("Category2", "#00FF00")

    if err1 != nil {
        t.Fatal(err1)
    }
    if err2 != nil {
        t.Fatal(err2)
    }
    
    if cat1.ID == cat2.ID {
        t.Errorf("expected unique IDs, got duplicates, %v vs %v", cat1.ID, cat2.ID)
    }
}

func TestIsValidHexColor(t * testing.T) {
    tests := []struct {
        name string
        color string
        want bool
    }{
        {"valid color", "#FF5733", true},
        {"valid lowercase", "#ff5733", true},
        {"missing hash", "FF5733", false},
        {"too short", "#FF573", false},
        {"too long", "#FF57333", false},
        {"invalid character", "#GG5733", false},
        {"empty", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := IsValidHexColor(tt.color)
            if got != tt.want {
                t.Errorf("isValidHexColor(%q) = %v, want %v", tt.color, got, tt.want)
            }
        })
    }
}