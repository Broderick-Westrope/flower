package msgs

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TickMsg is sent every second to drive timer updates.
type TickMsg struct{}

// Tick returns a command that sends a TickMsg after one second.
func Tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// StartSessionMsg requests starting a new flow session.
type StartSessionMsg struct{ Task string }

// ShowLogMsg requests switching to the session log view.
type ShowLogMsg struct{}

// BackMsg requests returning to the previous view.
type BackMsg struct{}

// ErrorMsg carries an error to display to the user.
type ErrorMsg struct{ Err error }
