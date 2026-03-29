package transactions

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/simdangelo/fingo/internal/app"
	"github.com/simdangelo/fingo/internal/domain"
	"github.com/simdangelo/fingo/internal/tui/styles"
)

// ── Field indices ─────────────────────────────────────────────────────────────

const (
	fieldDate        = iota // text input
	fieldDescription        // text input
	fieldAmount             // text input
	fieldCategory           // select with ←→
	fieldType               // select with ←→
	fieldCount
)

// ── Messages ──────────────────────────────────────────────────────────────────

type formSavedMsg struct{}
type formErrMsg struct{ err error }
type categoryCreatedMsg struct{ cat domain.Category }

// ── Form model ────────────────────────────────────────────────────────────────

type formModel struct {
	// Text inputs (date, description, amount)
	inputs     [3]textinput.Model
	focusIndex int

	// Category picker
	categories    []domain.Category
	categoryIndex int

	// "New category" sub-form (shown inline when user presses 'c')
	newCatMode  bool
	newCatInput textinput.Model

	// Transaction type
	txType domain.TransactionType

	// Services
	txService  *app.TransactionService
	catService *app.CategoryService

	// Feedback
	err        string
	submitting bool
}

func newForm(txService *app.TransactionService, catService *app.CategoryService, cats []domain.Category) formModel {
	date := textinput.New()
	date.Placeholder = "YYYY-MM-DD"
	date.SetValue(time.Now().Format("2006-01-02"))
	date.Focus()
	date.Width = 36

	desc := textinput.New()
	desc.Placeholder = "e.g. Grocery run"
	desc.Width = 36

	amount := textinput.New()
	amount.Placeholder = "0.00"
	amount.Width = 36

	newCat := textinput.New()
	newCat.Placeholder = "Category name"
	newCat.Width = 28

	return formModel{
		inputs:     [3]textinput.Model{date, desc, amount},
		newCatInput: newCat,
		categories: cats,
		txType:     domain.TypeExpense,
		txService:  txService,
		catService: catService,
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func (f formModel) Update(msg tea.Msg) (formModel, tea.Cmd) {
	// ── New-category sub-form ─────────────────────────────────────────────────
	if f.newCatMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				f.newCatMode = false
				return f, nil
			case "enter":
				name := strings.TrimSpace(f.newCatInput.Value())
				if name == "" {
					f.err = "Category name cannot be empty"
					return f, nil
				}
				svc := f.catService
				return f, func() tea.Msg {
					// Use a default color; Settings page will let users customize later
					cat, err := svc.CreateCategory(name, "#7aa2f7")
					if err != nil {
						return formErrMsg{err: err}
					}
					return categoryCreatedMsg{cat: *cat}
				}
			}
		}
		var cmd tea.Cmd
		f.newCatInput, cmd = f.newCatInput.Update(msg)
		return f, cmd
	}

	// ── Main form ─────────────────────────────────────────────────────────────
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "tab", "down":
			f.focusIndex = (f.focusIndex + 1) % fieldCount
			f.syncFocus()

		case "shift+tab", "up":
			f.focusIndex = (f.focusIndex + fieldCount - 1) % fieldCount
			f.syncFocus()

		case "left", "h":
			switch f.focusIndex {
			case fieldType:
				f.txType = domain.TypeExpense
			case fieldCategory:
				if len(f.categories) > 0 {
					f.categoryIndex = (f.categoryIndex + len(f.categories) - 1) % len(f.categories)
				}
			}

		case "right", "l":
			switch f.focusIndex {
			case fieldType:
				f.txType = domain.TypeIncome
			case fieldCategory:
				if len(f.categories) > 0 {
					f.categoryIndex = (f.categoryIndex + 1) % len(f.categories)
				}
			}

		case "c":
			// Open the new-category sub-form only when on the Category field
			if f.focusIndex == fieldCategory {
				f.newCatMode = true
				f.newCatInput.SetValue("")
				f.newCatInput.Focus()
				f.err = ""
			}

		case "enter":
			f.err = ""
			return f, f.submit()
		}

		// Forward typing keys to the focused text input
		if f.focusIndex < 3 {
			var cmd tea.Cmd
			f.inputs[f.focusIndex], cmd = f.inputs[f.focusIndex].Update(msg)
			return f, cmd
		}
	}

	return f, nil
}

