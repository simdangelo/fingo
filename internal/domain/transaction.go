package domain

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)


type TransactionType string

const (
	TypeExpense TransactionType = "expense"
	TypeIncome TransactionType = "income"
)

type Transaction struct {
	ID string
	Date time.Time
	Amount float64
	Type TransactionType
	CategoryID string
	Description string
	CreatedAt time.Time
}

func IsValid(t TransactionType) bool {
	return t == TypeExpense || t == TypeIncome
}

func GenerateTransactionID() string {
	return fmt.Sprintf("tx_%d_%d", time.Now().UnixNano(), rand.Intn(1000))
}

func NewTransaction(date time.Time, amount float64, txType TransactionType, categoryID string, description string) (*Transaction, error) {
	if amount <=0 {
		return nil, errors.New("amount must be greater than zero")
	}

	if !IsValid(txType) {
		return nil, fmt.Errorf("invalid transaction type: %s", txType)
	}

	categoryID = strings.TrimSpace(categoryID)
	if categoryID == "" {
		return nil, errors.New("category ID cannot be empty")
	}

	description = strings.TrimSpace(description)
	if len(description) > 200 {
		return nil, errors.New("description too long (max 200 characters)")
	}

	// Date shouldn't be too far in the future
    // (Allow today/tomorrow, but not years ahead - likely data entry error)
	if date.After(time.Now().AddDate(0, 0, 1)) {
		return nil, errors.New("transaction date cannot be more than 1 day in the future")
	}

	return &Transaction{
		ID: GenerateTransactionID(),
		Date: date,
		Amount: amount,
		Type: txType,
		CategoryID: categoryID,
		Description: description,
		CreatedAt: time.Now(),
	}, nil
}