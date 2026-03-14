package app

import (
	"errors"
	"testing"
	"time"

	"github.com/simdangelo/fingo/internal/domain"
	"github.com/simdangelo/fingo/internal/memory"
)

// setupTestService creates a service with in-memory repositories for testing.
// When a test fails, Go shows the line number where the failure occurred. By calling t.Helper(), you tell Go: "This is a helper function, show the caller's line number instead."
// Without t.Helper():
//
//	func setup(t *testing.T) {
//	    require(someCondition)  // If this fails, shows line in setup()
//	}
//
// With t.Helper():
//
//	func setup(t *testing.T) {
//	    t.Helper()
//	    require(someCondition)  // If this fails, shows line in test function
//	}
func setupTestService(t *testing.T) (*TransactionService, *memory.CategoryRepository, *memory.TransactionRepository) {
	t.Helper() // Marks this as a test helper

	txRepo := memory.NewTransactionRepository()
	catRepo := memory.NewCategoryRepository()

	service := NewTransactionService(txRepo, catRepo)
	return service, catRepo, txRepo
}

// Test 1: AddTransaction Success
func TestTransactionService_AddTransaction_Success(t *testing.T) {
	service, catRepo, txRepo := setupTestService(t)

	// Create a category first
	cat, err := domain.NewCategory("Groceries", "#FF5733")
	if err != nil {
		t.Fatalf("failed to create category %v", err)
	}
	catRepo.Save(*cat)

	err = service.AddTransaction(time.Now(), 45.50, domain.TypeExpense, cat.ID, "Weekly shopping")
	if err != nil {
		t.Errorf("AddTransaction() error %v, want nil", err)
	}

	// Verify transaction was saved
	all, _ := txRepo.FindAll()
	if len(all) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(all))
	}

	// Verify transaction details
	tx := all[0]
	if tx.Amount != 45.50 {
		t.Errorf("Amount: got %v, want %v", tx.Amount, 45.50)
	}
	if tx.Type != domain.TypeExpense {
		t.Errorf("Type: got %v, want %v", tx.Type, domain.TypeExpense)
	}
	if tx.CategoryID != cat.ID {
		t.Errorf("CategoryID: got %v, want %v", tx.CategoryID, cat.ID)
	}
}

// Test 2: AddTransaction with Nonexistent Category
func TestTransactionService_AddTransaction_CategoryNotFound(t *testing.T) {
	service, _, _ := setupTestService(t)

	err := service.AddTransaction(time.Now(), 45.50, domain.TypeExpense, "nonexistent_cateogory", "should fail")
	if err == nil {
		t.Errorf("AddTransaction() should fail, got nil")
	}

	if !errors.Is(err, ErrCategoryNotFound) {
		t.Errorf("AddTransaction() error %v, want ErrCategoryNotFound", err)
	}
}

// Test 3: AddTransaction with Invalid Transaction Data
func TestTransactionService_AddTransaction_InvalidTransaction(t *testing.T) {
	service, catRepo, _ := setupTestService(t)

	cat, _ := domain.NewCategory("Groceries", "#FF5733")
	catRepo.Save(*cat)

	err := service.AddTransaction(time.Now(), 0, domain.TypeExpense, cat.ID, "Invalid")

	if err == nil {
		t.Errorf("AddTransaction() should fail because Amount == 0, got nil")
	}
}

// Test 4: GetTransaction
func TestTransactionService_GetTransaction(t *testing.T) {
	service, catRepo, _ := setupTestService(t)

	cat, _ := domain.NewCategory("Groceries", "#FF5733")
	catRepo.Save(*cat)

	service.AddTransaction(time.Now(), 45.50, domain.TypeExpense, cat.ID, "Test transaction")

	transactions, _ := service.GetAllTransactions()
	txID := transactions[0].ID

	transaction, err := service.GetTransaction(txID)
	if err != nil {
		t.Fatalf("GetTransaction() error = %v", err)
	}

	if transaction.ID != txID {
		t.Errorf("GetTransaction() ID = %v, want %v", transaction.ID, txID)
	}
}

// Test 5: GetTransaction Not Found
func TestTransactionService_GetTransaction_NotFound(t *testing.T) {
	service, _, _ := setupTestService(t)

	_, err := service.GetTransaction("non_existent_ID")
	if err == nil {
		t.Fatalf("GetTransaction: expected error, got nil")
	}

	if !errors.Is(err, ErrTransactionNotFound) {
		t.Errorf("GetTransaction() error %v, want ErrTransactionNotFound", err)
	}
}

