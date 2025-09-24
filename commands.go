package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Broderick-Westrope/flower/pkg/core"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type StartCmd struct {
	Task string `arg:"" help:"Task description"`

	Detach bool `short:"d"`
}

func (cmd *StartCmd) Run() error {
	cmd.Task = strings.TrimSuffix(strings.TrimPrefix(strconv.Quote(cmd.Task), "\""), "\"")

	if len(cmd.Task) == 0 {
		return fmt.Errorf("task cannot be empty")
	}

	if len(cmd.Task) > 100 {
		return fmt.Errorf("task cannot exceed 100 characters")
	}

	state, err := core.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.CurrentSession != nil {
		return fmt.Errorf("session already running: %s", state.CurrentSession.Task)
	}

	startedAt := time.Now()
	state.CurrentSession = &core.CurrentSession{
		Task:      cmd.Task,
		StartTime: startedAt,
	}

	err = core.SaveState(state)
	if err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	if cmd.Detach {
		fmt.Printf("Started: %s at %s\n", cmd.Task, startedAt.Format("15:04"))
		return nil
	}

	err = startTUI(state)
	if err != nil {
		return fmt.Errorf("starting TUI: %w", err)
	}
	return nil
}

type BreakCmd struct {
	Detach bool `short:"d"`
}

func (cmd *BreakCmd) Run() error {
	state, err := core.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.CurrentSession == nil {
		return fmt.Errorf("no active work session")
	}

	if state.CurrentBreak != nil {
		return fmt.Errorf("not currently working")
	}

	err = state.TakeBreak()
	if err != nil {
		return fmt.Errorf("taking break: %w", err)
	}

	if cmd.Detach {
		fmt.Printf("Flow ended. Starting %s break.\n", core.FormatDuration(state.CurrentBreak.SuggestedDuration))
		return nil
	}

	err = startTUI(state)
	if err != nil {
		return fmt.Errorf("starting TUI: %w", err)
	}
	return nil
}

type ResumeCmd struct {
	Detach bool `short:"d"`
}

func (cmd *ResumeCmd) Run() error {
	state, err := core.LoadState()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	if state.CurrentSession != nil && state.CurrentBreak == nil {
		if cmd.Detach {
			return fmt.Errorf("cannot resume; already working")
		}

		err = startTUI(state)
		if err != nil {
			return fmt.Errorf("starting TUI: %w", err)
		}
		return nil
	}

	resumedCurrent, err := state.ResumeCurrentOrPreviousSession()
	if err != nil {
		if resumedCurrent {
			return fmt.Errorf("resuming current session: %w", err)
		}
		return fmt.Errorf("resuming previous session: %w", err)
	}

	if cmd.Detach {
		s := "Previous"
		if resumedCurrent {
			s = "Current"
		}

		fmt.Printf("Break ended. %s session resumed.\n", s)
		return nil
	}

	err = startTUI(state)
	if err != nil {
		return fmt.Errorf("starting TUI: %w", err)
	}
	return nil
}

type StopCmd struct{}

func (cmd *StopCmd) Run() error {
	state, err := core.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.CurrentSession == nil {
		return fmt.Errorf("no active session")
	}

	now := time.Now()
	workDuration := now.Sub(state.CurrentSession.StartTime)
	var breakDuration *time.Duration

	if state.CurrentBreak != nil {
		workDuration = state.CurrentBreak.StartTime.Sub(state.CurrentSession.StartTime)
		temp := now.Sub(state.CurrentBreak.StartTime)
		breakDuration = &temp
	}

	completedSession := core.CompletedSession{
		Task:          state.CurrentSession.Task,
		FlowDuration:  workDuration,
		BreakDuration: breakDuration,
		CompletedAt:   now,
	}
	state.CompletedSessions = append(state.CompletedSessions, completedSession)

	state.CurrentSession = nil
	state.CurrentBreak = nil

	if err := core.SaveState(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	fmt.Println("Session ended.")
	return nil
}

type StatusCmd struct{}

func (cmd *StatusCmd) Run() error {
	state, err := core.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.CurrentSession == nil {
		fmt.Println("No active session")
		return nil
	}

	now := time.Now()
	if state.CurrentBreak == nil {
		workDuration := now.Sub(state.CurrentSession.StartTime)
		fmt.Printf("Working on '%s' for %s\n",
			state.CurrentSession.Task,
			core.FormatDuration(workDuration))
	} else {
		breakDuration := now.Sub(state.CurrentBreak.StartTime)
		suggestedDuration := time.Duration(state.CurrentBreak.SuggestedDuration) * time.Minute
		remaining := suggestedDuration - breakDuration

		if remaining > 0 {
			fmt.Printf("Break: %s remaining\n", core.FormatDuration(remaining))
		} else {
			elapsed := breakDuration - suggestedDuration
			fmt.Printf("Break: %s overtime\n", core.FormatDuration(elapsed))
		}
	}

	return nil
}

type LogCmd struct {
	Count int `default:"10" help:"Entries per page"`
	Page  int `default:"1" help:"Page to display"`
}

func (cmd *LogCmd) Run() error {
	if cmd.Count <= 0 {
		return errors.New("count must be greater than zero")
	}
	if cmd.Page <= 0 {
		return errors.New("page must be greater than zero")
	}

	state, err := core.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if len(state.CompletedSessions) == 0 {
		fmt.Println("No completed sessions")
		return nil
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

	sessions := ReversePaginate(state.CompletedSessions, cmd.Page, cmd.Count)
	for _, session := range sessions {
		breakInfo := "none"
		if session.BreakDuration != nil {
			breakInfo = core.FormatDuration(*session.BreakDuration)
		}

		t.Row(
			core.FormatHumanDateTime(session.CompletedAt),
			session.Task,
			core.FormatDuration(session.FlowDuration),
			breakInfo,
		)
	}

	fmt.Printf("Recent sessions:\n%s\n", t.Render())
	return nil
}

type LocateCmd struct{}

func (cmd *LocateCmd) Run() error {
	fp, err := core.GetStateFilePath()
	if err != nil {
		return fmt.Errorf("getting state file path: %w", err)
	}

	fmt.Println(fp)
	return nil
}
