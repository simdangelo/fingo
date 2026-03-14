package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/simdangelo/fingo/internal/app"
	"github.com/simdangelo/fingo/internal/memory"
	"github.com/simdangelo/fingo/internal/tui"

	// Replace these with your real repository implementations.
	// For now we pass nil repos so the app compiles; swap in real ones when ready.
	_ "github.com/simdangelo/fingo/internal/domain"
)

func main() {
	// Wire up services.
	txRepo := memory.NewTransactionRepository()
	catRepo := memory.NewCategoryRepository()
	txService := app.NewTransactionService(txRepo, catRepo)

	p := tea.NewProgram(
		tui.New(txService),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}