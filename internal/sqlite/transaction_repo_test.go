package sqlite

import (
	"testing"
	"time"

	"github.com/simdangelo/fingo/internal/domain"
)

func setupTransactionRepoTest(t *testing.T) (*TransactionRepository, *CategoryRepository, func()) {
    t.Helper()
    
    db, err := OpenInMemory()
    if err != nil {
        t.Fatalf("failed to open in-memory db: %v", err)
    }
    
    if err := Migrate(db); err != nil {
        db.Close()
        t.Fatalf("failed to migrate: %v", err)
    }
    
    txRepo := NewTransactionRepository(db)
    catRepo := NewCategoryRepository(db)
    
    cleanup := func() {
        db.Close()
    }
    
    return txRepo, catRepo, cleanup
}

func TestTransactionRepository_Save_And_FindByID(t *testing.T) {
    txRepo, catRepo, cleanup := setupTransactionRepoTest(t)
    defer cleanup()
    
    // Create category first (foreign key requirement)
    cat, _ := domain.NewCategory("Groceries", "#FF5733")
    catRepo.Save(*cat)
    
    // Create transaction
    tx, err := domain.NewTransaction(
        time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
        45.50,
        domain.TypeExpense,
        cat.ID,
        "Weekly shopping",
    )
    if err != nil {
        t.Fatalf("failed to create transaction: %v", err)
    }
    
    // Save it
    if err := txRepo.Save(*tx); err != nil {
        t.Fatalf("Save() error = %v", err)
    }
    
    // Retrieve it
    found, err := txRepo.FindByID(tx.ID)
    if err != nil {
        t.Fatalf("FindByID() error = %v", err)
    }
    
    // Verify fields
    if found.ID != tx.ID {
        t.Errorf("ID = %v, want %v", found.ID, tx.ID)
    }
    if found.Amount != tx.Amount {
        t.Errorf("Amount = %v, want %v", found.Amount, tx.Amount)
    }
    if found.Type != tx.Type {
        t.Errorf("Type = %v, want %v", found.Type, tx.Type)
    }
    if found.CategoryID != tx.CategoryID {
        t.Errorf("CategoryID = %v, want %v", found.CategoryID, tx.CategoryID)
    }
    if found.Description != tx.Description {
        t.Errorf("Description = %v, want %v", found.Description, tx.Description)
    }
    
    // Verify dates (truncate to seconds since RFC3339 doesn't preserve nanoseconds)
    if !found.Date.Truncate(time.Second).Equal(tx.Date.Truncate(time.Second)) {
        t.Errorf("Date = %v, want %v", found.Date, tx.Date)
    }
    if !found.CreatedAt.Truncate(time.Second).Equal(tx.CreatedAt.Truncate(time.Second)) {
        t.Errorf("CreatedAt = %v, want %v", found.CreatedAt, tx.CreatedAt)
    }
}

func TestTransactionRepository_FindByID_NotFound(t *testing.T) {
    txRepo, _, cleanup := setupTransactionRepoTest(t)
    defer cleanup()
    
    _, err := txRepo.FindByID("nonexistent")
    
    if err != domain.ErrNotFound {
        t.Errorf("FindByID() error = %v, want ErrNotFound", err)
    }
}

func TestTransactionRepository_FindAll(t *testing.T) {
    txRepo, catRepo, cleanup := setupTransactionRepoTest(t)
    defer cleanup()
    
    cat, _ := domain.NewCategory("Test", "#FF5733")
    catRepo.Save(*cat)
    
    // Create transactions on different dates
    tx1, _ := domain.NewTransaction(
        time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
        100.00,
        domain.TypeExpense,
        cat.ID,
        "First",
    )
    tx2, _ := domain.NewTransaction(
        time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
        200.00,
        domain.TypeExpense,
        cat.ID,
        "Second",
    )
    tx3, _ := domain.NewTransaction(
        time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
        300.00,
        domain.TypeExpense,
        cat.ID,
        "Third",
    )
    
    txRepo.Save(*tx1)
    txRepo.Save(*tx2)
    txRepo.Save(*tx3)
    
    all, err := txRepo.FindAll()
    if err != nil {
        t.Fatalf("FindAll() error = %v", err)
    }
    
    if len(all) != 3 {
        t.Errorf("FindAll() returned %d transactions, want 3", len(all))
    }
    
    // Verify ordering (newest first)
    if all[0].ID != tx2.ID {
        t.Errorf("First transaction ID = %v, want %v (newest)", all[0].ID, tx2.ID)
    }
    if all[1].ID != tx1.ID {
        t.Errorf("Second transaction ID = %v, want %v", all[1].ID, tx1.ID)
    }
    if all[2].ID != tx3.ID {
        t.Errorf("Third transaction ID = %v, want %v (oldest)", all[2].ID, tx3.ID)
    }
}

