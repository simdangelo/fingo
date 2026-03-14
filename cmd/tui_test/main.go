package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/simdangelo/fingo/internal/app"
	"github.com/simdangelo/fingo/internal/sqlite"
	"github.com/simdangelo/fingo/internal/tui"
)

func main() {
    // Initialize database
    db, err := sqlite.InitDB("finance.db")
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer sqlite.Close(db)
    
    // Create repositories
    categoryRepo := sqlite.NewCategoryRepository(db)
    transactionRepo := sqlite.NewTransactionRepository(db)
    
    // Create services
    categoryService := app.NewCategoryService(categoryRepo, transactionRepo)
    transactionService := app.NewTransactionService(transactionRepo, categoryRepo)
    
    // Create TUI model
    m := tui.New(categoryService, transactionService)
    
    // Run the program
    p := tea.NewProgram(m, tea.WithAltScreen())
    
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error running program: %v\n", err)
        os.Exit(1)
    }
}