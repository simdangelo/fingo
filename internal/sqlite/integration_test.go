package sqlite

import (
	"errors"
	"testing"
	"time"

	"github.com/simdangelo/fingo/internal/app"
	"github.com/simdangelo/fingo/internal/domain"
)

// setupIntegration creates services with SQLite repositories.
func setupIntegration(t *testing.T) (*app.TransactionService, *app.CategoryService, func()) {
    t.Helper()
    
    db, err := OpenInMemory()
    if err != nil {
        t.Fatalf("failed to open database: %v", err)
    }
    
    if err := Migrate(db); err != nil {
        db.Close()
        t.Fatalf("failed to migrate: %v", err)
    }
    
    // Create SQLite repositories
    catRepo := NewCategoryRepository(db)
    txRepo := NewTransactionRepository(db)
    
    // Create services with SQLite repos
    txService := app.NewTransactionService(txRepo, catRepo)
    catService := app.NewCategoryService(catRepo, txRepo)
    
    cleanup := func() {
        db.Close()
    }
    
    return txService, catService, cleanup
}

func TestIntegration_CompleteWorkflow(t *testing.T) {
    txService, catService, cleanup := setupIntegration(t)
    defer cleanup()
    
    // Step 1: Create categories
    groceries, err := catService.CreateCategory("Groceries", "#FF5733")
    if err != nil {
        t.Fatalf("CreateCategory() error = %v", err)
    }
    
    salary, err := catService.CreateCategory("Salary", "#00FF00")
    if err != nil {
        t.Fatalf("CreateCategory() error = %v", err)
    }
    
    // Step 2: Add transactions
    jan := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
    
    err = txService.AddTransaction(jan, 100.00, domain.TypeExpense, groceries.ID, "Food")
    if err != nil {
        t.Fatalf("AddTransaction() error = %v", err)
    }
    
    err = txService.AddTransaction(jan, 150.00, domain.TypeExpense, groceries.ID, "More food")
    if err != nil {
        t.Fatalf("AddTransaction() error = %v", err)
    }
    
    err = txService.AddTransaction(jan, 3500.00, domain.TypeIncome, salary.ID, "Monthly salary")
    if err != nil {
        t.Fatalf("AddTransaction() error = %v", err)
    }
    
    // Step 3: Get monthly summary
    summary, err := txService.GetMonthlySummary(2026, time.January)
    if err != nil {
        t.Fatalf("GetMonthlySummary() error = %v", err)
    }
    
    // Verify summary
    if summary.TotalIncome != 3500.00 {
        t.Errorf("TotalIncome = %v, want 3500.00", summary.TotalIncome)
    }
    
    if summary.TotalExpense != 250.00 {
        t.Errorf("TotalExpense = %v, want 250.00", summary.TotalExpense)
    }
    
    if summary.NetAmount != 3250.00 {
        t.Errorf("NetAmount = %v, want 3250.00", summary.NetAmount)
    }
    
    if summary.TransactionCount != 3 {
        t.Errorf("TransactionCount = %v, want 3", summary.TransactionCount)
    }
    
    // Step 4: Check category totals
    groceriesTotal, _ := txService.CalculateCategoryTotal(groceries.ID)
    if groceriesTotal != -250.00 {
        t.Errorf("Groceries total = %v, want -250.00", groceriesTotal)
    }
    
    salaryTotal, _ := txService.CalculateCategoryTotal(salary.ID)
    if salaryTotal != 3500.00 {
        t.Errorf("Salary total = %v, want 3500.00", salaryTotal)
    }
}

