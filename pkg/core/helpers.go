package core

import (
	"fmt"
	"time"
)

func (s *AppState) TakeBreak() error {
	now := time.Now()
	workDuration := now.Sub(s.CurrentSession.StartTime)
	suggestedBreak := CalculateBreak(workDuration)

	s.CurrentBreak = &CurrentBreak{
		StartTime:         now,
		SuggestedDuration: suggestedBreak,
	}

	err := SaveState(s)
	if err != nil {
		return fmt.Errorf("saving state: %w", err)
	}
	return nil
}

// ResumeCurrentOrPreviousSession will complete the current session and begin a new one with the same name.
// If there is no current session or no current break then the name of the last completed session (ie. a previous session) will be used instead.
// After modifying the state it will be saved.
// The bool return value is true if the current session was resumed and false if a previous session was resumed.
func (s *AppState) ResumeCurrentOrPreviousSession() (bool, error) {
	resumedCurrent := false
	if s.CurrentSession == nil && s.CurrentBreak != nil {
		s.resumePreviousSession()
	} else {
		s.resumeCurrentSession()
		resumedCurrent = true
	}

	err := SaveState(s)
	if err != nil {
		return resumedCurrent, fmt.Errorf("saving state: %w", err)
	}
	return resumedCurrent, nil
}

// resumePreviousSession creates a new session with the same name as the last completed session.
// It also resets the current break.
func (s *AppState) resumePreviousSession() {
	now := time.Now()
	prevSession := s.CompletedSessions[len(s.CompletedSessions)-1]

	s.CurrentBreak = nil
	s.CurrentSession = &CurrentSession{
		Task:      prevSession.Task,
		StartTime: now,
	}
}

// resumeCurrentSession creates a new session with the same name as the current session.
// It also adds the current session to the list of completed sessions and resets the current break.
func (s *AppState) resumeCurrentSession() {
	now := time.Now()
	flowDuration := s.CurrentBreak.StartTime.Sub(s.CurrentSession.StartTime)
	breakDuration := now.Sub(s.CurrentBreak.StartTime)

	completedSession := CompletedSession{
		Task:          s.CurrentSession.Task,
		FlowDuration:  flowDuration,
		BreakDuration: &breakDuration,
		CompletedAt:   now,
	}
	s.CompletedSessions = append(s.CompletedSessions, completedSession)

	s.CurrentBreak = nil
	s.CurrentSession = &CurrentSession{
		Task:      completedSession.Task,
		StartTime: now,
	}
}