func TestTransactionRepository_FindByDateRange(t *testing.T) {
    txRepo, catRepo, cleanup := setupTransactionRepoTest(t)
    defer cleanup()
    
    cat, _ := domain.NewCategory("Test", "#FF5733")
    catRepo.Save(*cat)
    
    // Create transactions across different months
    jan1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
    jan15 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
    jan31 := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
    feb5 := time.Date(2026, 2, 5, 0, 0, 0, 0, time.UTC)
    
    tx1, _ := domain.NewTransaction(jan1, 100.00, domain.TypeExpense, cat.ID, "Jan 1")
    tx2, _ := domain.NewTransaction(jan15, 200.00, domain.TypeExpense, cat.ID, "Jan 15")
    tx3, _ := domain.NewTransaction(jan31, 300.00, domain.TypeExpense, cat.ID, "Jan 31")
    tx4, _ := domain.NewTransaction(feb5, 400.00, domain.TypeExpense, cat.ID, "Feb 5")
    
    txRepo.Save(*tx1)
    txRepo.Save(*tx2)
    txRepo.Save(*tx3)
    txRepo.Save(*tx4)
    
    // Query January only
    start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
    end := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)
    
    results, err := txRepo.FindByDateRange(start, end)
    if err != nil {
        t.Fatalf("FindByDateRange() error = %v", err)
    }
    
    // Should get Jan 1, 15, 31 (not Feb 5)
    if len(results) != 3 {
        t.Errorf("FindByDateRange() returned %d transactions, want 3", len(results))
    }
    
    // Verify Feb transaction is not included
    for _, tx := range results {
        if tx.ID == tx4.ID {
            t.Error("FindByDateRange() included Feb transaction")
        }
    }
}

func TestTransactionRepository_FindByCategory(t *testing.T) {
    txRepo, catRepo, cleanup := setupTransactionRepoTest(t)
    defer cleanup()
    
    // Create two categories
    groceries, _ := domain.NewCategory("Groceries", "#FF5733")
    rent, _ := domain.NewCategory("Rent", "#00FF00")
    catRepo.Save(*groceries)
    catRepo.Save(*rent)
    
    // Create transactions in different categories
    tx1, _ := domain.NewTransaction(time.Now(), 50.00, domain.TypeExpense, groceries.ID, "Food 1")
    tx2, _ := domain.NewTransaction(time.Now(), 60.00, domain.TypeExpense, groceries.ID, "Food 2")
    tx3, _ := domain.NewTransaction(time.Now(), 1200.00, domain.TypeExpense, rent.ID, "Rent")
    
    txRepo.Save(*tx1)
    txRepo.Save(*tx2)
    txRepo.Save(*tx3)
    
    // Query groceries only
    results, err := txRepo.FindByCategory(groceries.ID)
    if err != nil {
        t.Fatalf("FindByCategory() error = %v", err)
    }
    
    if len(results) != 2 {
        t.Errorf("FindByCategory() returned %d transactions, want 2", len(results))
    }
    
    // Verify all are groceries
    for _, tx := range results {
        if tx.CategoryID != groceries.ID {
            t.Errorf("FindByCategory() returned transaction with wrong category")
        }
    }
}

func TestTransactionRepository_Delete(t *testing.T) {
    txRepo, catRepo, cleanup := setupTransactionRepoTest(t)
    defer cleanup()
    
    cat, _ := domain.NewCategory("Test", "#FF5733")
    catRepo.Save(*cat)
    
    tx, _ := domain.NewTransaction(time.Now(), 50.00, domain.TypeExpense, cat.ID, "Test")
    txRepo.Save(*tx)
    
    // Delete
    if err := txRepo.Delete(tx.ID); err != nil {
        t.Fatalf("Delete() error = %v", err)
    }
    
    // Verify gone
    _, err := txRepo.FindByID(tx.ID)
    if err != domain.ErrNotFound {
        t.Errorf("FindByID() after delete error = %v, want ErrNotFound", err)
    }
}

func TestTransactionRepository_ForeignKeyConstraint(t *testing.T) {
    txRepo, _, cleanup := setupTransactionRepoTest(t)
    defer cleanup()
    
    // Try to save transaction with non-existent category
    tx, _ := domain.NewTransaction(
        time.Now(),
        50.00,
        domain.TypeExpense,
        "nonexistent_category",
        "Should fail",
    )
    
    err := txRepo.Save(*tx)
    
    // Should fail due to foreign key constraint
    if err == nil {
        t.Error("Save() with invalid category should fail, got nil error")
    }
}

func TestTransactionRepository_TimeHandling(t *testing.T) {
    txRepo, catRepo, cleanup := setupTransactionRepoTest(t)
    defer cleanup()
    
    cat, _ := domain.NewCategory("Test", "#FF5733")
    catRepo.Save(*cat)
    
    // Use specific timestamp with microseconds
    specificTime := time.Date(2026, 3, 10, 14, 30, 45, 123456789, time.UTC)
    
    tx, _ := domain.NewTransaction(
        specificTime,
        50.00,
        domain.TypeExpense,
        cat.ID,
        "Time test",
    )
    
    txRepo.Save(*tx)
    
    found, _ := txRepo.FindByID(tx.ID)
    
    // Times should be equal (RFC3339 preserves seconds, not nanoseconds)
    if !found.Date.Truncate(time.Second).Equal(specificTime.Truncate(time.Second)) {
        t.Errorf("Date mismatch: got %v, want %v", found.Date, specificTime)
    }
}