package memory

import (
	"errors"
	"testing"

	"github.com/simdangelo/fingo/internal/domain"
)

func TestCategoryRepository_Save_And_FindByID(t *testing.T) {
	repo := NewCategoryRepository()
	cat, err := domain.NewCategory("Groceries", "#FF5733")

	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	err = repo.Save(*cat)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	found, err := repo.FindByID(cat.ID)
	if err != nil {
		t.Errorf("FindByID() error = %v", err)
	}

	if found == nil {
		t.Fatal("FindByID() returned nil")
	}

	if found.ID != cat.ID {
		t.Errorf("ID: got %v, want %v", found.ID, cat.ID)
	}

	if found.Name != cat.Name {
		t.Errorf("Name: got %v, want %v", found.Name, cat.Name)
	}

	if found.Color != cat.Color {
		t.Errorf("Color: got %v, want %v", found.Color, cat.Color)
	}
}

func TestCategoryRepository_FindByID(t *testing.T) {
	repo := NewCategoryRepository()

	found, err := repo.FindByID("nonexistent_ID")

	if err == nil {
		t.Errorf("FindByID() expected error, got nil")
	}

	if found != nil {
		t.Errorf("FindByID expected nil category, got %v", found)
	}

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("FindByID() error = %v, want ErrNotFound", err)
	}
}

func TestCategoryRepository_Save_UpdateExisting(t *testing.T) {
	repo := NewCategoryRepository()

	cat, _ := domain.NewCategory("Groceries", "#FF5733")
	repo.Save(*cat)

	// Modify and save again
	cat.Name = "Updated Groceries"
	cat.Color = "#00F00"
	repo.Save(*cat)

	found, err := repo.FindByID(cat.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.Name != "Updated Groceries" {
		t.Errorf("Name: got %v, want 'Updated Groceries'", found.Name)
	}

	if found.Color != "#00F00" {
		t.Errorf("Color: got %v, want '#00F00'", found.Color)
	}
}

func TestCategoryRepository_FindAll(t *testing.T) {
	repo := NewCategoryRepository()

	cat1, _ := domain.NewCategory("Groceries", "#FF5733")
	cat2, _ := domain.NewCategory("Rent", "#00FF00")
	cat3, _ := domain.NewCategory("Transport", "#0000FF")

	repo.Save(*cat1)
	repo.Save(*cat2)
	repo.Save(*cat3)

	all, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}

	if len(all) != 3 {
		t.Errorf("FindAll() returned %v number of categories, expected 3", len(all))
	}

	ids := make(map[string]bool)
	for _, cat := range all {
		ids[cat.ID] = true
	}

	if !ids[cat1.ID] || !ids[cat2.ID] || !ids[cat3.ID] {
		t.Error("FindAll() didn't return all categories")
	}
}

func TestCategoryRepository_FindAll_Empty(t *testing.T) {
	repo := NewCategoryRepository()

	all, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll(): expected no error, got = %v", err)
	}

	if all == nil {
		t.Errorf("FindAll() should return empty slice, not nil")
	}

	if len(all) != 0 {
		t.Errorf("FindAll() returned %d categories, want 0", len(all))
	}
}

func TestCategoryRepository_Delete(t *testing.T) {
	repo := NewCategoryRepository()

	cat, _ := domain.NewCategory("Groceries", "#FF5733")
	repo.Save(*cat)

	err := repo.Delete(cat.ID)
	if err != nil {
		t.Errorf("Delete() should return nil, got %v", err)
	}

	found, err := repo.FindByID(cat.ID)
	if err == nil {
		t.Errorf("FindByID should return error after delete")
	}
	if found != nil {
		t.Errorf("FindByID should nil after delete")
	}
}

func TestCategoryRepository_Delete_NotFound(t *testing.T) {
	repo := NewCategoryRepository()

	err := repo.Delete("nonexistent")

	if err == nil {
		t.Error("Delete should return error for nonexistent ID")
	}

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete() error %v, want ErrNotFound", err)
	}
}
