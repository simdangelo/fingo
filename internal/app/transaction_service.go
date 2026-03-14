package app

import (
	"fmt"
	"time"

	"github.com/simdangelo/fingo/internal/domain"
)

// TransactionService provides business operations for transactions.
// It coordinates between domain logic and data persistence.
type TransactionService struct {
	transactionRepo domain.TransactionRepository
	categoryRepo    domain.CategoryRepository
}

// NewTransactionService creates a new transaction service with the given repositories.
func NewTransactionService(transanctionRepo domain.TransactionRepository, categoryRepo domain.CategoryRepository) *TransactionService {
	return &TransactionService{
		transactionRepo: transanctionRepo,
		categoryRepo:    categoryRepo,
	}
}

// AddTransaction creates and saves a new transaction.
// It validates that the category exists before creating the transaction.
// Returns ErrCategoryNotFound if the specified category doesn't exist.
func (s *TransactionService) AddTransaction(
	date time.Time,
	amount float64,
	txType domain.TransactionType,
	categoryID string,
	description string,
) error {
	// Step 1: Validate categoryQ exists
	_, err := s.categoryRepo.FindByID(categoryID)
	if err != nil {
		return fmt.Errorf("%w, %s", ErrCategoryNotFound, categoryID)
	}

	// Step 2: create transaction
	tx, err := domain.NewTransaction(date, amount, txType, categoryID, description)
	if err != nil {
		return err
	}

	// Step 3: save to repository
	err = s.transactionRepo.Save(*tx)
	if err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}
	return nil
}

// GetTransaction retrieves a transaction by ID.
// Returns ErrTransactionNotFound if the transaction doesn't exist.
func (s *TransactionService) GetTransaction(id string) (*domain.Transaction, error) {
	tx, err := s.transactionRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTransactionNotFound, id)
	}
	return tx, nil
}

// GetAllTransactions retrieves all transactions, ordered by date (newest first).
func (s *TransactionService) GetAllTransactions() ([]domain.Transaction, error) {
	transactions, err := s.transactionRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transactions: %w", err)
	}

	return transactions, nil
}

func (s *TransactionService) GetTransactionByMonth(year int, month time.Month) ([]domain.Transaction, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	nextMonth := start.AddDate(0, 1, 0)
	end := nextMonth.Add(-1)

	transactions, err := s.transactionRepo.FindByDateRange(start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction for %s %d: %w", month, year, err)
	}
	return transactions, nil
}

func (s *TransactionService) GetTransactionByDateRange(start time.Time, end time.Time) ([]domain.Transaction, error) {
	transactions, err := s.transactionRepo.FindByDateRange(start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions between %s and %s: %w", start.Format("2006-01-02"), end.Format("2006-01-02"), err)
	}
	return transactions, nil
}

// GetCategoryTransactions retrieves all transactions for a specific category.
// Returns an empty slice if the category has no transactions.
// Returns ErrCategoryNotFound if the category doesn't exist.
func (s *TransactionService) GetCategoryTransactions(categoryID string) ([]domain.Transaction, error) {
	_, err := s.categoryRepo.FindByID(categoryID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCategoryNotFound, categoryID)
	}

	transactions, err := s.transactionRepo.FindByCategory(categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions for category %s: %w", categoryID, err)
	}
	return transactions, nil
}

// CalculateCategoryTotal calculates the total amount for a category.
// For expense categories, returns negative total (money spent).
// For income categories, returns positive total (money earned).
// Returns zero if category has no transactions.
func (s *TransactionService) CalculateCategoryTotal(categoryID string) (float64, error) {
	transactions, err := s.GetCategoryTransactions(categoryID)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, tx := range transactions {
		if tx.Type == domain.TypeIncome {
			total += tx.Amount
		} else {
			total -= tx.Amount
		}
	}

	return total, nil
}

// GetMonthlySummary calculates financial summary for a specific month.
// Returns total income, expenses, net amount, and transaction count.
func (s *TransactionService) GetMonthlySummary(year int, month time.Month) (*MonthlySummary, error) {
	transactions, err := s.GetTransactionByMonth(year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly summary: %w", err)
	}

	summary := &MonthlySummary{
		Year:  year,
		Month: month,
	}

	for _, tx := range transactions {
		summary.TransactionCount++
		if tx.Type == domain.TypeIncome {
			summary.TotalIncome += tx.Amount
		} else {
			summary.TotalExpense += tx.Amount
		}
	}

	summary.NetAmount = summary.TotalIncome - summary.TotalExpense
	return summary, nil
}
