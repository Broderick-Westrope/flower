package views

import (
	"github.com/Broderick-Westrope/flower/internal/tui/msgs"
	"github.com/Broderick-Westrope/flower/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// IdleView lets the user type a task name and start a session.
type IdleView struct {
	input textinput.Model
}

// NewIdleView creates a focused text input with placeholder and char limit.
func NewIdleView() *IdleView {
	ti := textinput.New()
	ti.Placeholder = "Task name..."
	ti.CharLimit = 100
	ti.Focus()
	return &IdleView{input: ti}
}

// Init returns the text-input blink command.
func (v *IdleView) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles key events for the idle view.
func (v *IdleView) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyMsg); ok {
		// When the input is empty, intercept shortcut keys.
		if v.input.Value() == "" {
			switch msg.String() {
			case "enter":
				return nil // ignore enter on empty input
			case "l":
				return func() tea.Msg { return msgs.ShowLogMsg{} }
			case "q":
				return tea.Quit
			}
		} else {
			// When there is text, only intercept enter.
			if msg.String() == "enter" {
				task := v.input.Value()
				return func() tea.Msg { return msgs.StartSessionMsg{Task: task} }
			}
		}
	}

	var cmd tea.Cmd
	v.input, cmd = v.input.Update(msg)
	return cmd
}

// View renders the idle screen.
func (v *IdleView) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		styles.Title.Render("Flower"),
		"",
		"What are you working on?",
		"> "+v.input.View(),
		"",
		styles.Separator(20),
		RenderHelpBar([]KeyBinding{
			{Key: "enter", Description: "start"},
			{Key: "l", Description: "log"},
			{Key: "q", Description: "quit"},
		}),
	)
}

// Reset clears the text input and re-focuses it.
func (v *IdleView) Reset() {
	v.input.Reset()
	v.input.Focus()
}
