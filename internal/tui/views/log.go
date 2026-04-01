package views

import (
	"fmt"
	"time"

	"github.com/Broderick-Westrope/flower/internal/flowtime"
	"github.com/Broderick-Westrope/flower/internal/paginate"
	"github.com/Broderick-Westrope/flower/internal/tui/msgs"
	"github.com/Broderick-Westrope/flower/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// LogView displays a paginated table of completed sessions (newest first).
type LogView struct {
	sessions []flowtime.CompletedSession
	page     int
	pageSize int
}

// NewLogView creates a LogView with the given page size.
func NewLogView(pageSize int) *LogView {
	return &LogView{
		page:     1,
		pageSize: pageSize,
	}
}

// SetSessions updates the session data and resets to page 1.
func (v *LogView) SetSessions(sessions []flowtime.CompletedSession) {
	v.sessions = sessions
	v.page = 1
}

// totalPages returns the number of pages needed for all sessions.
func (v *LogView) totalPages() int {
	if len(v.sessions) == 0 {
		return 1
	}
	pages := len(v.sessions) / v.pageSize
	if len(v.sessions)%v.pageSize != 0 {
		pages++
	}
	return pages
}

// Update handles pagination keys.
func (v *LogView) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if v.page > 1 {
				v.page--
			}
		case "down", "j":
			if v.page < v.totalPages() {
				v.page++
			}
		case "esc":
			return func() tea.Msg { return msgs.BackMsg{} }
		case "q":
			return tea.Quit
		}
	}
	return nil
}

// View renders the session log table.
func (v *LogView) View() string {
	title := styles.Title.Render("Session Log")

	if len(v.sessions) == 0 {
		emptyMsg := "No completed sessions yet."
		helpBar := RenderHelpBar([]KeyBinding{
			{Key: "esc", Description: "back"},
			{Key: "q", Description: "quit"},
		})
		contentWidth := max(
			lipgloss.Width(title),
			lipgloss.Width(emptyMsg),
			lipgloss.Width(helpBar),
		)
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			emptyMsg,
			"",
			styles.Separator(contentWidth),
			helpBar,
		)
	}

	now := time.Now()
	page := paginate.ReversePaginate(v.sessions, v.page, v.pageSize)
	rows := make([][]string, len(page))
	for i, s := range page {
		breakStr := "-"
		if s.BreakDuration != nil {
			breakStr = flowtime.FormatDuration(*s.BreakDuration)
		}
		rows[i] = []string{
			flowtime.FormatHumanDateTime(s.CompletedAt, now),
			s.Task,
			flowtime.FormatDuration(s.FlowDuration),
			breakStr,
		}
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		Headers("COMPLETED AT", "TASK", "FLOW", "BREAK").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return styles.TableHeader
			}
			return lipgloss.NewStyle()
		})

	pageInfo := fmt.Sprintf("Page %d of %d", v.page, v.totalPages())
	tableRendered := t.Render()
	helpBar := RenderHelpBar([]KeyBinding{
		{Key: "esc", Description: "back"},
		{Key: "↑", Description: "newer"},
		{Key: "↓", Description: "older"},
		{Key: "q", Description: "quit"},
	})

	contentWidth := max(
		lipgloss.Width(title),
		lipgloss.Width(tableRendered),
		lipgloss.Width(helpBar),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		tableRendered,
		"",
		pageInfo,
		"",
		styles.Separator(contentWidth),
		helpBar,
	)
}
