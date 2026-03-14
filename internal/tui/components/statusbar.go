package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/simdangelo/fingo/internal/tui/styles"
)

// Binding is a single key + description pair shown in the status bar.
type Binding struct {
	Key  string
	Desc string
}

// globalBindings are always shown on the right side of the status bar.
var globalBindings = []Binding{
	{"?", "Help"},
	{"q", "Quit"},
}

// RenderStatusBar returns the bottom bar as a string.
// pageBindings are context-sensitive shortcuts supplied by the active page.
func RenderStatusBar(pageBindings []Binding, width int) string {
	// Separator line
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorBorder)).
		Render(strings.Repeat("─", width))

	// Render a single key+desc binding
	renderBinding := func(b Binding) string {
		key := styles.StatusKey.Render("[" + b.Key + "]")
		desc := styles.StatusBar.Render(b.Desc)
		return key + desc
	}

	// Left side: page-specific bindings
	var leftParts []string
	for _, b := range pageBindings {
		leftParts = append(leftParts, renderBinding(b))
	}
	left := strings.Join(leftParts, "  ")

	// Right side: global bindings
	var rightParts []string
	for _, b := range globalBindings {
		rightParts = append(rightParts, renderBinding(b))
	}
	right := strings.Join(rightParts, "  ")

	// Pad left + right to fill the full width
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	gap := width - leftWidth - rightWidth - 2 // -2 for side padding
	if gap < 1 {
		gap = 1
	}
	padding := strings.Repeat(" ", gap)

	bar := styles.StatusBar.
		Width(width).
		Render(left + padding + right)

	return separator + "\n" + bar
}