func TestIntegration_CategoryDeletionProtection(t *testing.T) {
    txService, catService, cleanup := setupIntegration(t)
    defer cleanup()
    
    // Create category
    cat, _ := catService.CreateCategory("Groceries", "#FF5733")
    
    // Try to delete - should succeed (no transactions)
    err := catService.DeleteCategory(cat.ID)
    if err != nil {
        t.Errorf("DeleteCategory() should succeed with no transactions, got error: %v", err)
    }
    
    // Recreate category
    cat, _ = catService.CreateCategory("Groceries", "#FF5733")
    
    // Add a transaction
    err = txService.AddTransaction(
        time.Now(),
        50.00,
        domain.TypeExpense,
        cat.ID,
        "Shopping",
    )
    if err != nil {
        t.Fatalf("AddTransaction() error = %v", err)
    }
    
    // Try to delete - should fail (has transactions)
    err = catService.DeleteCategory(cat.ID)
    if err == nil {
        t.Fatal("DeleteCategory() should fail with transactions, got nil")
    }
    
    if !errors.Is(err, app.ErrCategoryInUse) {
        t.Errorf("DeleteCategory() error = %v, want ErrCategoryInUse", err)
    }
    
    // Verify category still exists
    found, err := catService.GetCategory(cat.ID)
    if err != nil {
        t.Error("Category should still exist after failed delete")
    }
    if found == nil {
        t.Error("Category was deleted despite having transactions")
    }
}

func TestIntegration_Persistence(t *testing.T) {
    // Create a file-based database to test persistence
    tmpDir := t.TempDir()
    dbPath := tmpDir + "/test.db"
    
    // Session 1: Create and save data
    db1, err := InitDB(dbPath)
    if err != nil {
        t.Fatalf("InitDB() error = %v", err)
    }
    
    catRepo1 := NewCategoryRepository(db1)
    txRepo1 := NewTransactionRepository(db1)
    catService1 := app.NewCategoryService(catRepo1, txRepo1)
    txService1 := app.NewTransactionService(txRepo1, catRepo1)
    
    // Create category and transaction
    cat, _ := catService1.CreateCategory("Test Category", "#FF5733")
    txService1.AddTransaction(
        time.Now(),
        123.45,
        domain.TypeExpense,
        cat.ID,
        "Test transaction",
    )
    
    categoryID := cat.ID
    
    // Close database
    Close(db1)
    
    // Session 2: Reopen and verify data persists
    db2, err := InitDB(dbPath)
    if err != nil {
        t.Fatalf("InitDB() error = %v", err)
    }
    defer Close(db2)
    
    catRepo2 := NewCategoryRepository(db2)
    txRepo2 := NewTransactionRepository(db2)
    catService2 := app.NewCategoryService(catRepo2, txRepo2)
    
    // Verify category persisted
    foundCat, err := catService2.GetCategory(categoryID)
    if err != nil {
        t.Fatalf("GetCategory() error = %v (data didn't persist)", err)
    }
    
    if foundCat.Name != "Test Category" {
        t.Errorf("Category name = %v, want 'Test Category'", foundCat.Name)
    }
    
    // Verify transaction persisted
    txs, _ := txRepo2.FindByCategory(categoryID)
    if len(txs) != 1 {
        t.Errorf("Found %d transactions, want 1", len(txs))
    }
    
    if len(txs) > 0 && txs[0].Amount != 123.45 {
        t.Errorf("Transaction amount = %v, want 123.45", txs[0].Amount)
    }
}

func TestIntegration_MonthlyComparison(t *testing.T) {
    txService, catService, cleanup := setupIntegration(t)
    defer cleanup()
    
    cat, _ := catService.CreateCategory("Groceries", "#FF5733")
    
    // Add January transactions
    jan15 := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
    txService.AddTransaction(jan15, 200.00, domain.TypeExpense, cat.ID, "Jan groceries")
    
    // Add February transactions
    feb15 := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
    txService.AddTransaction(feb15, 300.00, domain.TypeExpense, cat.ID, "Feb groceries")
    
    // Get summaries
    janSummary, _ := txService.GetMonthlySummary(2026, time.January)
    febSummary, _ := txService.GetMonthlySummary(2026, time.February)
    
    // Verify January
    if janSummary.TotalExpense != 200.00 {
        t.Errorf("Jan TotalExpense = %v, want 200.00", janSummary.TotalExpense)
    }
    
    // Verify February
    if febSummary.TotalExpense != 300.00 {
        t.Errorf("Feb TotalExpense = %v, want 300.00", febSummary.TotalExpense)
    }
    
    // Verify months are independent
    if janSummary.TransactionCount != 1 {
        t.Errorf("Jan TransactionCount = %v, want 1", janSummary.TransactionCount)
    }
    
    if febSummary.TransactionCount != 1 {
        t.Errorf("Feb TransactionCount = %v, want 1", febSummary.TransactionCount)
    }
}

