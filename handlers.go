package main

import (
	"fmt"
	"time"
)

func handleWork(task string) error {
	if len(task) == 0 {
		return fmt.Errorf("task description cannot be empty")
	}

	if len(task) > 100 {
		return fmt.Errorf("task description cannot exceed 100 characters")
	}

	state, err := loadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.CurrentSession != nil {
		return fmt.Errorf("session already running: %s", state.CurrentSession.Task)
	}

	now := time.Now()
	state.CurrentSession = &CurrentSession{
		ID:        newSessionID(),
		Task:      task,
		StartTime: now,
		State:     StateWorking,
	}

	if err := saveState(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	fmt.Printf("Started: %s at %s\n", task, now.Format("15:04"))
	return nil
}

func handleBreak() error {
	state, err := loadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.CurrentSession == nil {
		return fmt.Errorf("no active work session")
	}

	if state.CurrentSession.State != StateWorking {
		return fmt.Errorf("not currently working")
	}

	now := time.Now()
	workDuration := now.Sub(state.CurrentSession.StartTime)
	suggestedBreakMinutes := calculateBreakMinutes(workDuration)

	state.CurrentSession.State = StateBreaking
	state.CurrentBreak = &CurrentBreak{
		StartTime:         now,
		SuggestedDuration: suggestedBreakMinutes,
	}

	if err := saveState(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	fmt.Printf("Work session: %s. Starting %d minute break.\n",
		formatDurationShort(workDuration), suggestedBreakMinutes)
	return nil
}

func handleResume() error {
	state, err := loadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.CurrentSession == nil || state.CurrentBreak == nil {
		return fmt.Errorf("no active break session")
	}

	if state.CurrentSession.State != StateBreaking {
		return fmt.Errorf("not currently on break")
	}

	now := time.Now()
	breakDuration := now.Sub(state.CurrentBreak.StartTime)
	actualBreakMinutes := int(breakDuration.Minutes())

	workDuration := state.CurrentBreak.StartTime.Sub(state.CurrentSession.StartTime)

	completedSession := CompletedSession{
		ID:            state.CurrentSession.ID,
		Task:          state.CurrentSession.Task,
		WorkDuration:  int(workDuration.Minutes()),
		BreakDuration: &actualBreakMinutes,
		CompletedAt:   now,
	}

	state.CompletedSessions = append(state.CompletedSessions, completedSession)
	state.CurrentSession = nil
	state.CurrentBreak = nil

	if err := saveState(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	fmt.Printf("Break ended. Ready for next session.\n")
	return nil
}

func handleStop() error {
	state, err := loadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.CurrentSession == nil {
		return fmt.Errorf("no active session")
	}

	now := time.Now()

	if state.CurrentSession.State == StateWorking {
		workDuration := now.Sub(state.CurrentSession.StartTime)
		completedSession := CompletedSession{
			ID:           state.CurrentSession.ID,
			Task:         state.CurrentSession.Task,
			WorkDuration: int(workDuration.Minutes()),
			CompletedAt:  now,
		}
		state.CompletedSessions = append(state.CompletedSessions, completedSession)
	} else if state.CurrentSession.State == StateBreaking && state.CurrentBreak != nil {
		breakDuration := now.Sub(state.CurrentBreak.StartTime)
		actualBreakMinutes := int(breakDuration.Minutes())
		workDuration := state.CurrentBreak.StartTime.Sub(state.CurrentSession.StartTime)

		completedSession := CompletedSession{
			ID:            state.CurrentSession.ID,
			Task:          state.CurrentSession.Task,
			WorkDuration:  int(workDuration.Minutes()),
			BreakDuration: &actualBreakMinutes,
			CompletedAt:   now,
		}
		state.CompletedSessions = append(state.CompletedSessions, completedSession)
	}

	state.CurrentSession = nil
	state.CurrentBreak = nil

	if err := saveState(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	fmt.Println("Session ended.")
	return nil
}

func handleStatus() error {
	state, err := loadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.CurrentSession == nil {
		fmt.Println("No active session")
		return nil
	}

	now := time.Now()

	switch state.CurrentSession.State {
	case StateWorking:
		workDuration := now.Sub(state.CurrentSession.StartTime)
		fmt.Printf("Working on '%s' for %s\n",
			state.CurrentSession.Task,
			formatDuration(workDuration))

	case StateBreaking:
		if state.CurrentBreak == nil {
			return fmt.Errorf("invalid state: breaking but no break data")
		}

		breakDuration := now.Sub(state.CurrentBreak.StartTime)
		suggestedDuration := time.Duration(state.CurrentBreak.SuggestedDuration) * time.Minute
		remaining := suggestedDuration - breakDuration

		if remaining > 0 {
			fmt.Printf("Break: %s remaining\n", formatDuration(remaining))
		} else {
			elapsed := breakDuration - suggestedDuration
			fmt.Printf("Break: %s overtime\n", formatDuration(elapsed))
		}

	default:
		fmt.Printf("Session state: %s\n", state.CurrentSession.State)
	}

	return nil
}

func handleLog() error {
	state, err := loadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if len(state.CompletedSessions) == 0 {
		fmt.Println("No completed sessions")
		return nil
	}

	maxSessions := 10
	sessions := state.CompletedSessions
	if len(sessions) > maxSessions {
		sessions = sessions[len(sessions)-maxSessions:]
	}

	fmt.Println("Recent sessions:")
	for i := len(sessions) - 1; i >= 0; i-- {
		session := sessions[i]
		workDur := time.Duration(session.WorkDuration) * time.Minute

		breakInfo := "no break"
		if session.BreakDuration != nil {
			breakDur := time.Duration(*session.BreakDuration) * time.Minute
			breakInfo = fmt.Sprintf("break %s", formatDurationShort(breakDur))
		}

		fmt.Printf("  %s - %s (work %s, %s)\n",
			session.CompletedAt.Format("Jan 2 15:04"),
			session.Task,
			formatDurationShort(workDur),
			breakInfo)
	}

	return nil
}
