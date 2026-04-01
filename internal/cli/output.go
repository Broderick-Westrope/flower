package cli

import (
	"fmt"
	"time"

	"github.com/Broderick-Westrope/flower/internal/flowtime"
	"github.com/Broderick-Westrope/flower/internal/paginate"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// PrintStatus prints the current flow state to stdout.
func PrintStatus(state *flowtime.FlowState, now time.Time) {
	if state.CurrentSession == nil {
		fmt.Println("No active session")
		return
	}

	if state.CurrentBreak == nil {
		workDuration := now.Sub(state.CurrentSession.StartTime)
		fmt.Printf("Working on '%s' for %s\n",
			state.CurrentSession.Task,
			flowtime.FormatDuration(workDuration))
		return
	}

	breakDuration := now.Sub(state.CurrentBreak.StartTime)
	remaining := state.CurrentBreak.SuggestedDuration - breakDuration

	if remaining > 0 {
		fmt.Printf("Break: %s remaining\n", flowtime.FormatDuration(remaining))
	} else {
		elapsed := breakDuration - state.CurrentBreak.SuggestedDuration
		fmt.Printf("Break: %s overtime\n", flowtime.FormatDuration(elapsed))
	}
}

// PrintLog prints completed sessions as a paginated table to stdout.
func PrintLog(sessions []flowtime.CompletedSession, page, count int, now time.Time) {
	if len(sessions) == 0 {
		fmt.Println("No completed sessions")
		return
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		Headers("COMPLETED AT", "TASK", "DURATION", "BREAK").
		StyleFunc(func(row, col int) lipgloss.Style {
			baseStyle := lipgloss.NewStyle().Padding(0, 1)

			if row == table.HeaderRow {
				baseStyle = baseStyle.Bold(true)
			}

			return baseStyle
		})

	paginated := paginate.ReversePaginate(sessions, page, count)
	for _, session := range paginated {
		breakInfo := "none"
		if session.BreakDuration != nil {
			breakInfo = flowtime.FormatDuration(*session.BreakDuration)
		}

		t.Row(
			flowtime.FormatHumanDateTime(session.CompletedAt, now),
			session.Task,
			flowtime.FormatDuration(session.FlowDuration),
			breakInfo,
		)
	}

	fmt.Printf("Recent sessions:\n%s\n", t.Render())
}
