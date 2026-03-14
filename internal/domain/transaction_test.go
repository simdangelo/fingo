package domain

import (
	"strings"
	"testing"
	"time"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		name string
		txType TransactionType
		want bool
	}{
		{
			name: "expense is valid",
			txType: TypeExpense,
			want: true,
		},
		{
			name: "income is valid",
			txType: TypeIncome,
			want: true,
		},
		{
			name: "empty string is invalid",
			txType: "",
			want: false,
		},
		{
			name: "random string is invalid",
			txType: "savings",
			want: false,
		},
	}

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := IsValid(tt.txType)
            if got != tt.want {
                t.Errorf("TransactionType.IsValid(%v) = %v, want %v", tt.txType, got, tt.want)
            }
        })
    }
}

func TestNewTransaction(t *testing.T) {
	validDate := time.Now().AddDate(0, 0, -1) // Yesterday
    validCategoryID := "cat_groceries"

	tests := []struct {
		name string
		date time.Time
		amount float64
		txType TransactionType
		categoryID string
		description string
		wantErr bool
		errContains string
	}{
		{
            name:        "valid expense",
            date:        validDate,
            amount:      45.50,
            txType:      TypeExpense,
            categoryID:  validCategoryID,
            description: "Weekly groceries",
            wantErr:     false,
        },
        {
            name:        "valid income",
            date:        validDate,
            amount:      3500.00,
            txType:      TypeIncome,
            categoryID:  "cat_salary",
            description: "January salary",
            wantErr:     false,
        },
        {
            name:        "valid with empty description",
            date:        validDate,
            amount:      10.00,
            txType:      TypeExpense,
            categoryID:  validCategoryID,
            description: "",
            wantErr:     false,
        },
        {
            name:        "valid with small amount",
            date:        validDate,
            amount:      0.01,
            txType:      TypeExpense,
            categoryID:  validCategoryID,
            description: "One cent",
            wantErr:     false,
        },
        {
            name:        "valid with today's date",
            date:        time.Now(),
            amount:      100.00,
            txType:      TypeExpense,
            categoryID:  validCategoryID,
            description: "Today's transaction",
            wantErr:     false,
        },

        // Invalid amount cases
        {
            name:        "zero amount",
            date:        validDate,
            amount:      0,
            txType:      TypeExpense,
            categoryID:  validCategoryID,
            description: "Zero",
            wantErr:     true,
            errContains: "greater than zero",
        },
        {
            name:        "negative amount",
            date:        validDate,
            amount:      -50.00,
            txType:      TypeExpense,
            categoryID:  validCategoryID,
            description: "Negative",
            wantErr:     true,
            errContains: "greater than zero",
        },

        // Invalid type cases
        {
            name:        "invalid transaction type",
            date:        validDate,
            amount:      50.00,
            txType:      "invalid",
            categoryID:  validCategoryID,
            description: "Bad type",
            wantErr:     true,
            errContains: "invalid transaction type",
        },
        {
            name:        "empty transaction type",
            date:        validDate,
            amount:      50.00,
            txType:      "",
            categoryID:  validCategoryID,
            description: "Empty type",
            wantErr:     true,
            errContains: "invalid transaction type",
        },

        // Invalid category cases
        {
            name:        "empty category ID",
            date:        validDate,
            amount:      50.00,
            txType:      TypeExpense,
            categoryID:  "",
            description: "No category",
            wantErr:     true,
            errContains: "category ID cannot be empty",
        },
        {
            name:        "whitespace only category ID",
            date:        validDate,
            amount:      50.00,
            txType:      TypeExpense,
            categoryID:  "   ",
            description: "Spaces only",
            wantErr:     true,
            errContains: "category ID cannot be empty",
        },

        // Invalid description cases
        {
            name:        "description too long",
            date:        validDate,
            amount:      50.00,
            txType:      TypeExpense,
            categoryID:  validCategoryID,
            description: strings.Repeat("a", 201), // 201 characters
            wantErr:     true,
            errContains: "description too long",
        },

        // Invalid date cases
        {
            name:        "date far in future",
            date:        time.Now().AddDate(1, 0, 0), // 1 year from now
            amount:      50.00,
            txType:      TypeExpense,
            categoryID:  validCategoryID,
            description: "Future date",
            wantErr:     true,
            errContains: "future",
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			got, err := NewTransaction(
				tt.date,
				tt.amount,
				tt.txType,
				tt.categoryID,
				tt.description,
			)

            // Valid cases
            if !tt.wantErr {
                if got.Date != tt.date {
                    t.Errorf("Date: want %v, got %v", tt.date, got.Date)
                }
                if got.Amount != tt.amount {
                    t.Errorf("Amount: want %v, got %v", tt.amount, got.Amount)
                }
                if got.Type != tt.txType {
                    t.Errorf("Type: want %v, got %v", tt.txType, got.Type)
                }
                if got.CategoryID != tt.categoryID {
                    t.Errorf("CategoryID: want %v, got %v", tt.categoryID, got.CategoryID)
                }
                if got.Description != tt.description {
                    t.Errorf("Error description: want %v, got %v", tt.description, got.Description)
                }
            }
            
            // If we expected an error, verify it contains the right message
            if tt.wantErr == true && err != nil {
                if !strings.Contains(err.Error(), tt.errContains) {
                    t.Errorf("NewTransaction() error = %v, should contain %q", err, tt.errContains)
                }
            }

            // unexpected error
            if tt.wantErr == false && err != nil {
                t.Errorf("No errors expected, got %v", err)
            }

            // expected an error, but got none
            if tt.wantErr == true && err == nil {
                t.Errorf("expected an error (%v), but got none", tt.wantErr)
            }

		})
	}
}

func TestGenerateTransactionID(t *testing.T) {
    got, err := NewTransaction(
        time.Now(),
        50.00,
        TypeExpense,
        "cat_test",
        "Test transaction",
    )

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if got.ID == "" {
        t.Errorf("expected ID to be generated, got empty string")
    }

    if !strings.HasPrefix(got.ID, "tx_") {
        t.Errorf("expected ID to start with 'tx_', got %q", got.ID)
    }
}

func TestNewTransaction_GeneratesUniqueIDs(t *testing.T) {
    got1, err1 := NewTransaction(
        time.Now(),
        50.00,
        TypeExpense,
        "cat_test",
        "transaction 1",
    )

    got2, err2 := NewTransaction(
        time.Now(),
        50.00,
        TypeExpense,
        "cat_test",
        "transaction 2",
    )

    if err1 != nil {
        t.Fatalf("unexpected error: %v", err1)
    }
    if err2 != nil {
        t.Fatalf("unexpected error: %v", err2)
    }

    if got1.ID == got2.ID {
        t.Errorf("expected unique ids, both got %q", got1.ID)
    }
}

func TestNewTransaction_SetsCreatedAt(t *testing.T) {
    before := time.Now()

    got, err := NewTransaction(
        time.Now().AddDate(0, 0, -1),
        50.00,
        TypeExpense,
        "cat_test",
        "Test transaction",
    )
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    after := time.Now()

    if got.CreatedAt.Before(before) || got.CreatedAt.After(after) {
        t.Errorf("CreatedAt: %v, should be between %v and %v", got.CreatedAt, before, after)
    }

    if got.CreatedAt.Equal(got.Date) {
        t.Errorf("CreatedAt should be current time, not transaction date")
    }

}