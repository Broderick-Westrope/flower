package tui

import (
	"fmt"
	"time"

	"github.com/Broderick-Westrope/flower/internal/flowtime"
	"github.com/Broderick-Westrope/flower/internal/storage"
	"github.com/Broderick-Westrope/flower/internal/tui/msgs"
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

	// Confirmation prompt state.
	confirming    bool
	confirmAction msgs.ConfirmAction

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
		// When a confirmation prompt is active, only handle y/n/esc.
		if m.confirming {
			return m.handleConfirmKey(msg)
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

	case RequestConfirmMsg:
		m.confirming = true
		m.confirmAction = msg.Action
		return m, nil

	case CancelSessionMsg:
		return m.handleCancelSession()

	case DeleteSessionMsg:
		return m.handleDeleteSession(msg.Index)

	case DeleteAllSessionsMsg:
		return m.handleDeleteAllSessions()

	case RequestDeleteSessionMsg:
		return m.handleRequestDeleteSession(msg.ActiveIndex)
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

	if m.confirming {
		content = lipgloss.JoinVertical(lipgloss.Left,
			content,
			"",
			styles.ConfirmPrompt.Render(m.confirmAction.Prompt+" [y/n]"),
		)
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
		case "c":
			return m.requestConfirm("Cancel session?", CancelSessionMsg{})
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
		case "c":
			return m.requestConfirm("Cancel session?", CancelSessionMsg{})
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
	return m, m.flowView.Init()
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
	return m, m.flowView.Init()
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
	return m, m.idleView.Init()
}

func (m *Model) handleShowLog() (tea.Model, tea.Cmd) {
	m.logView.SetSessions(m.state.ActiveSessions())
	m.activeView = viewLog
	return m, nil
}

func (m *Model) handleBack() (tea.Model, tea.Cmd) {
	switch {
	case m.state.CurrentSession != nil && m.state.CurrentBreak != nil:
		m.activeView = viewBreak
	case m.state.CurrentSession != nil:
		m.activeView = viewFlow
		return m, m.flowView.Init()
	default:
		m.activeView = viewIdle
		return m, m.idleView.Init()
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

func (m *Model) requestConfirm(prompt string, onYes tea.Msg) (tea.Model, tea.Cmd) {
	m.confirming = true
	m.confirmAction = msgs.ConfirmAction{Prompt: prompt, OnYes: onYes}
	return m, nil
}

func (m *Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.confirming = false
		action := m.confirmAction.OnYes
		m.confirmAction = msgs.ConfirmAction{}
		return m.Update(action)
	case "n", "N", "esc":
		m.confirming = false
		m.confirmAction = msgs.ConfirmAction{}
		return m, nil
	}
	return m, nil
}

func (m *Model) handleCancelSession() (tea.Model, tea.Cmd) {
	if err := m.state.CancelSession(); err != nil {
		return m, errCmd(err)
	}
	if err := m.store.Save(m.state); err != nil {
		return m, errCmd(fmt.Errorf("saving: %w", err))
	}

	m.activeView = viewIdle
	m.idleView.Reset()
	return m, m.idleView.Init()
}

func (m *Model) handleDeleteSession(index int) (tea.Model, tea.Cmd) {
	if err := m.state.DeleteSession(index); err != nil {
		return m, errCmd(err)
	}
	if err := m.store.Save(m.state); err != nil {
		return m, errCmd(fmt.Errorf("saving: %w", err))
	}

	// Refresh the log view with updated active sessions.
	m.logView.SetSessions(m.state.ActiveSessions())
	return m, nil
}

func (m *Model) handleRequestDeleteSession(activeIndex int) (tea.Model, tea.Cmd) {
	active := m.state.ActiveSessions()
	if activeIndex < 0 || activeIndex >= len(active) {
		return m, errCmd(fmt.Errorf("session index %d out of range", activeIndex))
	}

	target := active[activeIndex]

	// Find the matching entry in the full slice.
	fullIndex := -1
	for i, cs := range m.state.CompletedSessions {
		if cs.CompletedAt.Equal(target.CompletedAt) && cs.Task == target.Task && cs.DeletedAt == nil {
			fullIndex = i
			break
		}
	}
	if fullIndex == -1 {
		return m, errCmd(fmt.Errorf("session not found"))
	}

	return m.requestConfirm(
		fmt.Sprintf("Delete %q?", target.Task),
		DeleteSessionMsg{Index: fullIndex},
	)
}

func (m *Model) handleDeleteAllSessions() (tea.Model, tea.Cmd) {
	if err := m.state.DeleteAllSessions(); err != nil {
		return m, errCmd(err)
	}
	if err := m.store.Save(m.state); err != nil {
		return m, errCmd(fmt.Errorf("saving: %w", err))
	}

	// Refresh the log view with updated active sessions.
	m.logView.SetSessions(m.state.ActiveSessions())
	return m, nil
}

func errCmd(err error) tea.Cmd {
	return func() tea.Msg { return ErrorMsg{Err: err} }
}
