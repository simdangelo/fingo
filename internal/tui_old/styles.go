package tui

import "github.com/charmbracelet/lipgloss"

// Colors
var (
    ColorPrimary   = lipgloss.Color("#7C3AED")  // Purple
    ColorSuccess   = lipgloss.Color("#10B981")  // Green
    ColorDanger    = lipgloss.Color("#EF4444")  // Red
    ColorWarning   = lipgloss.Color("#F59E0B")  // Orange
    ColorInfo      = lipgloss.Color("#3B82F6")  // Blue
    ColorMuted     = lipgloss.Color("#6B7280")  // Gray
    ColorBorder    = lipgloss.Color("#374151")  // Dark gray
)

// Base styles
var (
    // Title bar style
    TitleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorPrimary).
        Background(lipgloss.Color("#1F2937")).
        Padding(0, 1)
    
    // Tab styles
    ActiveTabStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorPrimary).
        Background(lipgloss.Color("#1F2937")).
        Padding(0, 2)
    
    InactiveTabStyle = lipgloss.NewStyle().
        Foreground(ColorMuted).
        Background(lipgloss.Color("#1F2937")).
        Padding(0, 2)
    
    // Content area
    ContentStyle = lipgloss.NewStyle().
        Padding(1, 2)
    
    // Help text
    HelpStyle = lipgloss.NewStyle().
        Foreground(ColorMuted).
        Padding(0, 1)
    
    // Error style
    ErrorStyle = lipgloss.NewStyle().
        Foreground(ColorDanger).
        Bold(true)
    
    // Success style
    SuccessStyle = lipgloss.NewStyle().
        Foreground(ColorSuccess).
        Bold(true)
)