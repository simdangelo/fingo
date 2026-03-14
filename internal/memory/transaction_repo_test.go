package memory

import (
	"errors"
	"testing"
	"time"

	"github.com/simdangelo/fingo/internal/domain"
)

func TestTransactionRepository_Save_And_FindByID(t *testing.T) {
	repo := NewTransactionRepository()

	tx, err := domain.NewTransaction(
		time.Now(),
		45.50,
		domain.TypeExpense,
		"cat_groceries",
		"Weekly shopping",
	)

	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	err = repo.Save(*tx)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	found, err := repo.FindByID(tx.ID)
	if err != nil {
		t.Errorf("FindByID() error %v", err)
	}
	if found == nil {
		t.Fatalf("FindByID() returned nil")
	}

	if found.ID != tx.ID {
		t.Errorf("ID: want %v, got %v", tx.ID, found.ID)
	}
	if found.Amount != tx.Amount {
		t.Errorf("Amount: want %v, got %v", tx.Amount, found.Amount)
	}
	if found.Type != tx.Type {
		t.Errorf("Type: want %v, got %v", tx.Type, found.Type)
	}
	if found.CategoryID != tx.CategoryID {
		t.Errorf("CategoryID: want %v, got %v", tx.CategoryID, found.CategoryID)
	}
	if found.Description != tx.Description {
		t.Errorf("CategoryID: want %v, got %v", tx.Description, found.Description)
	}
}

func TestTransactionRepository_FindByDateRange(t *testing.T) {
	repo := NewTransactionRepository()

	jan1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	jan15 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	jan31 := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	feb1 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	tx1, _ := domain.NewTransaction(jan1, 100.00, domain.TypeExpense, "cat1", "Jan 1")
	tx2, _ := domain.NewTransaction(jan15, 200.00, domain.TypeExpense, "cat1", "Jan 15")
	tx3, _ := domain.NewTransaction(jan31, 300.00, domain.TypeExpense, "cat1", "Jan 31")
	tx4, _ := domain.NewTransaction(feb1, 400.00, domain.TypeExpense, "cat1", "Feb 1")

	repo.Save(*tx1)
	repo.Save(*tx2)
	repo.Save(*tx3)
	repo.Save(*tx4)

	results, err := repo.FindByDateRange(jan1, jan31)
	if err != nil {
		t.Fatalf("FindByDateRange() fatal error %v", err)
	}

	// Should return 3 transactions (Jan 1, Jan 15, Jan 31)
	if len(results) < 3 {
		t.Errorf("FindByDateRange() returned %d transactions, want 3", len(results))
	}

	foundIDs := make(map[string]bool)
	for _, tx := range results {
		foundIDs[tx.ID] = true
	}

	if !foundIDs[tx1.ID] || !foundIDs[tx2.ID] || !foundIDs[tx3.ID] {
		t.Errorf("FindByDateRange() missing expected transactions")
	}
	if foundIDs[tx4.ID] {
		t.Errorf("FindByDateRange() should not include Feb 1 transaction")
	}
}

func TestTransactionRepository_FindByDateRange_Empty(t *testing.T) {
	repo := NewTransactionRepository()

	jan1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	jan31 := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	results, err := repo.FindByDateRange(jan1, jan31)
	if err != nil {
		t.Fatalf("FindByDateRange() fatal error %v", err)
	}

	if results == nil {
		t.Errorf("FindByDateRange() should return empty slice, not nil %v", results)
	}

	if len(results) != 0 {
		t.Errorf("FindByDateRange() returned %d transactions, want 0", len(results))
	}
}

func TestTransactionRepository_FindByCategory(t *testing.T) {
	repo := NewTransactionRepository()
	tx1, _ := domain.NewTransaction(time.Now(), 50.00, domain.TypeExpense, "cat_groceries", "Groceries 1")
	tx2, _ := domain.NewTransaction(time.Now(), 60.00, domain.TypeExpense, "cat_groceries", "Groceries 2")
	tx3, _ := domain.NewTransaction(time.Now(), 1200.00, domain.TypeExpense, "cat_rent", "Rent")

	repo.Save(*tx1)
	repo.Save(*tx2)
	repo.Save(*tx3)

	results, err := repo.FindByCategory("cat_groceries")
	if err != nil {
		t.Fatalf("FindByCategory() error %v", err)
	}

	if len(results) != 2 {
		t.Errorf("FindByCategory() returned %d transactions, want 2", len(results))
	}

	for _, tx := range results {
		if tx.CategoryID != "cat_groceries" {
			t.Errorf("FindByCategory() returned transaction with category %v", tx.CategoryID)
		}
	}
}

func TestTransactionRepository_FindAll_SortedByDate(t *testing.T) {
	repo := NewTransactionRepository()

	jan1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	jan15 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	jan31 := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	tx1, _ := domain.NewTransaction(jan1, 100.00, domain.TypeExpense, "cat1", "first")
	tx2, _ := domain.NewTransaction(jan15, 200.00, domain.TypeExpense, "cat1", "middle")
	tx3, _ := domain.NewTransaction(jan31, 300.00, domain.TypeExpense, "cat1", "last")

	repo.Save(*tx2)
	repo.Save(*tx3)
	repo.Save(*tx1)

	results, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() error %v", err)
	}

	if len(results) != 3 {
		t.Errorf("FindAll() returned %d transactions, want 3", len(results))
	}

	// Verify order: Jan 31, Jan 15, Jan 1
	if results[0].ID != tx3.ID {
		t.Errorf("results[0] = %s, want %s (newest)", results[0].ID, tx3.ID)
	}
	if results[1].ID != tx2.ID {
		t.Errorf("results[1] = %s, want %s (newest)", results[1].ID, tx2.ID)
	}
	if results[2].ID != tx1.ID {
		t.Errorf("results[2] = %s, want %s (newest)", results[2].ID, tx1.ID)
	}
}

func TestTransactionRepository_Delete(t *testing.T) {
	repo := NewTransactionRepository()

	tx, _ := domain.NewTransaction(time.Now(), 50.00, domain.TypeExpense, "cat1", "tx_to_delete")
	repo.Save(*tx)

	err := repo.Delete(tx.ID)
	if err != nil {
		t.Errorf("Delete() error: %v", err)
	}

	found, err := repo.FindByID(tx.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("FindByID() after delete should return ErrNotFound error, but got %v", err)
	}

	if found != nil {
		t.Errorf("FindByID() should return nil after delete")
	}
}

func TestTransactionRepository_Delete_ErrorNotFound(t *testing.T) {
	repo := NewTransactionRepository()

	err := repo.Delete("some-non-existent-id")
	if err == nil {
		t.Errorf("Expected an error when deleting non-existent ID, but got nil")
	}

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Expected domain.ErrNotFound, but got: %v", err)
	}
}
