package transactions

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/simdangelo/fingo/internal/app"
	"github.com/simdangelo/fingo/internal/domain"
	"github.com/simdangelo/fingo/internal/tui/components"
	"github.com/simdangelo/fingo/internal/tui/styles"
)

// ── Tab filter ───────────────────────────────────────────────────────────────

type tabFilter int

const (
	tabAll tabFilter = iota
	tabExpenses
	tabIncome
)

// ── Messages ─────────────────────────────────────────────────────────────────

// txLoadedMsg is sent when transactions finish loading from the service.
type txLoadedMsg struct {
	transactions []domain.Transaction
	err          error
}

// ── Model ─────────────────────────────────────────────────────────────────────

type Model struct {
	service      *app.TransactionService
	transactions []domain.Transaction // all loaded transactions
	activeTab    tabFilter
	table        table.Model
	width        int
	height       int
	err          error
	loading      bool
}

func New(service *app.TransactionService) Model {
	m := Model{
		service: service,
		loading: true,
	}
	return m
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return loadTransactions(m.service)
}

// loadTransactions is a Bubble Tea command: runs in a goroutine, sends a msg back.
func loadTransactions(svc *app.TransactionService) tea.Cmd {
	return func() tea.Msg {
		txs, err := svc.GetAllTransactions()
		return txLoadedMsg{transactions: txs, err: err}
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.rebuildTable()

	case txLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.transactions = msg.transactions
		m.rebuildTable()

	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.activeTab > tabAll {
				m.activeTab--
				m.rebuildTable()
			}
		case "right", "l":
			if m.activeTab < tabIncome {
				m.activeTab++
				m.rebuildTable()
			}
		case "r":
			m.loading = true
			return m, loadTransactions(m.service)
		default:
			// Forward all other keys to the table (↑↓ scrolling, etc.)
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorMuted)).
			Render("\n  Loading transactions…")
	}
	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorRed)).
			Render(fmt.Sprintf("\n  Error: %v", m.err))
	}

	var b strings.Builder

	// ── Sub-tabs: All / Expenses / Income ────────────────────────────────────
	b.WriteString(renderTabs(m.activeTab))
	b.WriteString("\n")

	// ── Summary line ─────────────────────────────────────────────────────────
	b.WriteString(m.renderSummary())
	b.WriteString("\n\n")

	// ── Table ─────────────────────────────────────────────────────────────────
	b.WriteString(m.table.View())

	return b.String()
}

// Bindings returns the status bar shortcuts for this page.
func (m Model) Bindings() []components.Binding {
	return []components.Binding{
		{"n", "New"},
		{"d", "Delete"},
		{"r", "Refresh"},
		{"←→", "All/Exp/Inc"},
		{"↑↓", "Navigate"},
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// filtered returns only the transactions matching the active tab.
func (m Model) filtered() []domain.Transaction {
	var out []domain.Transaction
	for _, tx := range m.transactions {
		switch m.activeTab {
		case tabAll:
			out = append(out, tx)
		case tabExpenses:
			if tx.Type == domain.TypeExpense {
				out = append(out, tx)
			}
		case tabIncome:
			if tx.Type == domain.TypeIncome {
				out = append(out, tx)
			}
		}
	}
	return out
}

// rebuildTable recreates the bubbles/table with current data and dimensions.
func (m *Model) rebuildTable() {
	txs := m.filtered()

	// Column widths that adapt to terminal width.
	// Minimum sensible width is ~80 chars.
	w := m.width
	if w < 80 {
		w = 80
	}
	// Fixed columns: date(14) + type(10) + amount(14) + padding ≈ 44
	// Remainder goes to description and category, split 55/45.
	remaining := w - 44 - 6 // 6 for table borders/padding
	descW := remaining * 55 / 100
	catW := remaining - descW

	columns := []table.Column{
		{Title: "DATE", Width: 14},
		{Title: "DESCRIPTION", Width: descW},
		{Title: "CATEGORY", Width: catW},
		{Title: "TYPE", Width: 10},
		{Title: "AMOUNT", Width: 14},
	}

	rows := make([]table.Row, len(txs))
	for i, tx := range txs {
		rows[i] = table.Row{
			tx.Date.Format("Jan 02, 2006"),
			tx.Description,
			tx.CategoryID, // Step 4 will resolve category names
			string(tx.Type),
			formatAmount(tx),
		}
	}

	// Table height: total height minus rows used by header, tabs, summary, status.
	tableHeight := m.height - 10
	if tableHeight < 5 {
		tableHeight = 5
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	// Apply styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		Background(lipgloss.Color(styles.ColorSurface)).
		Foreground(lipgloss.Color(styles.ColorAccent)).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(styles.ColorBorder)).
		BorderBottom(true)
	s.Selected = s.Selected.
		Background(lipgloss.Color(styles.ColorSurface)).
		Foreground(lipgloss.Color(styles.ColorAccent)).
		Bold(true)
	t.SetStyles(s)

	m.table = t
}

func renderTabs(active tabFilter) string {
	labels := []struct {
		label string
		tab   tabFilter
	}{
		{"All", tabAll},
		{"Expenses", tabExpenses},
		{"Income", tabIncome},
	}

	var parts []string
	for _, l := range labels {
		text := l.label
		if l.tab == active {
			parts = append(parts, lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.ColorAccent)).
				Bold(true).
				Render(text))
		} else {
			parts = append(parts, lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.ColorMuted)).
				Render(text))
		}
	}

	return "  " + strings.Join(parts, "   ")
}

func (m Model) renderSummary() string {
	txs := m.filtered()
	var income, expense float64
	for _, tx := range txs {
		if tx.Type == domain.TypeIncome {
			income += tx.Amount
		} else {
			expense += tx.Amount
		}
	}
	net := income - expense

	count := lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorMuted)).
		Render(fmt.Sprintf("  %d transactions", len(txs)))
	inc := lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorGreen)).
		Render(fmt.Sprintf("  Income: +$%.2f", income))
	exp := lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorRed)).
		Render(fmt.Sprintf("  Expenses: -$%.2f", expense))

	var netStyle lipgloss.Style
	if net >= 0 {
		netStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorGreen))
	} else {
		netStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorRed))
	}
	netStr := netStyle.Render(fmt.Sprintf("  Net: $%.2f", net))

	return count + "  •" + inc + "  •" + exp + "  •" + netStr
}

func formatAmount(tx domain.Transaction) string {
	if tx.Type == domain.TypeIncome {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorGreen)).
			Render(fmt.Sprintf("+$%.2f", tx.Amount))
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorRed)).
		Render(fmt.Sprintf("-$%.2f", tx.Amount))
}