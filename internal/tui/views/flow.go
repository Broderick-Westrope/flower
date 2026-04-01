package views

import (
	"time"

	"github.com/Broderick-Westrope/flower/internal/flowtime"
	"github.com/Broderick-Westrope/flower/internal/tui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FlowView displays the running flow timer with a spinner.
type FlowView struct {
	session *flowtime.Session
	spinner spinner.Model
}

// NewFlowView creates a FlowView with a MiniDot spinner.
func NewFlowView() *FlowView {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	return &FlowView{spinner: s}
}

// SetSession updates the active session displayed by this view.
func (v *FlowView) SetSession(s *flowtime.Session) {
	v.session = s
}

// Init returns the spinner tick command.
func (v *FlowView) Init() tea.Cmd {
	return v.spinner.Tick
}

// Update handles spinner updates.
func (v *FlowView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	v.spinner, cmd = v.spinner.Update(msg)
	return cmd
}

// View renders the flow timer screen.
func (v *FlowView) View() string {
	if v.session == nil {
		return ""
	}

	elapsed := time.Since(v.session.StartTime)
	timerLine := styles.Timer.Render(flowtime.FormatDuration(elapsed)) + " " + v.spinner.View()
	taskLine := styles.TaskName.Render(v.session.Task)
	helpBar := RenderHelpBar([]KeyBinding{
		{Key: "space", Description: "break"},
		{Key: "s", Description: "stop"},
		{Key: "l", Description: "log"},
		{Key: "q", Description: "quit"},
	})

	contentWidth := max(
		lipgloss.Width(timerLine),
		lipgloss.Width(taskLine),
		lipgloss.Width(helpBar),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		styles.Title.Render("Flowing"),
		"",
		taskLine,
		timerLine,
		"",
		styles.Separator(contentWidth),
		helpBar,
	)
}
