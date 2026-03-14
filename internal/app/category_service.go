package app

import (
	"fmt"

	"github.com/simdangelo/fingo/internal/domain"
)

type CategoryService struct {
	categoryRepo    domain.CategoryRepository
	transactionRepo domain.TransactionRepository
}

// NewCategoryService creates a new category service.
func NewCategoryService(categoryRepo domain.CategoryRepository, transactionRepo domain.TransactionRepository) *CategoryService {
	return &CategoryService{
		categoryRepo:    categoryRepo,
		transactionRepo: transactionRepo,
	}
}

// CreateCategory creates and saves a new category.
func (s *CategoryService) CreateCategory(name, color string) (*domain.Category, error) {
	category, err := domain.NewCategory(name, color)
	if err != nil {
		return nil, err
	}

	err = s.categoryRepo.Save(*category)
	if err != nil {
		return nil, fmt.Errorf("failed to save category: %w", err)
	}

	return category, nil
}

// GetCategory retrieves a category by ID.
func (s *CategoryService) GetCategory(id string) (*domain.Category, error) {
	category, err := s.categoryRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCategoryNotFound, id)
	}

	return category, nil
}

// ListCategories retrieves all categories.
func (s *CategoryService) ListCategories() ([]domain.Category, error) {
	categories, err := s.categoryRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	return categories, nil
}

// CanDeleteCategory checks if a category can be safely deleted.
// Returns true if the category exists and has no associated transactions.
func (s *CategoryService) CanDeleteCategory(id string) (bool, error) {
	// Check if category exists
	_, err := s.categoryRepo.FindByID(id)
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrCategoryNotFound, id)
	}

	// Check if any transactions use this category
	transactions, err := s.transactionRepo.FindByCategory(id)
	if err != nil {
		return false, fmt.Errorf("failed to check category usage: %w", err)
	}
	return len(transactions) == 0, nil
}

// DeleteCategory removes a category if it's not in use.
// Returns ErrCategoryInUse if transactions reference this category.
func (s *CategoryService) DeleteCategory(id string) error {
	canDelete, err := s.CanDeleteCategory(id)
	if err != nil {
		return err
	}

	if !canDelete {
		return ErrCategoryInUse
	}

	err = s.categoryRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}