func (f *formModel) syncFocus() {
	for i := range f.inputs {
		f.inputs[i].Blur()
	}
	if f.focusIndex < 3 {
		f.inputs[f.focusIndex].Focus()
	}
}

func (f *formModel) submit() tea.Cmd {
	dateStr := strings.TrimSpace(f.inputs[fieldDate].Value())
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		f.err = "Invalid date — use YYYY-MM-DD"
		return nil
	}

	desc := strings.TrimSpace(f.inputs[fieldDescription].Value())
	if desc == "" {
		f.err = "Description cannot be empty"
		return nil
	}

	amountStr := strings.TrimSpace(f.inputs[fieldAmount].Value())
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		f.err = "Amount must be a positive number"
		return nil
	}

	if len(f.categories) == 0 {
		f.err = "No categories — press 'c' on the Category field to create one"
		return nil
	}
	categoryID := f.categories[f.categoryIndex].ID

	f.submitting = true
	txType := f.txType
	svc := f.txService

	return func() tea.Msg {
		err := svc.AddTransaction(date, amount, txType, categoryID, desc)
		if err != nil {
			return formErrMsg{err: err}
		}
		return formSavedMsg{}
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func (f formModel) View() string {
	var b strings.Builder

	focused := func(idx int) bool { return f.focusIndex == idx }

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorMuted)).
		Width(14)

	activeLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorAccent)).
		Bold(true).
		Width(14)

	renderRow := func(label string, isFocused bool, value string) string {
		ls := labelStyle
		if isFocused {
			ls = activeLabel
		}
		return ls.Render(label+":") + " " + value
	}

	// Date
	b.WriteString(renderRow("Date", focused(fieldDate), f.inputs[fieldDate].View()))
	b.WriteString("\n\n")

	// Description
	b.WriteString(renderRow("Description", focused(fieldDescription), f.inputs[fieldDescription].View()))
	b.WriteString("\n\n")

	// Amount
	b.WriteString(renderRow("Amount", focused(fieldAmount), f.inputs[fieldAmount].View()))
	b.WriteString("\n\n")
	
	// Type
	expStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorMuted))
	incStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorMuted))
	if f.txType == domain.TypeExpense {
		expStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorRed)).Bold(true)
	} else {
		incStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorGreen)).Bold(true)
	}
	typeVal := expStyle.Render("Expense") + "  /  " + incStyle.Render("Income")
	if focused(fieldType) {
		typeVal += lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorMuted)).
			Render("  [←→] toggle")
	}
	b.WriteString(renderRow("Type", focused(fieldType), typeVal))
	b.WriteString("\n\n")

	// Category
	catVal := ""
	if f.newCatMode {
		catVal = "New: " + f.newCatInput.View() +
			lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorMuted)).
				Render("  [enter] create  [esc] cancel")
	} else if len(f.categories) == 0 {
		catVal = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorYellow)).
			Render("(none) — press 'c' to create one")
	} else {
		catVal = fmt.Sprintf("◄  %s  ►  ", f.categories[f.categoryIndex].Name)
		if focused(fieldCategory) {
			catVal += lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorMuted)).
				Render("  [←→] cycle  [c] new")
		}
	}
	b.WriteString(renderRow("Category", focused(fieldCategory) && !f.newCatMode, catVal))
	b.WriteString("\n\n")


	// Error
	if f.err != "" {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorRed)).
			Render("✗ " + f.err))
		b.WriteString("\n\n")
	}

	// Footer
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorMuted)).
		Render("[tab/↑↓] move field   [enter] save   [esc] cancel"))

	return b.String()
}