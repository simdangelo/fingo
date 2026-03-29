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

// catsLoadedMsg is sent when categories finish loading.
type catsLoadedMsg struct {
	categories []domain.Category
	err        error
}

// ── Model ─────────────────────────────────────────────────────────────────────

type Model struct {
	txService  *app.TransactionService
	catService *app.CategoryService

	transactions []domain.Transaction
	categories   []domain.Category

	activeTab tabFilter
	table     table.Model
	width     int
	height    int
	err       error
	loading   bool

	// Form overlay
	showForm bool
	form     formModel
}

func New(txService *app.TransactionService, catService *app.CategoryService) Model {
	return Model{
		txService:  txService,
		catService: catService,
		loading:    true,
	}
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadTransactions(m.txService),
		loadCategories(m.catService),
	)
}

func loadTransactions(svc *app.TransactionService) tea.Cmd {
	return func() tea.Msg {
		txs, err := svc.GetAllTransactions()
		return txLoadedMsg{transactions: txs, err: err}
	}
}

func loadCategories(svc *app.CategoryService) tea.Cmd {
	return func() tea.Msg {
		cats, err := svc.ListCategories()
		return catsLoadedMsg{categories: cats, err: err}
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// ── Form is open: route all messages there ────────────────────────────────
	if m.showForm {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				m.showForm = false
				return m, nil
			}
		case formSavedMsg:
			// Transaction saved — close form and reload list
			m.showForm = false
			m.loading = true
			return m, loadTransactions(m.txService)
		case formErrMsg:
			m.form.err = msg.err.Error()
			m.form.submitting = false
			return m, nil
		case categoryCreatedMsg:
			// Append the new category and select it
			m.categories = append(m.categories, msg.cat)
			m.form.categories = m.categories
			m.form.categoryIndex = len(m.categories) - 1
			m.form.newCatMode = false
			return m, nil
		}
		var cmd tea.Cmd
		m.form, cmd = m.form.Update(msg)
		return m, cmd
	}

	// ── Normal table mode ─────────────────────────────────────────────────────
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

	case catsLoadedMsg:
		if msg.err == nil {
			m.categories = msg.categories
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			// Open the form, seeding it with the already-loaded categories
			m.form = newForm(m.txService, m.catService, m.categories)
			m.showForm = true
		case "left":
			if m.activeTab > tabAll {
				m.activeTab--
				m.rebuildTable()
			}
		case "right":
			if m.activeTab < tabIncome {
				m.activeTab++
				m.rebuildTable()
			}
		case "r":
			m.loading = true
			return m, loadTransactions(m.txService)
		default:
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	// If form is open, render it as an overlay on top of the table
	if m.showForm {
		return components.RenderModal("NEW TRANSACTION", m.form.View(), m.width, m.height-4)
	}

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

	b.WriteString(renderTabs(m.activeTab))
	b.WriteString("\n")
	b.WriteString(m.renderSummary())
	b.WriteString("\n\n")
	b.WriteString(m.table.View())

	return b.String()
}

// Bindings returns context-sensitive status bar shortcuts.
func (m Model) Bindings() []components.Binding {
	if m.showForm {
		return []components.Binding{
			{"tab/↑↓", "Move field"},
			{"←→", "Toggle"},
			{"enter", "Save"},
			{"esc", "Cancel"},
		}
	}
	return []components.Binding{
		{"n", "New"},
		{"r", "Refresh"},
		{"←→", "All/Exp/Inc"},
		{"↑↓", "Navigate"},
	}
}

// IsCapturingInput returns true when the page has a modal/form open
// that needs all keystrokes, so the root model must not intercept them.
func (m Model) IsCapturingInput() bool {
	return m.showForm
}

// ── Helpers ───────────────────────────────────────────────────────────────────

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

// categoryName resolves a category ID to its display name.
func (m Model) categoryName(id string) string {
	for _, c := range m.categories {
		if c.ID == id {
			return c.Name
		}
	}
	return id // fallback to raw ID if not found
}

func (m *Model) rebuildTable() {
	txs := m.filtered()

	w := m.width
	if w < 80 {
		w = 80
	}
	remaining := w - 44 - 6
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
			m.categoryName(tx.CategoryID), // now resolves to real name
			string(tx.Type),
			formatAmount(tx),
		}
	}

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