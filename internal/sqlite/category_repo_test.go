package sqlite

import (
	"testing"

	"github.com/simdangelo/fingo/internal/domain"
)

func setupCategoryRepoTest(t *testing.T) (*CategoryRepository, func()) {
    t.Helper()
    
    db, err := OpenInMemory()
    if err != nil {
        t.Fatalf("failed to open in-memory db: %v", err)
    }
    
    if err := Migrate(db); err != nil {
        db.Close()
        t.Fatalf("failed to migrate: %v", err)
    }
    
    repo := NewCategoryRepository(db)
    
    cleanup := func() {
        db.Close()
    }
    
    return repo, cleanup
}

func TestCategoryRepository_Save_And_FindByID(t *testing.T) {
    repo, cleanup := setupCategoryRepoTest(t)
    defer cleanup()
    
    // Create a category
    cat, err := domain.NewCategory("Groceries", "#FF5733")
    if err != nil {
        t.Fatalf("failed to create category: %v", err)
    }
    
    // Save it
    if err := repo.Save(*cat); err != nil {
        t.Fatalf("Save() error = %v", err)
    }
    
    // Retrieve it
    found, err := repo.FindByID(cat.ID)
    if err != nil {
        t.Fatalf("FindByID() error = %v", err)
    }
    
    // Verify fields match
    if found.ID != cat.ID {
        t.Errorf("ID = %v, want %v", found.ID, cat.ID)
    }
    if found.Name != cat.Name {
        t.Errorf("Name = %v, want %v", found.Name, cat.Name)
    }
    if found.Color != cat.Color {
        t.Errorf("Color = %v, want %v", found.Color, cat.Color)
    }
}

func TestCategoryRepository_FindByID_NotFound(t *testing.T) {
    repo, cleanup := setupCategoryRepoTest(t)
    defer cleanup()
    
    _, err := repo.FindByID("nonexistent")
    
    if err != domain.ErrNotFound {
        t.Errorf("FindByID() error = %v, want ErrNotFound", err)
    }
}

func TestCategoryRepository_Save_Update(t *testing.T) {
    repo, cleanup := setupCategoryRepoTest(t)
    defer cleanup()
    
    // Create and save
    cat, _ := domain.NewCategory("Groceries", "#FF5733")
    repo.Save(*cat)
    
    // Modify and save again
    cat.Name = "Food"
    cat.Color = "#00FF00"
    
    if err := repo.Save(*cat); err != nil {
        t.Fatalf("Save() update error = %v", err)
    }
    
    // Retrieve and verify
    found, _ := repo.FindByID(cat.ID)
    
    if found.Name != "Food" {
        t.Errorf("Name after update = %v, want Food", found.Name)
    }
    if found.Color != "#00FF00" {
        t.Errorf("Color after update = %v, want #00FF00", found.Color)
    }
}

func TestCategoryRepository_FindAll(t *testing.T) {
    repo, cleanup := setupCategoryRepoTest(t)
    defer cleanup()
    
    // Create multiple categories
    cat1, _ := domain.NewCategory("Groceries", "#FF5733")
    cat2, _ := domain.NewCategory("Rent", "#00FF00")
    cat3, _ := domain.NewCategory("Transport", "#0000FF")
    
    repo.Save(*cat1)
    repo.Save(*cat2)
    repo.Save(*cat3)
    
    // Retrieve all
    all, err := repo.FindAll()
    if err != nil {
        t.Fatalf("FindAll() error = %v", err)
    }
    
    if len(all) != 3 {
        t.Errorf("FindAll() returned %d categories, want 3", len(all))
    }
    
    // Verify all IDs are present
    ids := make(map[string]bool)
    for _, cat := range all {
        ids[cat.ID] = true
    }
    
    if !ids[cat1.ID] || !ids[cat2.ID] || !ids[cat3.ID] {
        t.Error("FindAll() didn't return all categories")
    }
}

func TestCategoryRepository_FindAll_Empty(t *testing.T) {
    repo, cleanup := setupCategoryRepoTest(t)
    defer cleanup()
    
    all, err := repo.FindAll()
    if err != nil {
        t.Fatalf("FindAll() error = %v", err)
    }
    
    if len(all) != 0 {
        t.Errorf("FindAll() on empty db returned %d categories, want 0", len(all))
    }
}

func TestCategoryRepository_FindAll_Ordering(t *testing.T) {
    repo, cleanup := setupCategoryRepoTest(t)
    defer cleanup()
    
    // Create categories in non-alphabetical order
    catZ, _ := domain.NewCategory("Zebra", "#FF5733")
    catA, _ := domain.NewCategory("Apple", "#00FF00")
    catM, _ := domain.NewCategory("Mango", "#0000FF")
    
    repo.Save(*catZ)
    repo.Save(*catA)
    repo.Save(*catM)
    
    all, _ := repo.FindAll()
    
    // Should be ordered alphabetically by name
    if all[0].Name != "Apple" {
        t.Errorf("First category = %v, want Apple", all[0].Name)
    }
    if all[1].Name != "Mango" {
        t.Errorf("Second category = %v, want Mango", all[1].Name)
    }
    if all[2].Name != "Zebra" {
        t.Errorf("Third category = %v, want Zebra", all[2].Name)
    }
}

func TestCategoryRepository_Delete(t *testing.T) {
    repo, cleanup := setupCategoryRepoTest(t)
    defer cleanup()
    
    // Create and save
    cat, _ := domain.NewCategory("Groceries", "#FF5733")
    repo.Save(*cat)
    
    // Delete
    if err := repo.Delete(cat.ID); err != nil {
        t.Fatalf("Delete() error = %v", err)
    }
    
    // Verify it's gone
    _, err := repo.FindByID(cat.ID)
    if err != domain.ErrNotFound {
        t.Errorf("FindByID() after delete error = %v, want ErrNotFound", err)
    }
}

func TestCategoryRepository_Delete_NotFound(t *testing.T) {
    repo, cleanup := setupCategoryRepoTest(t)
    defer cleanup()
    
    err := repo.Delete("nonexistent")
    
    if err != domain.ErrNotFound {
        t.Errorf("Delete() error = %v, want ErrNotFound", err)
    }
}