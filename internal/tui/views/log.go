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

// LogView displays a paginated table of completed sessions (newest first)
// with cursor-based row selection for deletion.
type LogView struct {
	sessions []flowtime.CompletedSession
	page     int
	pageSize int
	cursor   int // selected row on the current page (0-indexed)
}

// NewLogView creates a LogView with the given page size.
func NewLogView(pageSize int) *LogView {
	return &LogView{
		page:     1,
		pageSize: pageSize,
	}
}

// SetSessions updates the session data and resets to page 1 with cursor at top.
func (v *LogView) SetSessions(sessions []flowtime.CompletedSession) {
	v.sessions = sessions
	v.page = 1
	v.cursor = 0
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

// pageLen returns the number of items on the current page.
func (v *LogView) pageLen() int {
	page := paginate.ReversePaginate(v.sessions, v.page, v.pageSize)
	return len(page)
}

// activeIndex maps the current (page, cursor) to an index in the active sessions slice.
func (v *LogView) activeIndex() int {
	return len(v.sessions) - (v.page-1)*v.pageSize - v.cursor - 1
}

// Update handles cursor movement, pagination, and delete keys.
func (v *LogView) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if v.cursor > 0 {
				v.cursor--
			} else if v.page > 1 {
				v.page--
				v.cursor = v.pageLen() - 1
			}
		case "down", "j":
			if v.cursor < v.pageLen()-1 {
				v.cursor++
			} else if v.page < v.totalPages() {
				v.page++
				v.cursor = 0
			}
		case "d":
			if len(v.sessions) > 0 {
				idx := v.activeIndex()
				return func() tea.Msg { return msgs.RequestDeleteSessionMsg{ActiveIndex: idx} }
			}
		case "D":
			if len(v.sessions) > 0 {
				return func() tea.Msg {
					return msgs.RequestConfirmMsg{
						Action: msgs.ConfirmAction{
							Prompt: fmt.Sprintf("Delete all %d sessions?", len(v.sessions)),
							OnYes:  msgs.DeleteAllSessionsMsg{},
						},
					}
				}
			}
		case "esc":
			return func() tea.Msg { return msgs.BackMsg{} }
		case "q":
			return tea.Quit
		}
	}
	return nil
}

// View renders the session log table with cursor highlighting.
func (v *LogView) View() string {
	title := styles.Title.Render("📜 Session Log")

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

	// Clamp cursor if sessions were deleted and page shrunk.
	if v.cursor >= len(page) {
		v.cursor = max(0, len(page)-1)
	}

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

	selectedRow := v.cursor

	t := table.New().
		Border(lipgloss.NormalBorder()).
		Headers("COMPLETED AT", "TASK", "FLOW", "BREAK").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return styles.TableHeader
			}
			// row is 0-indexed for data rows in lipgloss/table
			if row == selectedRow {
				return styles.SelectedRow
			}
			return lipgloss.NewStyle()
		})

	pageInfo := fmt.Sprintf("Page %d of %d", v.page, v.totalPages())
	tableRendered := t.Render()
	helpBar := RenderHelpBar([]KeyBinding{
		{Key: "esc", Description: "back"},
		{Key: "j/k", Description: "navigate"},
		{Key: "d", Description: "delete"},
		{Key: "D", Description: "delete all"},
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
