package main

import (
	"fmt"
	"log"
	"time"

	"github.com/simdangelo/fingo/internal/app"
	"github.com/simdangelo/fingo/internal/domain"
	"github.com/simdangelo/fingo/internal/sqlite"
)

func main() {
    // Initialize database
    db, err := sqlite.InitDB("finance.db")
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer sqlite.Close(db)
    
    fmt.Println("✅ Database initialized: finance.db")
    
    // Create repositories
    categoryRepo := sqlite.NewCategoryRepository(db)
    transactionRepo := sqlite.NewTransactionRepository(db)
    
    // Create services
    categoryService := app.NewCategoryService(categoryRepo, transactionRepo)
    transactionService := app.NewTransactionService(transactionRepo, categoryRepo)
    
    fmt.Println("✅ Services initialized")
    
    // Run demo workflow
    if err := runDemo(categoryService, transactionService); err != nil {
        log.Fatalf("Demo failed: %v", err)
    }
    
    fmt.Println("\n✅ Demo completed successfully!")
    fmt.Println("📁 Data saved to finance.db")
    fmt.Println("💡 Run the program again to see data persistence!")
}

func runDemo(catService *app.CategoryService, txService *app.TransactionService) error {
    fmt.Println("\n--- Running Finance App Demo ---\n")
    
    // Check if we already have categories (data from previous run)
    existingCategories, err := catService.ListCategories()
    if err != nil {
        return fmt.Errorf("failed to list categories: %w", err)
    }
    
    if len(existingCategories) > 0 {
        fmt.Printf("📂 Found %d existing categories from previous run\n", len(existingCategories))
        return displayExistingData(catService, txService)
    }
    
    // First run - create sample data
    return createSampleData(catService, txService)
}

func createSampleData(catService *app.CategoryService, txService *app.TransactionService) error {
    fmt.Println("🆕 Creating sample data...")
    
    // Step 1: Create categories
    fmt.Println("\n1️⃣  Creating categories...")
    
    groceries, err := catService.CreateCategory("Groceries", "#FF5733")
    if err != nil {
        return err
    }
    fmt.Printf("   ✓ Created: %s (%s)\n", groceries.Name, groceries.Color)
    
    rent, err := catService.CreateCategory("Rent", "#3498DB")
    if err != nil {
        return err
    }
    fmt.Printf("   ✓ Created: %s (%s)\n", rent.Name, rent.Color)
    
    salary, err := catService.CreateCategory("Salary", "#2ECC71")
    if err != nil {
        return err
    }
    fmt.Printf("   ✓ Created: %s (%s)\n", salary.Name, salary.Color)
    
    transport, err := catService.CreateCategory("Transport", "#F39C12")
    if err != nil {
        return err
    }
    fmt.Printf("   ✓ Created: %s (%s)\n", transport.Name, transport.Color)
    
    // Step 2: Add transactions
    fmt.Println("\n2️⃣  Adding transactions...")
    
    now := time.Now()
    currentMonth := now.Month()
    currentYear := now.Year()
    
    // Add some expenses
    transactions := []struct {
        date        time.Time
        amount      float64
        txType      domain.TransactionType
        categoryID  string
        description string
    }{
        {time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.UTC), 1200.00, domain.TypeExpense, rent.ID, "Monthly rent"},
        {time.Date(currentYear, currentMonth, 5, 0, 0, 0, 0, time.UTC), 85.50, domain.TypeExpense, groceries.ID, "Weekly groceries"},
        {time.Date(currentYear, currentMonth, 4, 0, 0, 0, 0, time.UTC), 120.75, domain.TypeExpense, groceries.ID, "Grocery shopping"},
        {time.Date(currentYear, currentMonth, 8, 0, 0, 0, 0, time.UTC), 45.00, domain.TypeExpense, transport.ID, "Gas"},
        {time.Date(currentYear, currentMonth, 10, 0, 0, 0, 0, time.UTC), 3500.00, domain.TypeIncome, salary.ID, "Monthly salary"},
        {time.Date(currentYear, currentMonth, 2, 0, 0, 0, 0, time.UTC), 95.25, domain.TypeExpense, groceries.ID, "Food shopping"},
        {time.Date(currentYear, currentMonth, 7, 0, 0, 0, 0, time.UTC), 60.00, domain.TypeExpense, transport.ID, "Public transport pass"},
    }
    
    for _, tx := range transactions {
        err := txService.AddTransaction(
            tx.date,
            tx.amount,
            tx.txType,
            tx.categoryID,
            tx.description,
        )
        if err != nil {
            return err
        }
        
        typeIcon := "💸"
        if tx.txType == domain.TypeIncome {
            typeIcon = "💰"
        }
        fmt.Printf("   %s $%.2f - %s\n", typeIcon, tx.amount, tx.description)
    }
    
    // Step 3: Display summary
    return displaySummary(catService, txService, currentYear, currentMonth)
}

func displayExistingData(catService *app.CategoryService, txService *app.TransactionService) error {
    fmt.Println("\n📊 Displaying existing data...")
    
    now := time.Now()
    return displaySummary(catService, txService, now.Year(), now.Month())
}

func displaySummary(catService *app.CategoryService, txService *app.TransactionService, year int, month time.Month) error {
    fmt.Printf("\n3️⃣  Monthly Summary for %s %d\n", month.String(), year)
    
    // Get monthly summary
    summary, err := txService.GetMonthlySummary(year, month)
    if err != nil {
        return err
    }
    
    fmt.Println("\n┌─────────────────────────────────────┐")
    fmt.Printf("│ Income:      $%18.2f │\n", summary.TotalIncome)
    fmt.Printf("│ Expenses:    $%18.2f │\n", summary.TotalExpense)
    fmt.Println("│─────────────────────────────────────│")
    fmt.Printf("│ Net:         $%18.2f │\n", summary.NetAmount)
    fmt.Printf("│ Transactions: %17d │\n", summary.TransactionCount)
    fmt.Println("└─────────────────────────────────────┘")
    
    // Show category breakdown
    fmt.Println("\n4️⃣  Spending by Category")
    
    categories, err := catService.ListCategories()
    if err != nil {
        return err
    }
    
    for _, cat := range categories {
        total, err := txService.CalculateCategoryTotal(cat.ID)
        if err != nil {
            return err
        }
        
        // Only show categories with transactions
        if total != 0 {
            icon := "💸"
            if total > 0 {
                icon = "💰"
            }
            fmt.Printf("   %s %-15s $%.2f\n", icon, cat.Name, total)
        }
    }
    
    // Show recent transactions
    fmt.Println("\n5️⃣  Recent Transactions")
    
    allTxs, err := txService.GetAllTransactions()
    if err != nil {
        return err
    }
    
    // Show up to 5 most recent
    count := len(allTxs)
    if count > 5 {
        count = 5
    }
    
    for i := 0; i < count; i++ {
        tx := allTxs[i]
        
        icon := "💸"
        if tx.Type == domain.TypeIncome {
            icon = "💰"
        }
        
        fmt.Printf("   %s %s - $%.2f - %s\n",
            icon,
            tx.Date.Format("Jan 02"),
            tx.Amount,
            tx.Description,
        )
    }
    
    if len(allTxs) > 5 {
        fmt.Printf("   ... and %d more transactions\n", len(allTxs)-5)
    }
    
    return nil
}