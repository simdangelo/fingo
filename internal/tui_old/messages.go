package tui

import (
	"github.com/simdangelo/fingo/internal/app"
	"github.com/simdangelo/fingo/internal/domain"
)

// Navigation messages
type changeScreenMsg Screen

// Data loading messages
type categoriesLoadedMsg struct {
    categories []domain.Category
    err        error
}

type transactionsLoadedMsg struct {
    transactions []domain.Transaction
    err          error
}

type summaryLoadedMsg struct {
    summary *app.MonthlySummary
    err     error
}

// Action result messages
type transactionAddedMsg struct {
    err error
}

type categoryCreatedMsg struct {
    category *domain.Category
    err      error
}

type categoryDeletedMsg struct {
    err error
}

// UI messages
type errorMsg struct {
    err error
}

type successMsg struct {
    message string
}