// Test 6: GetTransactionsByMonth
func TestTransactionService_GetTransactionsByMonth(t *testing.T) {
	service, _, _ := setupTestService(t)

	cat, _ := domain.NewCategory("Groceries", "#FF5733")
	service.categoryRepo.Save(*cat)

	jan15 := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	jan20 := time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)
	feb10 := time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC)

	service.AddTransaction(jan15, 100.00, domain.TypeExpense, cat.ID, "Jan tx 1")
	service.AddTransaction(jan20, 150.00, domain.TypeExpense, cat.ID, "Jan tx 2")
	service.AddTransaction(feb10, 200.00, domain.TypeExpense, cat.ID, "Feb tx 1")

	transactions_jan, err := service.GetTransactionByMonth(2026, time.January)
	if err != nil {
		t.Fatalf("GetTransactionByMonth() error = %v", err)
	}

	if len(transactions_jan) != 2 {
		t.Errorf("GetTransactionByMonth() on January returns %d transactions, want 2", len(transactions_jan))
	}

	// Verify all are from January
	for _, tx := range transactions_jan {
		if tx.Date.Month() != time.January {
			t.Errorf("Transaction date month = %v, want January", tx.Date.Month())
		}
		if tx.Date.Year() != 2026 {
			t.Errorf("Transaction date year = %v, want 2026", tx.Date.Year())
		}
	}
}

// Test 7: GetTransactionsByDateRange
func TestTransactionService_GetTransactionsByDateRange(t *testing.T) {
	service, _, _ := setupTestService(t)

	cat, _ := domain.NewCategory("Test", "#FF5733")
	service.categoryRepo.Save(*cat)

	// Add transactions across a range
	jan1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	jan15 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	jan31 := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	feb5 := time.Date(2026, 2, 5, 0, 0, 0, 0, time.UTC)

	service.AddTransaction(jan1, 100.00, domain.TypeExpense, cat.ID, "Jan 1")
	service.AddTransaction(jan15, 150.00, domain.TypeExpense, cat.ID, "Jan 15")
	service.AddTransaction(jan31, 200.00, domain.TypeExpense, cat.ID, "Jan 31")
	service.AddTransaction(feb5, 250.00, domain.TypeExpense, cat.ID, "Feb 5")

	start := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)

	transactions, err := service.GetTransactionByDateRange(start, end)
	if err != nil {
		t.Fatalf("GetTransactionByDateRange() error: %v", err)
	}

	if len(transactions) != 2 {
		t.Errorf("GetTransactionByDateRange() returned %d transactions, want 2", len(transactions))
	}
}

// Test 8: GetCategoryTransactions
func TestTransactionService_GetCategoryTransactions(t *testing.T) {
	service, _, _ := setupTestService(t)

	groceries, _ := domain.NewCategory("Groceries", "#FF5733")
	rent, _ := domain.NewCategory("Rent", "#00FF00")
	service.categoryRepo.Save(*groceries)
	service.categoryRepo.Save(*rent)

	service.AddTransaction(time.Now(), 50.00, domain.TypeExpense, groceries.ID, "Food")
	service.AddTransaction(time.Now(), 60.00, domain.TypeExpense, groceries.ID, "More Food")
	service.AddTransaction(time.Now(), 1200.00, domain.TypeExpense, rent.ID, "Monthly rent")

	groceryTxs, err := service.GetCategoryTransactions(groceries.ID)
	if err != nil {
		t.Fatalf("GetCategoryTransactions() error = %v", err)
	}

	if len(groceryTxs) != 2 {
		t.Errorf("GetCategoryTransactions() for 'Groceries' returned %d transactions, want 2", len(groceryTxs))
	}

	for _, tx := range groceryTxs {
		if tx.CategoryID != groceries.ID {
			t.Errorf("Transaction category %v, want %v", tx.CategoryID, groceries.ID)
		}
	}
}

// func TestTransactionService_GetCategoryTransactions2(t *testing.T) {
// 	service, _, _ := setupTestService(t)

// 	_, err := service.GetCategoryTransactions("non_existent")
// 	if err == nil {
// 		t.Fatalf("GetCategoryTransactions() expected error, got nil")
// 	}
// }

