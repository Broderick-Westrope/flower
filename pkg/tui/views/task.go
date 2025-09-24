package views

import (
	"time"

	"github.com/Broderick-Westrope/flower/pkg/core"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TaskView struct {
	session *core.CurrentSession

	spinner spinner.Model
}

func NewTaskView(session *core.CurrentSession) *TaskView {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	return &TaskView{
		session: session,
		spinner: s,
	}
}

func (v *TaskView) Init() tea.Cmd {
	return v.spinner.Tick
}

func (v *TaskView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	v.spinner, cmd = v.spinner.Update(msg)
	return cmd
}

func (v *TaskView) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		"Task: \""+v.session.Task+"\"",
		"Started At: "+core.FormatHumanDateTime(v.session.StartTime),
		"Elapsed: "+core.FormatDuration(time.Since(v.session.StartTime))+" "+v.spinner.View(),
	)
}

func (v *TaskView) SetSession(s *core.CurrentSession) {
	v.session = s
}
