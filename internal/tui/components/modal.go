package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/simdangelo/fingo/internal/tui/styles"
)

// RenderModal wraps content in a centered bordered box with a title.
// termW and termH are the full terminal dimensions.
func RenderModal(title, content string, termW, termH int) string {
	boxW := 60
	if termW < boxW+4 {
		boxW = termW - 4
	}

	box := lipgloss.NewStyle().
		Width(boxW).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(styles.ColorAccent)).
		Background(lipgloss.Color(styles.ColorSurface)).
		Foreground(lipgloss.Color(styles.ColorText)).
		Padding(1, 2).
		Render(content)

	// Stamp the title onto the top border
	titled := lipgloss.NewStyle().Render(
		strings.Replace(box, "╭", "╭─ "+title+" ", 1),
	)

	// Center horizontally and vertically
	return lipgloss.Place(termW, termH, lipgloss.Center, lipgloss.Center, titled,
		lipgloss.WithWhitespaceBackground(lipgloss.Color(styles.ColorBg)),
	)
}