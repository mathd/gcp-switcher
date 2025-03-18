package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles defines the UI styling configuration
type Styles struct {
	App           lipgloss.Style
	Title         lipgloss.Style
	Subtitle      lipgloss.Style
	Info          lipgloss.Style
	Success       lipgloss.Style
	Error         lipgloss.Style
	Highlight     lipgloss.Style
	FocusedButton lipgloss.Style
	BlurredButton lipgloss.Style
	ActiveItem    lipgloss.Style
}

// NewStyles initializes and returns the UI styles
func NewStyles() Styles {
	return Styles{
		App: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("99")).
			Padding(1, 2),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("213")).
			Bold(true).
			MarginLeft(2),

		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("105")),

		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("247")),

		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("84")),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")),

		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("159")).
			Bold(true),

		FocusedButton: lipgloss.NewStyle().
			Foreground(lipgloss.Color("231")).
			Background(lipgloss.Color("99")).
			Padding(0, 2).
			Bold(true),

		BlurredButton: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 2),

		ActiveItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("159")).
			Bold(true),
	}
}
