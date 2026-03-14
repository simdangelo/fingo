package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/simdangelo/fingo/internal/app"
	"github.com/simdangelo/fingo/internal/domain"
)

// Screen represents different views in the application
type Screen int

const (
    ScreenDashboard Screen = iota
    ScreenTransactions
    ScreenAddTransaction
    ScreenCategories
)

// Model is the main TUI model
type Model struct {
    // Services
    categoryService    *app.CategoryService
    transactionService *app.TransactionService
    
    // Current state
    currentScreen Screen
    
    // Cached data
    categories   []domain.Category
    transactions []domain.Transaction
    summary      *app.MonthlySummary
    
    // UI state
    width  int
    height int
    
    // Messages
    err     error
    success string
    
    // Loading states
    loading bool
    
    // Screen-specific models
    transactionsModel TransactionsModel  // Add this
}

// New creates a new TUI model with services
func New(
    categoryService *app.CategoryService,
    transactionService *app.TransactionService,
) Model {
    return Model{
        categoryService:    categoryService,
        transactionService: transactionService,
        currentScreen:      ScreenDashboard,
    }
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
    // Load initial data
    return tea.Batch(
        loadCategoriesCmd(m.categoryService),
        loadTransactionsCmd(m.transactionService),
        loadSummaryCmd(m.transactionService),
    )
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    
    // Window size
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        return m, nil
    
    // Keyboard input
    case tea.KeyMsg:
        // Handle screen-specific keys first
        if m.currentScreen == ScreenTransactions {
            switch msg.String() {
            case "up", "k":
                if m.transactionsModel.cursor > 0 {
                    m.transactionsModel.cursor--
                }
                return m, nil
                
            case "down", "j":
                if m.transactionsModel.cursor < len(m.transactionsModel.filteredTxs)-1 {
                    m.transactionsModel.cursor++
                }
                return m, nil
                
            case "f":
                // Toggle filter
                if m.transactionsModel.filter == FilterAll {
                    m.transactionsModel.filter = FilterCurrentMonth
                } else {
                    m.transactionsModel.filter = FilterAll
                }
                m.transactionsModel.filteredTxs = m.filterTransactions(m.transactionsModel.filter)
                m.transactionsModel.cursor = 0  // Reset cursor
                return m, nil
            }
        }
        switch msg.String() {
        
        case "ctrl+c", "q":
            return m, tea.Quit
        
        // Screen navigation
        case "1":
            m.currentScreen = ScreenDashboard
            return m, nil
            
        case "2":
            m.currentScreen = ScreenTransactions
            m.transactionsModel = m.initTransactionsModel()  // Initialize
            return m, nil
            
        case "3":
            m.currentScreen = ScreenCategories
            return m, nil
            
        case "a":
            // Add transaction from any screen
            m.currentScreen = ScreenAddTransaction
            return m, nil
        }
    
    // Navigation messages
    case changeScreenMsg:
        m.currentScreen = Screen(msg)
        return m, nil
    
    // Data loaded messages
    case categoriesLoadedMsg:
        m.loading = false
        if msg.err != nil {
            m.err = msg.err
        } else {
            m.categories = msg.categories
        }
        return m, nil
    
    case transactionsLoadedMsg:
        m.loading = false
        if msg.err != nil {
            m.err = msg.err
        } else {
            m.transactions = msg.transactions
        }
        return m, nil
    
    case summaryLoadedMsg:
        m.loading = false
        if msg.err != nil {
            m.err = msg.err
        } else {
            m.summary = msg.summary
        }
        return m, nil
    
    // Action result messages
    case transactionAddedMsg:
        if msg.err != nil {
            m.err = msg.err
        } else {
            m.success = "Transaction added successfully!"
            m.currentScreen = ScreenDashboard
            
            // Reload data
            return m, tea.Batch(
                loadTransactionsCmd(m.transactionService),
                loadSummaryCmd(m.transactionService),
            )
        }
        return m, nil
    
    case categoryCreatedMsg:
        if msg.err != nil {
            m.err = msg.err
        } else {
            m.success = "Category created successfully!"
            
            // Reload categories
            return m, loadCategoriesCmd(m.categoryService)
        }
        return m, nil
    
    // Error and success messages
    case errorMsg:
        m.err = msg.err
        return m, nil
    
    case successMsg:
        m.success = msg.message
        return m, nil
    }
    
    return m, nil
}

