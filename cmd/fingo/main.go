package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/simdangelo/fingo/internal/app"
	"github.com/simdangelo/fingo/internal/sqlite"
	"github.com/simdangelo/fingo/internal/tui"

	// Replace these with your real repository implementations.
	// For now we pass nil repos so the app compiles; swap in real ones when ready.
	_ "github.com/simdangelo/fingo/internal/domain"
)

func main() {
	// Wire up services.
	// txRepo := memory.NewTransactionRepository()
	// catRepo := memory.NewCategoryRepository()
	// txService := app.NewTransactionService(txRepo, catRepo)
	// catService := app.NewCategoryService(catRepo, txRepo)
	db, err := sqlite.InitDB("finance.db")
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer sqlite.Close(db)
    
    // Create repositories
    categoryRepo := sqlite.NewCategoryRepository(db)
    transactionRepo := sqlite.NewTransactionRepository(db)
    
    // Create services
    catService := app.NewCategoryService(categoryRepo, transactionRepo)
    txService := app.NewTransactionService(transactionRepo, categoryRepo)
  

	p := tea.NewProgram(
		tui.New(txService, catService),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}