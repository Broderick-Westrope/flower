package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Container wraps the entire TUI with a rounded border and padding.
	Container = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			Margin(1, 2)

	// Title is a bold style for state indicators like "Flowing", "Break", etc.
	Title = lipgloss.NewStyle().Bold(true)

	// Timer is a bold, coloured style for the prominent time display.
	Timer = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))

	// TaskName styles the task description text.
	TaskName = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

	// HelpBar renders faint keyboard-shortcut hints.
	HelpBar = lipgloss.NewStyle().Faint(true)

	// ErrorText renders error messages in red.
	ErrorText = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	// TableHeader is bold text for log table column headers.
	TableHeader = lipgloss.NewStyle().Bold(true)
)

// Separator returns a horizontal rule of the given width using the "─" character.
func Separator(width int) string {
	if width <= 0 {
		return ""
	}
	return HelpBar.Render(strings.Repeat("─", width))
}
