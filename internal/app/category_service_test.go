package app

import (
	"errors"
	"testing"
	"time"

	"github.com/simdangelo/fingo/internal/domain"
	"github.com/simdangelo/fingo/internal/memory"
)

func setupCategoryService(t *testing.T) *CategoryService {
	t.Helper()

	catRepo := memory.NewCategoryRepository()
	txRepo := memory.NewTransactionRepository()
	service := NewCategoryService(catRepo, txRepo)

	return service
}

func TestCategoryService_CreateCategory_Success(t *testing.T) {
	service := setupCategoryService(t)

	cat, err := service.CreateCategory("Groceries", "#FF5733")
	if err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}

	// Verify category was returned
	if cat == nil {
		t.Fatal("CreateCategory() returned nil category")
	}

	if cat.Name != "Groceries" {
		t.Errorf("Name = %v, want Groceries", cat.Name)
	}
	if cat.Color != "#FF5733" {
		t.Errorf("Color =%v, want FF5733", cat.Color)
	}
	if cat.ID == "" {
		t.Errorf("ID should not be empty")
	}

	found, _ := service.categoryRepo.FindByID(cat.ID)
	if found == nil {
		t.Errorf("Category was not saved to repository")
	}
}

func TestCategoryService_CreateCategory_InvalidData(t *testing.T) {
	service := setupCategoryService(t)

	tests := []struct {
		name    string
		catName string
		color   string
		wantErr bool
	}{
		{
			name:    "empty name",
			catName: "",
			color:   "#FF5733",
			wantErr: true,
		},
		{
			name:    "invalid color",
			catName: "Groceries",
			color:   "not-a-color",
			wantErr: true,
		},
		{
			name:    "name too long",
			catName: "This is a very long category name that exceeds the maximum allowed length",
			color:   "#FF5733",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateCategory(tt.catName, tt.color)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCategory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCategoryService_GetCategory(t *testing.T) {
	service := setupCategoryService(t)

	// Create a category
	created, _ := service.CreateCategory("Groceries", "#FF5733")

	// Retrieve it
	found, err := service.GetCategory(created.ID)
	if err != nil {
		t.Fatalf("GetCategory() error = %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("ID = %v, want %v", found.ID, created.ID)
	}

	if found.Name != created.Name {
		t.Errorf("Name = %v, want %v", found.Name, created.Name)
	}
}

func TestCategoryService_GetCategory_NotFound(t *testing.T) {
	service := setupCategoryService(t)

	_, err := service.GetCategory("nonexistent")

	if err == nil {
		t.Fatal("GetCategory() expected error, got nil")
	}

	if !errors.Is(err, ErrCategoryNotFound) {
		t.Errorf("GetCategory() error = %v, want ErrCategoryNotFound", err)
	}
}

func TestCategoryService_ListCategories(t *testing.T) {
	service := setupCategoryService(t)

	// Create multiple categories
	service.CreateCategory("Groceries", "#FF5733")
	service.CreateCategory("Rent", "#00FF00")
	service.CreateCategory("Transport", "#0000FF")

	// List all
	categories, err := service.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}

	if len(categories) != 3 {
		t.Errorf("ListCategories() returned %d categories, want 3", len(categories))
	}
}

func TestCategoryService_ListCategories_Empty(t *testing.T) {
	service := setupCategoryService(t)

	categories, err := service.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}

	if categories == nil {
		t.Error("ListCategories() should return empty slice, not nil")
	}

	if len(categories) != 0 {
		t.Errorf("ListCategories() returned %d categories, want 0", len(categories))
	}
}

func TestCategoryService_CanDeleteCategory_NoTransactions(t *testing.T) {
	service := setupCategoryService(t)

	// Create category with no transactions
	cat, _ := service.CreateCategory("Unused", "#FF5733")

	canDelete, err := service.CanDeleteCategory(cat.ID)
	if err != nil {
		t.Fatalf("CanDeleteCategory() error = %v", err)
	}

	if !canDelete {
		t.Error("CanDeleteCategory() = false, want true (no transactions)")
	}
}

func TestCategoryService_CanDeleteCategory_HasTransactions(t *testing.T) {
	service := setupCategoryService(t)

	// Create category
	cat, _ := service.CreateCategory("Groceries", "#FF5733")

	// Add a transaction using this category
	tx, _ := domain.NewTransaction(
		time.Now(),
		50.00,
		domain.TypeExpense,
		cat.ID,
		"Shopping",
	)
	service.transactionRepo.Save(*tx)

	// Check if can delete
	canDelete, err := service.CanDeleteCategory(cat.ID)
	if err != nil {
		t.Fatalf("CanDeleteCategory() error = %v", err)
	}

	if canDelete {
		t.Error("CanDeleteCategory() = true, want false (has transactions)")
	}
}

func TestCategoryService_CanDeleteCategory_NotFound(t *testing.T) {
	service := setupCategoryService(t)

	_, err := service.CanDeleteCategory("nonexistent")

	if err == nil {
		t.Fatal("CanDeleteCategory() expected error, got nil")
	}

	if !errors.Is(err, ErrCategoryNotFound) {
		t.Errorf("CanDeleteCategory() error = %v, want ErrCategoryNotFound", err)
	}
}

func TestCategoryService_DeleteCategory_Success(t *testing.T) {
	service := setupCategoryService(t)

	// Create category with no transactions
	cat, _ := service.CreateCategory("Unused", "#FF5733")

	// Delete should succeed
	err := service.DeleteCategory(cat.ID)
	if err != nil {
		t.Errorf("DeleteCategory() error = %v", err)
	}

	// Verify it's gone
	_, err = service.categoryRepo.FindByID(cat.ID)
	if err == nil {
		t.Error("Category should be deleted but still exists")
	}
}

func TestCategoryService_DeleteCategory_WithTransactions(t *testing.T) {
	service := setupCategoryService(t)

	// Create category
	cat, _ := service.CreateCategory("Groceries", "#FF5733")

	// Add transaction
	tx, _ := domain.NewTransaction(
		time.Now(),
		50.00,
		domain.TypeExpense,
		cat.ID,
		"Shopping",
	)
	service.transactionRepo.Save(*tx)

	// Delete should fail
	err := service.DeleteCategory(cat.ID)

	if err == nil {
		t.Fatal("DeleteCategory() expected error, got nil")
	}

	if !errors.Is(err, ErrCategoryInUse) {
		t.Errorf("DeleteCategory() error = %v, want ErrCategoryInUse", err)
	}

	// Verify category still exists
	found, _ := service.categoryRepo.FindByID(cat.ID)
	if found == nil {
		t.Error("Category should still exist after failed delete")
	}
}

func TestCategoryService_DeleteCategory_CompleteWorkflow(t *testing.T) {
	service := setupCategoryService(t)

	// Create category
	cat, _ := service.CreateCategory("Temp", "#FF5733")

	// Add transaction
	tx, _ := domain.NewTransaction(
		time.Now(),
		50.00,
		domain.TypeExpense,
		cat.ID,
		"Test",
	)
	service.transactionRepo.Save(*tx)

	// Try to delete - should fail
	err := service.DeleteCategory(cat.ID)
	if !errors.Is(err, ErrCategoryInUse) {
		t.Errorf("DeleteCategory() error = %v, want ErrCategoryInUse", err)
	}

	// Remove the transaction
	service.transactionRepo.Delete(tx.ID)

	// Now delete should succeed
	err = service.DeleteCategory(cat.ID)
	if err != nil {
		t.Errorf("DeleteCategory() after removing transactions error = %v", err)
	}
}