func TestIntegration_MultiCategorySpending(t *testing.T) {
    txService, catService, cleanup := setupIntegration(t)
    defer cleanup()
    
    // Create multiple categories
    groceries, _ := catService.CreateCategory("Groceries", "#FF5733")
    rent, _ := catService.CreateCategory("Rent", "#00FF00")
    transport, _ := catService.CreateCategory("Transport", "#0000FF")
    
    // Add transactions across categories
    jan := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
    
    txService.AddTransaction(jan, 100.00, domain.TypeExpense, groceries.ID, "Food 1")
    txService.AddTransaction(jan, 150.00, domain.TypeExpense, groceries.ID, "Food 2")
    txService.AddTransaction(jan, 1200.00, domain.TypeExpense, rent.ID, "Monthly rent")
    txService.AddTransaction(jan, 50.00, domain.TypeExpense, transport.ID, "Gas")
    txService.AddTransaction(jan, 75.00, domain.TypeExpense, transport.ID, "Bus pass")
    
    // Get all categories
    categories, _ := catService.ListCategories()
    
    // Calculate spending per category
    expected := map[string]float64{
        "Groceries": -250.00,
        "Rent":      -1200.00,
        "Transport": -125.00,
    }
    
    for _, cat := range categories {
        total, _ := txService.CalculateCategoryTotal(cat.ID)
        
        if exp, ok := expected[cat.Name]; ok {
            if total != exp {
                t.Errorf("%s total = %v, want %v", cat.Name, total, exp)
            }
        }
    }
}

func TestIntegration_InvalidCategoryReference(t *testing.T) {
    txService, _, cleanup := setupIntegration(t)
    defer cleanup()
    
    // Try to add transaction with non-existent category
    err := txService.AddTransaction(
        time.Now(),
        50.00,
        domain.TypeExpense,
        "nonexistent_category",
        "Should fail",
    )
    
    if err == nil {
        t.Fatal("AddTransaction() should fail with invalid category")
    }
    
    if !errors.Is(err, app.ErrCategoryNotFound) {
        t.Errorf("AddTransaction() error = %v, want ErrCategoryNotFound", err)
    }
    
    // Verify no transaction was created
    all, _ := txService.GetAllTransactions()
    if len(all) != 0 {
        t.Error("No transaction should be created with invalid category")
    }
}

func TestIntegration_EmptyState(t *testing.T) {
    txService, catService, cleanup := setupIntegration(t)
    defer cleanup()
    
    // No categories exist
    categories, err := catService.ListCategories()
    if err != nil {
        t.Fatalf("ListCategories() error = %v", err)
    }
    
    if len(categories) != 0 {
        t.Errorf("Expected 0 categories, got %d", len(categories))
    }
    
    // No transactions exist
    transactions, err := txService.GetAllTransactions()
    if err != nil {
        t.Fatalf("GetAllTransactions() error = %v", err)
    }
    
    if len(transactions) != 0 {
        t.Errorf("Expected 0 transactions, got %d", len(transactions))
    }
    
    // Monthly summary should be zero
    summary, err := txService.GetMonthlySummary(2026, time.January)
    if err != nil {
        t.Fatalf("GetMonthlySummary() error = %v", err)
    }
    
    if summary.TotalIncome != 0 || summary.TotalExpense != 0 || summary.NetAmount != 0 {
        t.Error("Empty state should have zero summary")
    }
}
