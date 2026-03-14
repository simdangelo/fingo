package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/simdangelo/fingo/internal/tui/styles"
)

// Page constants — used throughout the app to identify the active page.
type Page int

const (
	PageDashboard Page = iota
	PageTransactions
	PageAccounts
	PageBudgets
	PageReports
	PageGoals
	PageRecurring
	PageSettings
)

// tab holds the display label for each top-level page.
var tabs = []struct {
	page  Page
	label string
}{
	{PageDashboard, "Dashboard"},
	{PageTransactions, "Transactions"},
	{PageAccounts, "Accounts"},
	{PageBudgets, "Budgets"},
	{PageReports, "Reports"},
	{PageGoals, "Goals"},
	{PageRecurring, "Recurring"},
	{PageSettings, "Settings"},
}

// RenderHeader returns the top nav bar as a string.
// width is the terminal width so the bar fills the full line.
func RenderHeader(activePage Page, width int) string {
	// App title badge
	title := styles.AppTitle.Render("💰 Fingo")

	// Build tab list
	var tabParts []string
	for i, t := range tabs {
		label := "[" + string(rune('1'+i)) + "]" + t.label
		if t.page == activePage {
			tabParts = append(tabParts, styles.TabActive.Render(label))
		} else {
			tabParts = append(tabParts, styles.TabInactive.Render(label))
		}
	}
	tabRow := strings.Join(tabParts, "")

	// Combine title + tabs, pad to full width
	content := title + "  " + tabRow
	bar := styles.Header.
		Width(width).
		Render(content)

	// Separator line
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorBorder)).
		Render(strings.Repeat("─", width))

	return bar + "\n" + separator
}