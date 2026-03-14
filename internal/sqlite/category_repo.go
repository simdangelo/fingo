// internal/sqlite/category_repo.go
package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/simdangelo/fingo/internal/domain"
)

// CategoryRepository implements domain.CategoryRepository using SQLite.
type CategoryRepository struct {
    db *sql.DB
}

// NewCategoryRepository creates a new SQLite category repository.
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
    return &CategoryRepository{
        db: db,
    }
}

// Save inserts or updates a category in the database.
func (r *CategoryRepository) Save(category domain.Category) error {
    query := `
        INSERT OR REPLACE INTO categories (id, name, color)
        VALUES (?, ?, ?)
    `
    
    _, err := r.db.Exec(query, category.ID, category.Name, category.Color)
    if err != nil {
        return fmt.Errorf("failed to save category: %w", err)
    }
    
    return nil
}

// FindByID retrieves a category by its ID.
func (r *CategoryRepository) FindByID(id string) (*domain.Category, error) {
    query := `
        SELECT id, name, color
        FROM categories
        WHERE id = ?
    `
    
    var category domain.Category
    err := r.db.QueryRow(query, id).Scan(&category.ID, &category.Name, &category.Color)
    
    if err == sql.ErrNoRows {
        return nil, domain.ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("failed to find category: %w", err)
    }
    
    return &category, nil
}

// FindAll retrieves all categories.
func (r *CategoryRepository) FindAll() ([]domain.Category, error) {
    query := `
        SELECT id, name, color
        FROM categories
        ORDER BY name
    `
    
    rows, err := r.db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to query categories: %w", err)
    }
    defer rows.Close()
    
    var categories []domain.Category
    
    for rows.Next() {
        var cat domain.Category
        if err := rows.Scan(&cat.ID, &cat.Name, &cat.Color); err != nil {
            return nil, fmt.Errorf("failed to scan category: %w", err)
        }
        categories = append(categories, cat)
    }
    
    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("row iteration error: %w", err)
    }
    
    return categories, nil
}

// Delete removes a category by ID.
func (r *CategoryRepository) Delete(id string) error {
    query := `DELETE FROM categories WHERE id = ?`
    
    result, err := r.db.Exec(query, id)
    if err != nil {
        return fmt.Errorf("failed to delete category: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to check rows affected: %w", err)
    }
    
    if rowsAffected == 0 {
        return domain.ErrNotFound
    }
    
    return nil
}