package shared

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type TickMsg struct{}

func Tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}