// Test 9: CalculateCategoryTotal
func TestTransactionService_CalculateCategoryTotal(t *testing.T) {
	service, _, _ := setupTestService(t)

	groceries, _ := domain.NewCategory("Groceries", "#FF5733")
	rent, _ := domain.NewCategory("Rent", "#00FF00")
	service.categoryRepo.Save(*groceries)
	service.categoryRepo.Save(*rent)

	service.AddTransaction(time.Now(), 200.00, domain.TypeExpense, groceries.ID, "Groceries 1")
	service.AddTransaction(time.Now(), 400.00, domain.TypeExpense, groceries.ID, "Groceries 2")
	service.AddTransaction(time.Now(), 350.00, domain.TypeExpense, groceries.ID, "Groceries 3")
	service.AddTransaction(time.Now(), 400.00, domain.TypeExpense, rent.ID, "Rent 1")

	total, err := service.CalculateCategoryTotal(groceries.ID)
	if err != nil {
		t.Fatalf("CalculateCategoryTotal() err %v", err)
	}

	expected := -950.00
	if total != expected {
		t.Errorf("CalculateCategoryTotal() = %v, want %v", total, expected)
	}
}

func TestTransactionService_CalculateCategoryTotal_MixedTypes(t *testing.T) {
	service, _, _ := setupTestService(t)

	bank, _ := domain.NewCategory("Bank", "#FF5733")
	rent, _ := domain.NewCategory("Rent", "#00FF00")
	service.categoryRepo.Save(*bank)
	service.categoryRepo.Save(*rent)

	service.AddTransaction(time.Now(), 200.00, domain.TypeIncome, bank.ID, "Bank deposit 1")
	service.AddTransaction(time.Now(), 400.00, domain.TypeIncome, bank.ID, "Bank deposit 2")
	service.AddTransaction(time.Now(), 350.00, domain.TypeExpense, bank.ID, "Bank Fee 1")
	service.AddTransaction(time.Now(), 400.00, domain.TypeExpense, rent.ID, "Rent 1")

	total, err := service.CalculateCategoryTotal(bank.ID)
	if err != nil {
		t.Fatalf("CalculateCategoryTotal() err %v", err)
	}

	expected := 250.00
	if total != expected {
		t.Errorf("CalculateCategoryTotal() = %v, want %v", total, expected)
	}
}

// Test 11: GetMonthlySummary
func TestTransactionService_GetMonthlySummary(t *testing.T) {
	service, _, _ := setupTestService(t)

	groceries, _ := domain.NewCategory("Groceries", "#FF5733")
	salary, _ := domain.NewCategory("Salary", "#00FF00")
	service.categoryRepo.Save(*groceries)
	service.categoryRepo.Save(*salary)

	jan := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
	service.AddTransaction(jan, 100.00, domain.TypeExpense, groceries.ID, "Food")
	service.AddTransaction(jan, 400.00, domain.TypeExpense, groceries.ID, "More Food")
	service.AddTransaction(jan, 3000.00, domain.TypeIncome, salary.ID, "Salary")

	service.AddTransaction(feb, 500.00, domain.TypeExpense, groceries.ID, "Food")
	service.AddTransaction(feb, 3000.00, domain.TypeIncome, salary.ID, "Salary")

	summary, err := service.GetMonthlySummary(2026, time.January)
	if err != nil {
		t.Fatalf("GetMonthlySummary() error = %v", err)
	}

	if summary.Year != 2026 {
		t.Errorf("Year: %v, want 2026", summary.Year)
	}
	if summary.Month != time.January {
		t.Errorf("Month: %v, want January", summary.Month)
	}
	if summary.TotalIncome != 3000.00 {
		t.Errorf("TotalIncome: %f, want 3000.00", summary.TotalIncome)
	}
	if summary.TotalExpense != 500.00 {
		t.Errorf("TotalExpense: %f, want 500.00", summary.TotalExpense)
	}
	if summary.NetAmount != 2500.00 {
		t.Errorf("NetAmount: %f, want 2500.00", summary.NetAmount)
	}
	if float64(summary.TransactionCount) != 3 {
		t.Errorf("NetAmount: %d, want 3", summary.TransactionCount)
	}
}

func TestTransactionService_GetMonthlySummary_EmptyMonth(t *testing.T) {
	service, _, _ := setupTestService(t)

	summary, err := service.GetMonthlySummary(2026, time.March)
    if err != nil {
        t.Fatalf("GetMonthlySummary() error = %v", err)
    }
    
    if summary.TotalIncome != 0 {
        t.Errorf("TotalIncome = %v, want 0", summary.TotalIncome)
    }
    if summary.TotalExpense != 0 {
        t.Errorf("TotalExpense = %v, want 0", summary.TotalExpense)
    }
    if summary.NetAmount != 0 {
        t.Errorf("NetAmount = %v, want 0", summary.NetAmount)
    }
    if summary.TransactionCount != 0 {
        t.Errorf("TransactionCount = %v, want 0", summary.TransactionCount)
    }
}