// View renders the UI
func (m Model) View() string {
    // Build UI parts
    header := m.renderHeader()
    content := m.renderContent()
    footer := m.renderFooter()
    
    // Combine parts
    return fmt.Sprintf("%s\n%s\n%s", header, content, footer)
}

// renderHeader renders the title bar and tabs
func (m Model) renderHeader() string {
    // Title bar
    title := TitleStyle.Render("💰 Personal Finance")
    quit := HelpStyle.Render("[q] Quit")
    
    // Calculate spacing (handle case where width isn't set yet)
    spacing := ""
    if m.width > 40 {
        // Right-align quit text
        titleLen := 20  // Approximate length of title
        quitLen := 10   // Approximate length of quit
        spaces := m.width - titleLen - quitLen
        if spaces > 0 {
            spacing = repeat(" ", spaces)
        }
    }
    
    titleBar := fmt.Sprintf("%s%s%s", title, spacing, quit)
    
    // Tab bar
    tabs := []string{
        m.renderTab("1", "Dashboard", ScreenDashboard),
        m.renderTab("2", "Transactions", ScreenTransactions),
        m.renderTab("3", "Categories", ScreenCategories),
    }
    tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
    
    return titleBar + "\n" + tabBar
}

// renderTab renders a single tab
func (m Model) renderTab(key, label string, screen Screen) string {
    style := InactiveTabStyle
    if m.currentScreen == screen {
        style = ActiveTabStyle
    }
    return style.Render(fmt.Sprintf("[%s] %s", key, label))
}

// renderContent renders the current screen
func (m Model) renderContent() string {
    // Show errors if any
    if m.err != nil {
        errorMsg := ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
        return ContentStyle.Render(errorMsg)
    }
    
    // Show success messages
    if m.success != "" {
        successMsg := SuccessStyle.Render(m.success)
        m.success = "" // Clear after showing
        return ContentStyle.Render(successMsg)
    }
    
    // Show loading
    if m.loading {
        return ContentStyle.Render("Loading...")
    }
    
    // Render current screen
    switch m.currentScreen {
    case ScreenDashboard:
        return m.renderDashboard()
    case ScreenTransactions:
        return m.renderTransactions()
    case ScreenAddTransaction:
        return m.renderAddTransaction()
    case ScreenCategories:
        return m.renderCategories()
    default:
        return ContentStyle.Render("Unknown screen")
    }
}

// renderFooter renders help text
func (m Model) renderFooter() string {
    var help string
    
    switch m.currentScreen {
    case ScreenDashboard:
        help = "1/2/3: Switch screens  a: Add Transaction  q: Quit"
        
    case ScreenTransactions:
        help = "↑/↓: Navigate  f: Filter  a: Add Transaction  q: Quit"
        
    case ScreenCategories:
        help = "↑/↓: Navigate  a: Add Category  d: Delete  q: Quit"
        
    default:
        help = "Esc: Back  q: Quit"
    }
    
    return HelpStyle.Render(help)
}

func (m Model) renderTransactions() string {
    var s string
    
    // Title
    title := lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorPrimary).
        Render("📝 Transactions")
    
    s += title + "\n\n"
    
    // Filter info
    filterStyle := lipgloss.NewStyle().
        Foreground(ColorInfo).
        Bold(true)
    
    filterText := fmt.Sprintf("Filter: %s (press [f] to change)", 
        filterStyle.Render(m.transactionsModel.filter.String()))
    
    s += lipgloss.NewStyle().
        Foreground(ColorMuted).
        Render(filterText) + "\n\n"
    
    // Get filtered transactions
    txs := m.transactionsModel.filteredTxs
    
    if len(txs) == 0 {
        emptyMsg := "No transactions found"
        if m.transactionsModel.filter == FilterCurrentMonth {
            emptyMsg += " for current month"
        }
        s += lipgloss.NewStyle().
            Foreground(ColorMuted).
            Italic(true).
            Render(emptyMsg)
        return ContentStyle.Render(s)
    }
    
    // Transaction count and total
    var totalIncome, totalExpense float64
    for _, tx := range txs {
        if tx.Type == domain.TypeIncome {
            totalIncome += tx.Amount
        } else {
            totalExpense += tx.Amount
        }
    }
    
    statsLine := fmt.Sprintf("%d transactions | Income: %s | Expenses: %s",
        len(txs),
        lipgloss.NewStyle().Foreground(ColorSuccess).Render(formatCurrency(totalIncome)),
        lipgloss.NewStyle().Foreground(ColorDanger).Render(formatCurrency(totalExpense)),
    )
    
    s += lipgloss.NewStyle().
        Foreground(ColorMuted).
        Render(statsLine) + "\n\n"
    
    // Transaction list
    s += m.renderTransactionList(txs)
    
    return ContentStyle.Render(s)
}

