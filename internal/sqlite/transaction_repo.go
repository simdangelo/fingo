// internal/sqlite/transaction_repo.go
package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/simdangelo/fingo/internal/domain"
)

// TransactionRepository implements domain.TransactionRepository using SQLite.
type TransactionRepository struct {
    db *sql.DB
}

// NewTransactionRepository creates a new SQLite transaction repository.
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
    return &TransactionRepository{
        db: db,
    }
}

// formatTime converts Go time.Time to SQLite timestamp string.
func formatTime(t time.Time) string {
    return t.Format(time.RFC3339)
}

// parseTime converts SQLite timestamp string to Go time.Time.
func parseTime(s string) (time.Time, error) {
    return time.Parse(time.RFC3339, s)
}

// Save inserts or updates a transaction in the database.
func (r *TransactionRepository) Save(transaction domain.Transaction) error {
    query := `
        INSERT OR REPLACE INTO transactions 
        (id, date, amount, type, category_id, description, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `
    
    _, err := r.db.Exec(
        query,
        transaction.ID,
        formatTime(transaction.Date),
        transaction.Amount,
        string(transaction.Type),
        transaction.CategoryID,
        transaction.Description,
        formatTime(transaction.CreatedAt),
    )
    
    if err != nil {
        return fmt.Errorf("failed to save transaction: %w", err)
    }
    
    return nil
}

// FindByID retrieves a transaction by its ID.
func (r *TransactionRepository) FindByID(id string) (*domain.Transaction, error) {
    query := `
        SELECT id, date, amount, type, category_id, description, created_at
        FROM transactions
        WHERE id = ?
    `
    
    var tx domain.Transaction
    var dateStr, createdAtStr, typeStr string
    
    err := r.db.QueryRow(query, id).Scan(
        &tx.ID,
        &dateStr,
        &tx.Amount,
        &typeStr,
        &tx.CategoryID,
        &tx.Description,
        &createdAtStr,
    )
    
    if err == sql.ErrNoRows {
        return nil, domain.ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("failed to find transaction: %w", err)
    }
    
    // Parse timestamps
    tx.Date, err = parseTime(dateStr)
    if err != nil {
        return nil, fmt.Errorf("failed to parse date: %w", err)
    }
    
    tx.CreatedAt, err = parseTime(createdAtStr)
    if err != nil {
        return nil, fmt.Errorf("failed to parse created_at: %w", err)
    }
    
    // Convert type
    tx.Type = domain.TransactionType(typeStr)
    
    return &tx, nil
}

// FindAll retrieves all transactions, ordered by date (newest first).
func (r *TransactionRepository) FindAll() ([]domain.Transaction, error) {
    query := `
        SELECT id, date, amount, type, category_id, description, created_at
        FROM transactions
        ORDER BY date DESC
    `
    
    rows, err := r.db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to query transactions: %w", err)
    }
    defer rows.Close()
    
    return r.scanTransactions(rows)
}

// FindByDateRange retrieves transactions within a date range (inclusive).
func (r *TransactionRepository) FindByDateRange(start, end time.Time) ([]domain.Transaction, error) {
    query := `
        SELECT id, date, amount, type, category_id, description, created_at
        FROM transactions
        WHERE date >= ? AND date <= ?
        ORDER BY date DESC
    `
    
    rows, err := r.db.Query(query, formatTime(start), formatTime(end))
    if err != nil {
        return nil, fmt.Errorf("failed to query transactions by date range: %w", err)
    }
    defer rows.Close()
    
    return r.scanTransactions(rows)
}

// FindByCategory retrieves all transactions for a specific category.
func (r *TransactionRepository) FindByCategory(categoryID string) ([]domain.Transaction, error) {
    query := `
        SELECT id, date, amount, type, category_id, description, created_at
        FROM transactions
        WHERE category_id = ?
        ORDER BY date DESC
    `
    
    rows, err := r.db.Query(query, categoryID)
    if err != nil {
        return nil, fmt.Errorf("failed to query transactions by category: %w", err)
    }
    defer rows.Close()
    
    return r.scanTransactions(rows)
}

// Delete removes a transaction by ID.
func (r *TransactionRepository) Delete(id string) error {
    query := `DELETE FROM transactions WHERE id = ?`
    
    result, err := r.db.Exec(query, id)
    if err != nil {
        return fmt.Errorf("failed to delete transaction: %w", err)
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

// scanTransactions is a helper that scans multiple transaction rows.
func (r *TransactionRepository) scanTransactions(rows *sql.Rows) ([]domain.Transaction, error) {
    var transactions []domain.Transaction
    
    for rows.Next() {
        var tx domain.Transaction
        var dateStr, createdAtStr, typeStr string
        
        err := rows.Scan(
            &tx.ID,
            &dateStr,
            &tx.Amount,
            &typeStr,
            &tx.CategoryID,
            &tx.Description,
            &createdAtStr,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan transaction: %w", err)
        }
        
        // Parse timestamps
        tx.Date, err = parseTime(dateStr)
        if err != nil {
            return nil, fmt.Errorf("failed to parse date: %w", err)
        }
        
        tx.CreatedAt, err = parseTime(createdAtStr)
        if err != nil {
            return nil, fmt.Errorf("failed to parse created_at: %w", err)
        }
        
        // Convert type
        tx.Type = domain.TransactionType(typeStr)
        
        transactions = append(transactions, tx)
    }
    
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("row iteration error: %w", err)
    }
    
    return transactions, nil
}