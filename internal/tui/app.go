package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/simdangelo/fingo/internal/app"
	"github.com/simdangelo/fingo/internal/tui/components"
	"github.com/simdangelo/fingo/internal/tui/pages/transactions"
	"github.com/simdangelo/fingo/internal/tui/styles"
)

type Model struct {
	width      int
	height     int
	activePage components.Page
	txPage     transactions.Model
}

func New(txService *app.TransactionService, catService *app.CategoryService) Model {
	return Model{
		activePage: components.PageDashboard,
		txPage:     transactions.New(txService, catService),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.txPage.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		var c tea.Cmd
		m.txPage, c = m.txPage.Update(msg)
		cmds = append(cmds, c)

	case tea.KeyMsg:
		// Always allow quitting, even from inside a form.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// If the active page is capturing input (e.g. a form is open),
		// forward ALL keys to it — do not intercept anything for navigation.
		if m.activePage == components.PageTransactions && m.txPage.IsCapturingInput() {
			var c tea.Cmd
			m.txPage, c = m.txPage.Update(msg)
			cmds = append(cmds, c)
			break
		}

		// Normal navigation — no form is open.
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "1":
			m.activePage = components.PageDashboard
		case "2":
			m.activePage = components.PageTransactions
		case "3":
			m.activePage = components.PageAccounts
		case "4":
			m.activePage = components.PageBudgets
		case "5":
			m.activePage = components.PageReports
		case "6":
			m.activePage = components.PageGoals
		case "7":
			m.activePage = components.PageRecurring
		case "8":
			m.activePage = components.PageSettings
		case "tab":
			m.activePage = (m.activePage + 1) % 8
		case "shift+tab":
			m.activePage = (m.activePage + 7) % 8
		default:
			switch m.activePage {
			case components.PageTransactions:
				var c tea.Cmd
				m.txPage, c = m.txPage.Update(msg)
				cmds = append(cmds, c)
			}
		}

	default:
		// Non-key messages (loaded data, etc.) always forwarded.
		var c tea.Cmd
		m.txPage, c = m.txPage.Update(msg)
		cmds = append(cmds, c)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	header := components.RenderHeader(m.activePage, m.width)

	headerHeight := 2
	statusHeight := 2
	contentHeight := m.height - headerHeight - statusHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	var content string
	var pageBindings []components.Binding

	switch m.activePage {
	case components.PageTransactions:
		content = m.txPage.View()
		pageBindings = m.txPage.Bindings()
	default:
		pageName := pageTitle(m.activePage)
		placeholder := fmt.Sprintf("  [ %s ]  — coming soon", pageName)
		content = lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorMuted)).
			PaddingLeft(2).
			PaddingTop(1).
			Render(placeholder)
		pageBindings = []components.Binding{
			{"1-8", "Navigate"},
			{"tab", "Cycle"},
		}
	}

	contentBox := lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight).
		Background(lipgloss.Color(styles.ColorBg)).
		Render(content)

	statusBar := components.RenderStatusBar(pageBindings, m.width)

	return header + "\n" + contentBox + "\n" + statusBar
}

func pageTitle(p components.Page) string {
	switch p {
	case components.PageDashboard:
		return "Dashboard"
	case components.PageTransactions:
		return "Transactions"
	case components.PageAccounts:
		return "Accounts & Balances"
	case components.PageBudgets:
		return "Budgets"
	case components.PageReports:
		return "Reports"
	case components.PageGoals:
		return "Goals & Savings"
	case components.PageRecurring:
		return "Recurring"
	case components.PageSettings:
		return "Settings"
	default:
		return "Unknown"
	}
}