func (m Model) renderTransactionList(txs []domain.Transaction) string {
    var s string
    
    // Calculate visible window (pagination)
    maxVisible := 15  // Show 15 transactions at a time
    cursor := m.transactionsModel.cursor
    
    start := 0
    end := len(txs)
    
    if len(txs) > maxVisible {
        // Center cursor in view
        start = cursor - maxVisible/2
        if start < 0 {
            start = 0
        }
        end = start + maxVisible
        if end > len(txs) {
            end = len(txs)
            start = end - maxVisible
            if start < 0 {
                start = 0
            }
        }
    }
    
    // Render visible transactions
    for i := start; i < end; i++ {
        tx := txs[i]
        
        // Cursor indicator
        cursorStr := "  "
        if i == cursor {
            cursorStr = lipgloss.NewStyle().
                Foreground(ColorPrimary).
                Render("→ ")
        }
        
        // Icon
        icon := "💸"
        if tx.Type == domain.TypeIncome {
            icon = "💰"
        }
        
        // Date
        dateStr := tx.Date.Format("Jan 02, 2006")
        dateStyle := lipgloss.NewStyle().
            Foreground(ColorMuted).
            Width(15)
        
        // Amount
        amountColor := ColorDanger
        if tx.Type == domain.TypeIncome {
            amountColor = ColorSuccess
        }
        amountStr := formatCurrency(tx.Amount)
        amountStyle := lipgloss.NewStyle().
            Foreground(amountColor).
            Bold(true).
            Width(12).
            Align(lipgloss.Right)
        
        // Category name
        catName := m.getCategoryName(tx.CategoryID)
        catStyle := lipgloss.NewStyle().
            Foreground(ColorInfo).
            Width(15)
        
        // Description
        desc := tx.Description
        if len(desc) > 30 {
            desc = desc[:27] + "..."
        }
        
        // Highlight selected row
        line := fmt.Sprintf("%s%s %s %s %s %s",
            cursorStr,
            icon,
            dateStyle.Render(dateStr),
            amountStyle.Render(amountStr),
            catStyle.Render(catName),
            desc,
        )
        
        if i == cursor {
            line = lipgloss.NewStyle().
                Background(lipgloss.Color("#374151")).
                Render(line)
        }
        
        s += line + "\n"
    }
    
    // Show pagination info if needed
    if len(txs) > maxVisible {
        paginationStyle := lipgloss.NewStyle().
            Foreground(ColorMuted).
            Italic(true)
        
        s += "\n" + paginationStyle.Render(
            fmt.Sprintf("Showing %d-%d of %d transactions", 
                start+1, end, len(txs)),
        )
    }
    
    return s
}

func (m Model) getCategoryName(categoryID string) string {
    for _, cat := range m.categories {
        if cat.ID == categoryID {
            return cat.Name
        }
    }
    return "Unknown"
}

func (m Model) renderAddTransaction() string {
    return ContentStyle.Render("Add Transaction - Coming soon!")
}

func (m Model) renderCategories() string {
    return ContentStyle.Render("Categories - Coming soon!")
}

// Helper: repeat string n times
func repeat(s string, n int) string {
    result := ""
    for i := 0; i < n; i++ {
        result += s
    }
    return result
}


// loadCategoriesCmd creates a command to load categories
func loadCategoriesCmd(service *app.CategoryService) tea.Cmd {
    return func() tea.Msg {
        categories, err := service.ListCategories()
        return categoriesLoadedMsg{
            categories: categories,
            err:        err,
        }
    }
}

// loadTransactionsCmd creates a command to load transactions
func loadTransactionsCmd(service *app.TransactionService) tea.Cmd {
    return func() tea.Msg {
        transactions, err := service.GetAllTransactions()
        return transactionsLoadedMsg{
            transactions: transactions,
            err:          err,
        }
    }
}

