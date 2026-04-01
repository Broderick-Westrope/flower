package tui

import (
	"fmt"
	"time"

	"github.com/Broderick-Westrope/flower/internal/flowtime"
	"github.com/Broderick-Westrope/flower/internal/storage"
	"github.com/Broderick-Westrope/flower/internal/tui/styles"
	"github.com/Broderick-Westrope/flower/internal/tui/views"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewKind int

const (
	viewIdle viewKind = iota
	viewFlow
	viewBreak
	viewLog
)

// Model is the top-level Bubble Tea model for the flower TUI.
type Model struct {
	store      storage.Store
	state      *flowtime.FlowState
	activeView viewKind

	idleView  *views.IdleView
	flowView  *views.FlowView
	breakView *views.BreakView
	logView   *views.LogView

	err         error
	errDeadline time.Time
	width       int
	height      int
}

var _ tea.Model = (*Model)(nil)

const logPageSize = 10

// New creates a Model, loading persisted state from the store.
func New(store storage.Store) (*Model, error) {
	state, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("loading state: %w", err)
	}

	m := &Model{
		store:     store,
		state:     state,
		idleView:  views.NewIdleView(),
		flowView:  views.NewFlowView(),
		breakView: views.NewBreakView(),
		logView:   views.NewLogView(logPageSize),
	}

	// Determine initial view from restored state.
	switch {
	case state.CurrentSession != nil && state.CurrentBreak != nil:
		m.activeView = viewBreak
		m.flowView.SetSession(state.CurrentSession)
		m.breakView.SetBreak(state.CurrentSession.Task, state.CurrentBreak)
	case state.CurrentSession != nil:
		m.activeView = viewFlow
		m.flowView.SetSession(state.CurrentSession)
	default:
		m.activeView = viewIdle
	}

	return m, nil
}

// Init starts the tick loop plus sub-component initialisation.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		Tick(),
		m.idleView.Init(),
		m.flowView.Init(),
	)
}

// Update handles all messages for the TUI.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		m.errDeadline = time.Now().Add(5 * time.Second)
		return m, nil

	case TickMsg:
		return m.handleTick()

	case StartSessionMsg:
		return m.handleStartSession(msg.Task)

	case ShowLogMsg:
		return m.handleShowLog()

	case BackMsg:
		return m.handleBack()
	}

	// Delegate spinner, progress frames, etc. to active view.
	return m, m.delegateToActiveView(msg)
}

// View renders the active view wrapped in the container style.
func (m *Model) View() string {
	var content string
	switch m.activeView {
	case viewIdle:
		content = m.idleView.View()
	case viewFlow:
		content = m.flowView.View()
	case viewBreak:
		content = m.breakView.View()
	case viewLog:
		content = m.logView.View()
	}

	if m.err != nil {
		content = lipgloss.JoinVertical(lipgloss.Left,
			content,
			"",
			styles.ErrorText.Render("Error: "+m.err.Error()),
		)
	}

	return styles.Container.Render(content)
}

// --- private helpers ---

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.activeView {
	case viewIdle:
		cmd := m.idleView.Update(msg)
		return m, cmd

	case viewLog:
		cmd := m.logView.Update(msg)
		return m, cmd

	case viewFlow:
		switch msg.String() {
		case " ":
			return m.handleTakeBreak()
		case "s":
			return m.handleStop()
		case "l":
			return m.handleShowLog()
		case "q":
			return m, tea.Quit
		}

	case viewBreak:
		switch msg.String() {
		case " ":
			return m.handleResume()
		case "s":
			return m.handleStop()
		case "l":
			return m.handleShowLog()
		case "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *Model) handleTick() (tea.Model, tea.Cmd) {
	if !m.errDeadline.IsZero() && time.Now().After(m.errDeadline) {
		m.err = nil
		m.errDeadline = time.Time{}
	}

	cmd := m.delegateToActiveView(TickMsg{})
	return m, tea.Batch(cmd, Tick())
}

func (m *Model) handleStartSession(task string) (tea.Model, tea.Cmd) {
	if err := m.state.StartSession(task); err != nil {
		return m, errCmd(fmt.Errorf("starting session: %w", err))
	}
	if err := m.store.Save(m.state); err != nil {
		return m, errCmd(fmt.Errorf("saving: %w", err))
	}

	m.activeView = viewFlow
	m.flowView.SetSession(m.state.CurrentSession)
	return m, nil
}

func (m *Model) handleTakeBreak() (tea.Model, tea.Cmd) {
	if err := m.state.TakeBreak(); err != nil {
		return m, errCmd(err)
	}
	if err := m.store.Save(m.state); err != nil {
		return m, errCmd(fmt.Errorf("saving: %w", err))
	}

	m.activeView = viewBreak
	m.breakView.SetBreak(m.state.CurrentSession.Task, m.state.CurrentBreak)
	return m, nil
}

func (m *Model) handleResume() (tea.Model, tea.Cmd) {
	_, err := m.state.Resume()
	if err != nil {
		return m, errCmd(err)
	}
	if err := m.store.Save(m.state); err != nil {
		return m, errCmd(fmt.Errorf("saving: %w", err))
	}

	m.activeView = viewFlow
	m.flowView.SetSession(m.state.CurrentSession)
	return m, nil
}

func (m *Model) handleStop() (tea.Model, tea.Cmd) {
	_, err := m.state.Stop()
	if err != nil {
		return m, errCmd(err)
	}
	if err := m.store.Save(m.state); err != nil {
		return m, errCmd(fmt.Errorf("saving: %w", err))
	}

	m.activeView = viewIdle
	m.idleView.Reset()
	return m, nil
}

func (m *Model) handleShowLog() (tea.Model, tea.Cmd) {
	m.logView.SetSessions(m.state.CompletedSessions)
	m.activeView = viewLog
	return m, nil
}

func (m *Model) handleBack() (tea.Model, tea.Cmd) {
	switch {
	case m.state.CurrentSession != nil && m.state.CurrentBreak != nil:
		m.activeView = viewBreak
	case m.state.CurrentSession != nil:
		m.activeView = viewFlow
	default:
		m.activeView = viewIdle
	}
	return m, nil
}

func (m *Model) delegateToActiveView(msg tea.Msg) tea.Cmd {
	switch m.activeView {
	case viewIdle:
		return m.idleView.Update(msg)
	case viewFlow:
		return m.flowView.Update(msg)
	case viewBreak:
		return m.breakView.Update(msg)
	case viewLog:
		return m.logView.Update(msg)
	}
	return nil
}

func errCmd(err error) tea.Cmd {
	return func() tea.Msg { return ErrorMsg{Err: err} }
}
