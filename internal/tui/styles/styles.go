package styles

import "github.com/charmbracelet/lipgloss"

// Color palette — one place to change the whole app's look.
const (
	ColorBg        = "#1a1b26" // dark background
	ColorSurface   = "#24283b" // slightly lighter surface
	ColorBorder    = "#414868" // panel borders
	ColorMuted     = "#565f89" // dimmed / inactive text
	ColorText      = "#c0caf5" // primary text
	ColorTextBold  = "#ffffff" // bright headings
	ColorAccent    = "#7aa2f7" // blue accent (active tab, highlights)
	ColorGreen     = "#9ece6a" // income / positive
	ColorRed       = "#f7768e" // expense / negative / over budget
	ColorYellow    = "#e0af68" // warning
	ColorCyan      = "#7dcfff" // info / links
)

// ── Base styles ──────────────────────────────────────────────────────────────

var Base = lipgloss.NewStyle().
	Background(lipgloss.Color(ColorBg)).
	Foreground(lipgloss.Color(ColorText))

// ── Header (top nav bar) ─────────────────────────────────────────────────────

var Header = lipgloss.NewStyle().
	Background(lipgloss.Color(ColorSurface)).
	Foreground(lipgloss.Color(ColorText)).
	PaddingLeft(1).
	PaddingRight(1)

var AppTitle = lipgloss.NewStyle().
	Background(lipgloss.Color(ColorAccent)).
	Foreground(lipgloss.Color(ColorBg)).
	Bold(true).
	Padding(0, 1)

var TabInactive = lipgloss.NewStyle().
	Background(lipgloss.Color(ColorSurface)).
	Foreground(lipgloss.Color(ColorMuted)).
	Padding(0, 1)

var TabActive = lipgloss.NewStyle().
	Background(lipgloss.Color(ColorBg)).
	Foreground(lipgloss.Color(ColorAccent)).
	Bold(true).
	Padding(0, 1)

// ── Status bar (bottom bar) ──────────────────────────────────────────────────

var StatusBar = lipgloss.NewStyle().
	Background(lipgloss.Color(ColorSurface)).
	Foreground(lipgloss.Color(ColorMuted)).
	PaddingLeft(1).
	PaddingRight(1)

var StatusKey = lipgloss.NewStyle().
	Background(lipgloss.Color(ColorSurface)).
	Foreground(lipgloss.Color(ColorAccent)).
	Bold(true)

// ── Content area ─────────────────────────────────────────────────────────────

var ContentArea = lipgloss.NewStyle().
	Background(lipgloss.Color(ColorBg)).
	Foreground(lipgloss.Color(ColorText)).
	PaddingLeft(2).
	PaddingTop(1)