package views

import (
	"fmt"
	"time"

	"github.com/Broderick-Westrope/flower/internal/flowtime"
	"github.com/Broderick-Westrope/flower/internal/tui/msgs"
	"github.com/Broderick-Westrope/flower/internal/tui/styles"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BreakView displays a break countdown with a progress bar.
type BreakView struct {
	taskName string
	brk      *flowtime.Break
	progress progress.Model
}

// NewBreakView creates a BreakView with a default gradient progress bar.
func NewBreakView() *BreakView {
	return &BreakView{
		progress: progress.New(progress.WithDefaultGradient()),
	}
}

// SetBreak updates the break and task name displayed by this view.
func (v *BreakView) SetBreak(taskName string, b *flowtime.Break) {
	v.taskName = taskName
	v.brk = b
}

// Update handles tick and progress-frame messages.
func (v *BreakView) Update(msg tea.Msg) tea.Cmd {
	if v.brk == nil {
		return nil
	}

	switch msg := msg.(type) {
	case progress.FrameMsg:
		progressModel, cmd := v.progress.Update(msg)
		v.progress = progressModel.(progress.Model)
		return cmd

	case msgs.TickMsg:
		elapsed := time.Since(v.brk.StartTime)
		pct := float64(elapsed) / float64(v.brk.SuggestedDuration)
		if pct > 1.0 {
			pct = 1.0
		}
		return v.progress.SetPercent(pct)
	}

	return nil
}

// View renders the break countdown screen.
func (v *BreakView) View() string {
	if v.brk == nil {
		return ""
	}

	elapsed := time.Since(v.brk.StartTime)
	suggested := v.brk.SuggestedDuration

	elapsedStr := flowtime.FormatDuration(elapsed)
	suggestedStr := flowtime.FormatDuration(suggested)

	var statusStr string
	if elapsed > suggested {
		overtime := elapsed - suggested
		statusStr = fmt.Sprintf("(%s overtime)", flowtime.FormatDuration(overtime))
	} else {
		remaining := suggested - elapsed
		statusStr = fmt.Sprintf("(%s remaining)", flowtime.FormatDuration(remaining))
	}

	timerLine := styles.Timer.Render(fmt.Sprintf("%s / %s", elapsedStr, suggestedStr)) +
		"  " + statusStr
	taskLine := styles.TaskName.Render(v.taskName)
	progressLine := v.progress.View()
	helpBar := RenderHelpBar([]KeyBinding{
		{Key: "space", Description: "resume"},
		{Key: "s", Description: "stop"},
		{Key: "c", Description: "cancel"},
		{Key: "l", Description: "log"},
		{Key: "q", Description: "quit"},
	})

	contentWidth := max(
		lipgloss.Width(timerLine),
		lipgloss.Width(taskLine),
		lipgloss.Width(progressLine),
		lipgloss.Width(helpBar),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		styles.Title.Render("🧘 Break"),
		"",
		taskLine,
		timerLine,
		progressLine,
		"",
		styles.Separator(contentWidth),
		helpBar,
	)
}