// loadSummaryCmd creates a command to load monthly summary
func loadSummaryCmd(service *app.TransactionService) tea.Cmd {
    return func() tea.Msg {
        now := time.Now()
        summary, err := service.GetMonthlySummary(now.Year(), now.Month())
        return summaryLoadedMsg{
            summary: summary,
            err:     err,
        }
    }
}

func (m Model) renderDashboard() string {
    var s string
    
    // Month header
    now := time.Now()
    monthTitle := lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorPrimary).
        Render(fmt.Sprintf("📊 %s %d Summary", now.Month().String(), now.Year()))
    
    s += monthTitle + "\n\n"
    
    // If no summary loaded yet
    if m.summary == nil {
        return ContentStyle.Render(s + "Loading summary...")
    }
    
    // Monthly summary box
    s += m.renderMonthlySummary()
    s += "\n\n"
    
    // Category breakdown
    s += m.renderCategoryBreakdown()
    s += "\n\n"
    
    // Recent transactions
    s += m.renderRecentTransactions()
    
    return ContentStyle.Render(s)
}

func (m Model) renderMonthlySummary() string {
    // Box style
    boxStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(ColorBorder).
        Padding(1, 2).
        Width(40)
    
    // Income line
    incomeLabel := lipgloss.NewStyle().Foreground(ColorMuted).Render("Income:")
    incomeValue := lipgloss.NewStyle().
        Foreground(ColorSuccess).
        Bold(true).
        Render(formatCurrency(m.summary.TotalIncome))
    incomeLine := fmt.Sprintf("%s  %s", incomeLabel, incomeValue)
    
    // Expenses line
    expenseLabel := lipgloss.NewStyle().Foreground(ColorMuted).Render("Expenses:")
    expenseValue := lipgloss.NewStyle().
        Foreground(ColorDanger).
        Bold(true).
        Render(formatCurrency(m.summary.TotalExpense))
    expenseLine := fmt.Sprintf("%s %s", expenseLabel, expenseValue)
    
    // Separator
    separator := lipgloss.NewStyle().
        Foreground(ColorBorder).
        Render("─────────────────────────────────")
    
    // Net line
    netLabel := lipgloss.NewStyle().
        Foreground(ColorMuted).
        Bold(true).
        Render("Net:")
    
    netColor := ColorSuccess
    if m.summary.NetAmount < 0 {
        netColor = ColorDanger
    }
    netValue := lipgloss.NewStyle().
        Foreground(netColor).
        Bold(true).
        Render(formatCurrency(m.summary.NetAmount))
    netLine := fmt.Sprintf("%s      %s", netLabel, netValue)
    
    // Transaction count
    countLine := lipgloss.NewStyle().
        Foreground(ColorMuted).
        Render(fmt.Sprintf("%d transactions", m.summary.TransactionCount))
    
    // Combine
    content := fmt.Sprintf("%s\n%s\n%s\n%s\n\n%s",
        incomeLine,
        expenseLine,
        separator,
        netLine,
        countLine,
    )
    
    return boxStyle.Render(content)
}

func (m Model) renderCategoryBreakdown() string {
    title := lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorInfo).
        Render("💳 Top Categories")
    
    if len(m.categories) == 0 {
        return title + "\n" + lipgloss.NewStyle().
            Foreground(ColorMuted).
            Render("  No categories yet")
    }
    
    // Calculate totals for each category
    type categoryTotal struct {
        name  string
        color string
        total float64
    }
    
    var totals []categoryTotal
    
    for _, cat := range m.categories {
        total, err := m.transactionService.CalculateCategoryTotal(cat.ID)
        if err != nil {
            continue
        }
        
        // Only show categories with transactions
        if total != 0 {
            totals = append(totals, categoryTotal{
                name:  cat.Name,
                color: cat.Color,
                total: total,
            })
        }
    }
    
    // Sort by absolute value (biggest impact first)
    for i := 0; i < len(totals); i++ {
        for j := i + 1; j < len(totals); j++ {
            if abs(totals[j].total) > abs(totals[i].total) {
                totals[i], totals[j] = totals[j], totals[i]
            }
        }
    }
    
    // Show top 5
    s := title + "\n"
    count := len(totals)
    if count > 5 {
        count = 5
    }
    
    for i := 0; i < count; i++ {
        cat := totals[i]
        
        // Icon based on income/expense
        icon := "💸"
        if cat.total > 0 {
            icon = "💰"
        }
        
        // Color based on category color (use first char as identifier)
        nameStyle := lipgloss.NewStyle().
            Foreground(lipgloss.Color(cat.color))
        
        // Amount style
        amountColor := ColorDanger
        if cat.total > 0 {
            amountColor = ColorSuccess
        }
        amountStyle := lipgloss.NewStyle().
            Foreground(amountColor).
            Bold(true)
        
        line := fmt.Sprintf("  %s %-15s %s",
            icon,
            nameStyle.Render(cat.name),
            amountStyle.Render(formatCurrency(cat.total)),
        )
        
        s += line + "\n"
    }
    
    if len(totals) > 5 {
        s += lipgloss.NewStyle().
            Foreground(ColorMuted).
            Render(fmt.Sprintf("  ... and %d more", len(totals)-5)) + "\n"
    }
    
    return s
}

