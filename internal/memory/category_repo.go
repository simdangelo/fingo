package memory

import (
	"fmt"
	"sync"

	"github.com/simdangelo/fingo/internal/domain"
)

// CategoryRepository is an in-memory implementation of domain.CategoryRepository.
// It stores categories in a map and is safe for concurrent use.
type CategoryRepository struct {
	categories map[string]domain.Category
	mu         sync.RWMutex
}

// NewCategoryRepository creates a new in-memory category repository.
func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{
		categories: make(map[string]domain.Category),
	}
}

// Save stores a category in memory.
// If a category with the same ID exists, it will be overwritten.
func (r *CategoryRepository) Save(category domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.categories[category.ID] = category
	return nil
}

// FindByID retrieves a category by its ID.
// Returns domain.ErrNotFound if the category doesn't exist.
func (r *CategoryRepository) FindByID(id string) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	category, exists := r.categories[id]
	if !exists {
		return nil, fmt.Errorf("category %s: %w", id, domain.ErrNotFound)
	}

	return &category, nil
}

func (r *CategoryRepository) FindAll() ([]domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create slice with exact capacity needed
	categories := make([]domain.Category, 0, len(r.categories))
	for _, category := range r.categories {
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *CategoryRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.categories[id]
	if !exists {
		return fmt.Errorf("category %s: %w", id, domain.ErrNotFound)
	}

	delete(r.categories, id)
	return nil
}
