package domain

import "time"

type CategoryRepository interface {
	Save(category Category) error
	FindByID(id string) (*Category, error)
	FindAll() ([]Category, error)
	Delete(id string) error
}

type TransactionRepository interface {
	Save(transaction Transaction) error
	FindByID(id string) (*Transaction, error)
	FindAll() ([]Transaction, error)
	FindByDateRange(start time.Time, end time.Time) ([]Transaction, error)
	FindByCategory(categoryID string) ([]Transaction, error)
	Delete(id string) error
}
