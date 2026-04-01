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

// CancelSessionMsg requests cancelling the current session.
type CancelSessionMsg struct{}

// DeleteSessionMsg requests soft-deleting a specific completed session.
// Index is into the full CompletedSessions slice (not the active-only list).
type DeleteSessionMsg struct{ Index int }

// DeleteAllSessionsMsg requests soft-deleting all completed sessions.
type DeleteAllSessionsMsg struct{}

// RequestDeleteSessionMsg is emitted by the log view with the active-list index.
// The model maps this to the full-slice index before creating a DeleteSessionMsg.
type RequestDeleteSessionMsg struct{ ActiveIndex int }

// ConfirmAction represents a pending action that requires user confirmation.
type ConfirmAction struct {
	Prompt string
	OnYes  tea.Msg
}

// RequestConfirmMsg asks the model to show a confirmation prompt.
type RequestConfirmMsg struct{ Action ConfirmAction }

// ConfirmResultMsg carries the user's yes/no answer.
type ConfirmResultMsg struct{ Confirmed bool }
