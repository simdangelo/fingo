package app

import (
	"errors"
	"testing"
	"time"

	"github.com/simdangelo/fingo/internal/domain"
	"github.com/simdangelo/fingo/internal/memory"
)

// setupIntegrationTest creates both services sharing the same repositories.
func setupIntegrationTest(t *testing.T) (*TransactionService, *CategoryService, *memory.CategoryRepository, *memory.TransactionRepository) {
	t.Helper()

	catRepo := memory.NewCategoryRepository()
	txRepo := memory.NewTransactionRepository()

	txService := NewTransactionService(txRepo, catRepo)
	catService := NewCategoryService(catRepo, txRepo)

	return txService, catService, catRepo, txRepo
}

func TestIntegration_CompleteTransactionWorkflow(t *testing.T) {
	txService, catService, _, _ := setupIntegrationTest(t)

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

func TestIntegration_CategoryProtection(t *testing.T) {
	txService, catService, _, _ := setupIntegrationTest(t)

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

	if !errors.Is(err, ErrCategoryInUse) {
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

func TestIntegration_MultiCategoryAnalysis(t *testing.T) {
	txService, catService, _, _ := setupIntegrationTest(t)

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
	type CategorySpending struct {
		Name  string
		Total float64
	}

	var spending []CategorySpending
	for _, cat := range categories {
		total, _ := txService.CalculateCategoryTotal(cat.ID)
		spending = append(spending, CategorySpending{
			Name:  cat.Name,
			Total: total,
		})
	}

	// Verify we have all three categories
	if len(spending) != 3 {
		t.Fatalf("Expected 3 categories, got %d", len(spending))
	}

	// Verify totals
	expected := map[string]float64{
		"Groceries": -250.00,
		"Rent":      -1200.00,
		"Transport": -125.00,
	}

	for _, s := range spending {
		if exp, ok := expected[s.Name]; ok {
			if s.Total != exp {
				t.Errorf("%s total = %v, want %v", s.Name, s.Total, exp)
			}
		}
	}
}

func TestIntegration_MonthlyComparison(t *testing.T) {
	txService, catService, _, _ := setupIntegrationTest(t)

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

func TestIntegration_CategoryDeletionAfterTransactionRemoval(t *testing.T) {
	txService, catService, _, txRepo := setupIntegrationTest(t)

	// Create category
	cat, _ := catService.CreateCategory("Temporary", "#FF5733")

	// Add transaction
	err := txService.AddTransaction(
		time.Now(),
		50.00,
		domain.TypeExpense,
		cat.ID,
		"Test",
	)
	if err != nil {
		t.Fatalf("AddTransaction() error = %v", err)
	}

	// Get the transaction ID
	txs, _ := txService.GetCategoryTransactions(cat.ID)
	txID := txs[0].ID

	// Can't delete category yet
	canDelete, _ := catService.CanDeleteCategory(cat.ID)
	if canDelete {
		t.Error("CanDeleteCategory() = true, want false (has transactions)")
	}

	// Remove the transaction
	err = txRepo.Delete(txID)
	if err != nil {
		t.Fatalf("Delete transaction error = %v", err)
	}

	// Now should be able to delete
	canDelete, _ = catService.CanDeleteCategory(cat.ID)
	if !canDelete {
		t.Error("CanDeleteCategory() = false, want true (transactions removed)")
	}

	// Delete should succeed
	err = catService.DeleteCategory(cat.ID)
	if err != nil {
		t.Errorf("DeleteCategory() error = %v", err)
	}

	// Verify deleted
	_, err = catService.GetCategory(cat.ID)
	if err == nil {
		t.Error("Category should be deleted")
	}
}

func TestIntegration_InvalidCategoryReference(t *testing.T) {
	txService, _, _, _ := setupIntegrationTest(t)

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

	if !errors.Is(err, ErrCategoryNotFound) {
		t.Errorf("AddTransaction() error = %v, want ErrCategoryNotFound", err)
	}

	// Verify no transaction was created
	all, _ := txService.GetAllTransactions()
	if len(all) != 0 {
		t.Error("No transaction should be created with invalid category")
	}
}

func TestIntegration_EmptyState(t *testing.T) {
	txService, catService, _, _ := setupIntegrationTest(t)

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
