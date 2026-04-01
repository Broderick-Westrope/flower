package flowtime

import (
	"errors"
	"fmt"
	"time"
)

// Session represents an active flow session.
type Session struct {
	Task      string
	StartTime time.Time
}

// Break represents an active break period.
type Break struct {
	StartTime         time.Time
	SuggestedDuration time.Duration
}

// CompletedSession represents a finished flow session with its recorded durations.
type CompletedSession struct {
	Task          string
	FlowDuration  time.Duration
	BreakDuration *time.Duration
	CompletedAt   time.Time
}

// FlowState holds the full state of the flowtime timer.
type FlowState struct {
	Clock             Clock
	CurrentSession    *Session
	CurrentBreak      *Break
	CompletedSessions []CompletedSession
}

// NewFlowState creates a new empty FlowState using the provided clock.
func NewFlowState(clock Clock) *FlowState {
	return &FlowState{
		Clock:             clock,
		CompletedSessions: []CompletedSession{},
	}
}

// StartSession begins a new flow session with the given task name.
// Returns an error if a session is already active, the task is empty, or the task exceeds 100 characters.
func (s *FlowState) StartSession(task string) error {
	if task == "" {
		return errors.New("task must not be empty")
	}
	if len(task) > 100 {
		return fmt.Errorf("task must be 100 characters or less, got %d", len(task))
	}
	if s.CurrentSession != nil {
		return errors.New("session already active")
	}

	s.CurrentSession = &Session{
		Task:      task,
		StartTime: s.Clock.Now(),
	}
	return nil
}

// TakeBreak pauses the current flow session and starts a break.
// Returns an error if no session is active or if already on a break.
func (s *FlowState) TakeBreak() error {
	if s.CurrentSession == nil {
		return errors.New("no active session")
	}
	if s.CurrentBreak != nil {
		return errors.New("already on break")
	}

	now := s.Clock.Now()
	workDuration := now.Sub(s.CurrentSession.StartTime)
	suggestedBreak := CalculateBreak(workDuration)

	s.CurrentBreak = &Break{
		StartTime:         now,
		SuggestedDuration: suggestedBreak,
	}
	return nil
}

// Resume returns to a flow session. If currently on a break, the current session is completed
// and a new session is started with the same task name (returns true). If idle with completed
// sessions, a new session is started with the last completed task name (returns false).
func (s *FlowState) Resume() (resumedCurrent bool, err error) {
	// Path 1: on break with active session — complete current and start new
	if s.CurrentSession != nil && s.CurrentBreak != nil {
		now := s.Clock.Now()
		flowDuration := s.CurrentBreak.StartTime.Sub(s.CurrentSession.StartTime)
		breakDuration := now.Sub(s.CurrentBreak.StartTime)

		completed := CompletedSession{
			Task:          s.CurrentSession.Task,
			FlowDuration:  flowDuration,
			BreakDuration: &breakDuration,
			CompletedAt:   now,
		}
		s.CompletedSessions = append(s.CompletedSessions, completed)

		s.CurrentBreak = nil
		s.CurrentSession = &Session{
			Task:      completed.Task,
			StartTime: now,
		}
		return true, nil
	}

	// Path 2: idle with history — start new session with last task
	if s.CurrentSession == nil && s.CurrentBreak == nil && len(s.CompletedSessions) > 0 {
		lastTask := s.CompletedSessions[len(s.CompletedSessions)-1].Task
		s.CurrentSession = &Session{
			Task:      lastTask,
			StartTime: s.Clock.Now(),
		}
		return false, nil
	}

	// Path 3: already flowing (session active, no break)
	if s.CurrentSession != nil && s.CurrentBreak == nil {
		return false, errors.New("already in flow state")
	}

	// Path 4: no session and no history
	return false, errors.New("no session to resume")
}

// Stop ends the current session and returns the completed session.
// Returns an error if no session is active.
func (s *FlowState) Stop() (*CompletedSession, error) {
	if s.CurrentSession == nil {
		return nil, errors.New("no active session")
	}

	now := s.Clock.Now()
	var completed CompletedSession

	if s.CurrentBreak != nil {
		// On break: flow = break start - session start, break = now - break start
		flowDuration := s.CurrentBreak.StartTime.Sub(s.CurrentSession.StartTime)
		breakDuration := now.Sub(s.CurrentBreak.StartTime)
		completed = CompletedSession{
			Task:          s.CurrentSession.Task,
			FlowDuration:  flowDuration,
			BreakDuration: &breakDuration,
			CompletedAt:   now,
		}
	} else {
		// Flowing: flow = now - session start, no break
		flowDuration := now.Sub(s.CurrentSession.StartTime)
		completed = CompletedSession{
			Task:         s.CurrentSession.Task,
			FlowDuration: flowDuration,
			CompletedAt:  now,
		}
	}

	s.CompletedSessions = append(s.CompletedSessions, completed)
	s.CurrentSession = nil
	s.CurrentBreak = nil

	return &completed, nil
}
