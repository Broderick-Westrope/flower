package views

import (
	"fmt"
	"time"

	"github.com/Broderick-Westrope/flower/internal/flowtime"
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

// reversePaginate returns sessions for the current page in newest-first order.
func (v *LogView) reversePaginate() []flowtime.CompletedSession {
	total := len(v.sessions)
	if total == 0 {
		return nil
	}

	// Sessions are stored oldest-first; we want newest-first pages.
	// Page 1 shows the last pageSize items (reversed), page 2 the next batch, etc.
	endIdx := total - (v.page-1)*v.pageSize
	startIdx := endIdx - v.pageSize
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx < 0 {
		return nil
	}

	slice := v.sessions[startIdx:endIdx]
	// Reverse to newest-first.
	result := make([]flowtime.CompletedSession, len(slice))
	for i, s := range slice {
		result[len(slice)-1-i] = s
	}
	return result
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
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			"No completed sessions yet.",
			"",
			styles.Separator(20),
			RenderHelpBar([]KeyBinding{
				{Key: "esc", Description: "back"},
				{Key: "q", Description: "quit"},
			}),
		)
	}

	now := time.Now()
	page := v.reversePaginate()
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

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		t.Render(),
		"",
		pageInfo,
		"",
		styles.Separator(20),
		RenderHelpBar([]KeyBinding{
			{Key: "esc", Description: "back"},
			{Key: "↑/↓", Description: "page"},
			{Key: "q", Description: "quit"},
		}),
	)
}
