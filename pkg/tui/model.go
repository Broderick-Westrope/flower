package tui

import (
	"fmt"

	"github.com/Broderick-Westrope/flower/pkg/core"
	"github.com/Broderick-Westrope/flower/pkg/tui/shared"
	"github.com/Broderick-Westrope/flower/pkg/tui/views"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ tea.Model = &Model{}

type Model struct {
	state *core.AppState

	taskView  *views.TaskView
	breakView *views.BreakView
}

func New(state *core.AppState) (*Model, error) {
	if state == nil {
		var err error
		state, err = core.LoadState()
		if err != nil {
			return nil, fmt.Errorf("loading state: %w", err)
		}
	}

	return &Model{
		state:     state,
		taskView:  views.NewTaskView(state.CurrentSession),
		breakView: views.NewBreakView(state.CurrentBreak),
	}, nil
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		shared.Tick(),
		m.taskView.Init(),
	)
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case " ":
			if m.state.CurrentBreak != nil {
				_, err := m.state.ResumeCurrentOrPreviousSession()
				if err != nil {
					panic(fmt.Sprintf("failed to resume current session: %v", err))
				}
			} else if m.state.CurrentSession != nil {
				err := m.state.TakeBreak()
				if err != nil {
					panic(fmt.Sprintf("failed to take break: %v", err))
				}
			}

			m.taskView.SetSession(m.state.CurrentSession)
			m.breakView.SetBreak(m.state.CurrentBreak)
			return m, nil
		}

	case shared.TickMsg:
		cmds = append(cmds, shared.Tick())
	}

	return m, tea.Batch(
		m.taskView.Update(msg),
		m.breakView.Update(msg),
		tea.Batch(cmds...),
	)
}

// View implements tea.Model.
func (m *Model) View() string {
	var output string
	switch {
	case m.state.CurrentBreak != nil:
		output = m.breakView.View()
	case m.state.CurrentSession != nil:
		output = m.taskView.View()
	default:
		output = "NO SESSION FOUND"
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		Padding(0, 1).
		Margin(1, 2).
		Render(output)
}
