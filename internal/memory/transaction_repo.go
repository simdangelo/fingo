package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/simdangelo/fingo/internal/domain"
)

// TransactionRepository is an in-memory implementation of domain.TransactionRepository.
// It stores transactions in a map and is safe for concurrent use.
type TransactionRepository struct {
	transactions map[string]domain.Transaction
	mu           sync.RWMutex
}

// NewTransactionRepository creates a new in-memory transaction repository.
func NewTransactionRepository() *TransactionRepository {
	return &TransactionRepository{
		transactions: make(map[string]domain.Transaction),
	}
}

// Save stores a transaction in memory
// If a transaction with the same ID exists, it will be overwritten.
func (r *TransactionRepository) Save(transaction domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.transactions[transaction.ID] = transaction
	return nil
}

// FindByID retrieves a transaction by its ID.
// Returns domain.ErrNotFound if the transaction doesn't exist.
func (r *TransactionRepository) FindByID(id string) (*domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transaction, exists := r.transactions[id]
	if !exists {
		return nil, fmt.Errorf("transaction %s, %w", id, domain.ErrNotFound)
	}

	return &transaction, nil
}

func (r *TransactionRepository) FindAll() ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transactions := make([]domain.Transaction, 0, len(r.transactions))

	for _, tx := range r.transactions {
		transactions = append(transactions, tx)
	}

	// Sort by date, newest first
	// We'll use a simple bubble sort for now (good enough for in-memory)
	for i := 0; i < len(transactions); i++ {
		for j := i + 1; j < len(transactions); j++ {
			if transactions[i].Date.Before(transactions[j].Date) {
				transactions[i], transactions[j] = transactions[j], transactions[i]
			}
		}
	}

	return transactions, nil
}

// FindByDateRange retrieves transactions within a date range (inclusive).
// Both start and end are inclusive.
func (r *TransactionRepository) FindByDateRange(start time.Time, end time.Time) ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Transaction, 0)

	for _, tx := range r.transactions {
		if !tx.Date.Before(start) && !tx.Date.After(end) {
			result = append(result, tx)
		}
	}
	// Sort by date, newest first
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Date.Before(result[j].Date) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

// FindByCategory retrieves all transactions for a specific category.
func (r *TransactionRepository) FindByCategory(categoryID string) ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Transaction

	for _, tx := range r.transactions {
		if tx.CategoryID == categoryID {
			result = append(result, tx)
		}
	}

	// Sort by date, newest first
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Date.Before(result[j].Date) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

func (r *TransactionRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.transactions[id]
	if !exists {
		return fmt.Errorf("transaction %s: %w", id, domain.ErrNotFound)
	}

	delete(r.transactions, id)
	return nil
}
