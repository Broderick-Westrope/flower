package views

import (
	"time"

	"github.com/Broderick-Westrope/flower/pkg/core"
	"github.com/Broderick-Westrope/flower/pkg/tui/shared"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BreakView struct {
	b *core.CurrentBreak

	progress progress.Model
}

func NewBreakView(b *core.CurrentBreak) *BreakView {
	return &BreakView{
		b:        b,
		progress: progress.New(progress.WithDefaultGradient()),
	}
}

func (v *BreakView) Update(msg tea.Msg) tea.Cmd {
	if v.b == nil {
		return nil
	}

	switch msg := msg.(type) {
	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := v.progress.Update(msg)
		v.progress = progressModel.(progress.Model)
		return cmd

	case shared.TickMsg:
		elapsed := time.Since(v.b.StartTime)
		pct := min(float64(elapsed)/float64(v.b.SuggestedDuration), 1.0)

		return v.progress.SetPercent(pct)
	}

	return nil
}

func (v *BreakView) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		"Started At: "+core.FormatHumanDateTime(v.b.StartTime),
		"Suggested Duration: "+core.FormatDuration(v.b.SuggestedDuration),
		"Elapsed: "+core.FormatDuration(time.Since(v.b.StartTime)),
		v.progress.View(),
	)
}

func (v *BreakView) SetBreak(b *core.CurrentBreak) {
	v.b = b
}
