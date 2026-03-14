package app

import "time"

type MonthlySummary struct {
	Year             int
	Month            time.Month
	TotalIncome      float64
	TotalExpense     float64
	NetAmount        float64
	TransactionCount int
}