func (m Model) renderRecentTransactions() string {
    title := lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorInfo).
        Render("📝 Recent Transactions")
    
    if len(m.transactions) == 0 {
        return title + "\n" + lipgloss.NewStyle().
            Foreground(ColorMuted).
            Render("  No transactions yet")
    }
    
    s := title + "\n"
    
    // Show up to 5 most recent
    count := len(m.transactions)
    if count > 5 {
        count = 5
    }
    
    for i := 0; i < count; i++ {
        tx := m.transactions[i]
        
        // Icon
        icon := "💸"
        if tx.Type == domain.TypeIncome {
            icon = "💰"
        }
        
        // Date format
        dateStr := tx.Date.Format("Jan 02")
        dateStyle := lipgloss.NewStyle().Foreground(ColorMuted)
        
        // Amount
        amountColor := ColorDanger
        if tx.Type == domain.TypeIncome {
            amountColor = ColorSuccess
        }
        amountStyle := lipgloss.NewStyle().
            Foreground(amountColor).
            Bold(true)
        
        // Description
        desc := tx.Description
        if len(desc) > 25 {
            desc = desc[:22] + "..."
        }
        
        line := fmt.Sprintf("  %s %s  %s  %s",
            icon,
            dateStyle.Render(dateStr),
            amountStyle.Render(formatCurrency(tx.Amount)),
            desc,
        )
        
        s += line + "\n"
    }
    
    if len(m.transactions) > 5 {
        moreStyle := lipgloss.NewStyle().
            Foreground(ColorMuted).
            Italic(true)
        s += "  " + moreStyle.Render(fmt.Sprintf("... and %d more (press [2] to view all)", len(m.transactions)-5)) + "\n"
    }
    
    return s
}

// Helper functions

func formatCurrency(amount float64) string {
    sign := ""
    if amount < 0 {
        sign = "-"
        amount = -amount
    } else if amount > 0 {
        sign = "+"
    }
    
    return fmt.Sprintf("%s$%.2f", sign, amount)
}

func abs(x float64) float64 {
    if x < 0 {
        return -x
    }
    return x
}

// TransactionsModel holds state for the transactions screen
type TransactionsModel struct {
    cursor       int  // Selected transaction index
    filter       TransactionFilter
    filteredTxs  []domain.Transaction
}

type TransactionFilter int

const (
    FilterAll TransactionFilter = iota
    FilterCurrentMonth
)

func (f TransactionFilter) String() string {
    switch f {
    case FilterAll:
        return "All Time"
    case FilterCurrentMonth:
        return "Current Month"
    default:
        return "Unknown"
    }
}

func (m Model) initTransactionsModel() TransactionsModel {
    return TransactionsModel{
        cursor:      0,
        filter:      FilterCurrentMonth,
        filteredTxs: m.filterTransactions(FilterCurrentMonth),
    }
}

func (m Model) filterTransactions(filter TransactionFilter) []domain.Transaction {
    if filter == FilterAll {
        return m.transactions
    }
    
    // Filter current month
    now := time.Now()
    var filtered []domain.Transaction
    
    for _, tx := range m.transactions {
        if tx.Date.Year() == now.Year() && tx.Date.Month() == now.Month() {
            filtered = append(filtered, tx)
        }
    }
    
    return filtered
}