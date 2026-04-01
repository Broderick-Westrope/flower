package flowtime

import (
	"errors"
	"strings"
	"testing"
	"time"
)

type mockClock struct {
	now time.Time
}

func (c *mockClock) Now() time.Time          { return c.now }
func (c *mockClock) Advance(d time.Duration) { c.now = c.now.Add(d) }

func newTestClock() *mockClock {
	return &mockClock{now: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)}
}

func TestStartSession(t *testing.T) {
	t.Run("from idle succeeds", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		err := state.StartSession("write code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.CurrentSession == nil {
			t.Fatal("expected current session to be set")
		}
		if state.CurrentSession.Task != "write code" {
			t.Errorf("task = %q, want %q", state.CurrentSession.Task, "write code")
		}
		if !state.CurrentSession.StartTime.Equal(clock.Now()) {
			t.Errorf("start time = %v, want %v", state.CurrentSession.StartTime, clock.Now())
		}
	})

	t.Run("errors when session active", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		_ = state.StartSession("task 1")
		err := state.StartSession("task 2")
		if err == nil {
			t.Fatal("expected error when session already active")
		}
		if !errors.Is(err, ErrSessionActive) {
			t.Errorf("error = %q, want %q", err.Error(), ErrSessionActive)
		}
	})

	t.Run("errors on empty task", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		err := state.StartSession("")
		if err == nil {
			t.Fatal("expected error on empty task")
		}
		if !errors.Is(err, ErrTaskEmpty) {
			t.Errorf("error = %q, want %q", err.Error(), ErrTaskEmpty)
		}
	})

	t.Run("errors on task over 100 chars", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		longTask := strings.Repeat("a", 101)
		err := state.StartSession(longTask)
		if err == nil {
			t.Fatal("expected error on task > 100 chars")
		}
		if !errors.Is(err, ErrTaskTooLong) {
			t.Errorf("error = %q, want %q", err.Error(), ErrTaskTooLong)
		}
	})
}

func TestTakeBreak(t *testing.T) {
	t.Run("from flow succeeds", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		_ = state.StartSession("write code")
		clock.Advance(30 * time.Minute)

		err := state.TakeBreak()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.CurrentBreak == nil {
			t.Fatal("expected current break to be set")
		}
		if !state.CurrentBreak.StartTime.Equal(clock.Now()) {
			t.Errorf("break start = %v, want %v", state.CurrentBreak.StartTime, clock.Now())
		}
		// 30 minutes of work -> 8 minute suggested break
		if state.CurrentBreak.SuggestedDuration != 8*time.Minute {
			t.Errorf("suggested duration = %v, want %v", state.CurrentBreak.SuggestedDuration, 8*time.Minute)
		}
	})

	t.Run("errors when no session", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		err := state.TakeBreak()
		if err == nil {
			t.Fatal("expected error when no session active")
		}
		if !errors.Is(err, ErrNoActiveSession) {
			t.Errorf("error = %q, want %q", err.Error(), ErrNoActiveSession)
		}
	})

	t.Run("errors when already on break", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		_ = state.StartSession("write code")
		clock.Advance(10 * time.Minute)
		_ = state.TakeBreak()

		err := state.TakeBreak()
		if err == nil {
			t.Fatal("expected error when already on break")
		}
		if !errors.Is(err, ErrAlreadyOnBreak) {
			t.Errorf("error = %q, want %q", err.Error(), ErrAlreadyOnBreak)
		}
	})
}

func TestResume(t *testing.T) {
	t.Run("from break returns true with correct completed session", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		_ = state.StartSession("write code")
		clock.Advance(25 * time.Minute) // 25 min flow
		_ = state.TakeBreak()
		clock.Advance(5 * time.Minute) // 5 min break

		resumed, err := state.Resume()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resumed {
			t.Error("expected resumedCurrent = true")
		}

		// Should have one completed session
		if len(state.CompletedSessions) != 1 {
			t.Fatalf("completed sessions = %d, want 1", len(state.CompletedSessions))
		}
		cs := state.CompletedSessions[0]
		if cs.Task != "write code" {
			t.Errorf("completed task = %q, want %q", cs.Task, "write code")
		}
		if cs.FlowDuration != 25*time.Minute {
			t.Errorf("flow duration = %v, want %v", cs.FlowDuration, 25*time.Minute)
		}
		if cs.BreakDuration == nil || *cs.BreakDuration != 5*time.Minute {
			t.Errorf("break duration = %v, want %v", cs.BreakDuration, 5*time.Minute)
		}

		// New session should be started with same task
		if state.CurrentSession == nil {
			t.Fatal("expected new current session")
		}
		if state.CurrentSession.Task != "write code" {
			t.Errorf("new session task = %q, want %q", state.CurrentSession.Task, "write code")
		}
		if state.CurrentBreak != nil {
			t.Error("expected current break to be cleared")
		}
	})

	t.Run("from idle with history returns false", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		// Set up: completed session exists, no active session
		state.CompletedSessions = []CompletedSession{
			{Task: "old task", FlowDuration: 20 * time.Minute, CompletedAt: clock.Now()},
		}

		clock.Advance(10 * time.Minute)

		resumed, err := state.Resume()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resumed {
			t.Error("expected resumedCurrent = false")
		}
		if state.CurrentSession == nil {
			t.Fatal("expected new current session")
		}
		if state.CurrentSession.Task != "old task" {
			t.Errorf("new session task = %q, want %q", state.CurrentSession.Task, "old task")
		}
	})

	t.Run("errors when already flowing", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		_ = state.StartSession("write code")

		_, err := state.Resume()
		if err == nil {
			t.Fatal("expected error when already flowing")
		}
		if !errors.Is(err, ErrAlreadyFlowing) {
			t.Errorf("error = %q, want %q", err.Error(), ErrAlreadyFlowing)
		}
	})

	t.Run("errors with no session and no history", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		_, err := state.Resume()
		if err == nil {
			t.Fatal("expected error with no session and no history")
		}
		if !errors.Is(err, ErrNoSessionToResume) {
			t.Errorf("error = %q, want %q", err.Error(), ErrNoSessionToResume)
		}
	})
}

func TestStop(t *testing.T) {
	t.Run("from flow returns completed session with nil break", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		_ = state.StartSession("write code")
		clock.Advance(45 * time.Minute)

		completed, err := state.Stop()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if completed.Task != "write code" {
			t.Errorf("task = %q, want %q", completed.Task, "write code")
		}
		if completed.FlowDuration != 45*time.Minute {
			t.Errorf("flow duration = %v, want %v", completed.FlowDuration, 45*time.Minute)
		}
		if completed.BreakDuration != nil {
			t.Errorf("break duration = %v, want nil", completed.BreakDuration)
		}
		if state.CurrentSession != nil {
			t.Error("expected current session to be cleared")
		}
		if len(state.CompletedSessions) != 1 {
			t.Fatalf("completed sessions = %d, want 1", len(state.CompletedSessions))
		}
	})

	t.Run("from break returns completed session with break duration", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		_ = state.StartSession("write code")
		clock.Advance(30 * time.Minute) // flow
		_ = state.TakeBreak()
		clock.Advance(7 * time.Minute) // break

		completed, err := state.Stop()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if completed.FlowDuration != 30*time.Minute {
			t.Errorf("flow duration = %v, want %v", completed.FlowDuration, 30*time.Minute)
		}
		if completed.BreakDuration == nil {
			t.Fatal("expected break duration to be set")
		}
		if *completed.BreakDuration != 7*time.Minute {
			t.Errorf("break duration = %v, want %v", *completed.BreakDuration, 7*time.Minute)
		}
		if state.CurrentSession != nil {
			t.Error("expected current session to be cleared")
		}
		if state.CurrentBreak != nil {
			t.Error("expected current break to be cleared")
		}
	})

	t.Run("errors when no session", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		_, err := state.Stop()
		if err == nil {
			t.Fatal("expected error when no session active")
		}
		if !errors.Is(err, ErrNoActiveSession) {
			t.Errorf("error = %q, want %q", err.Error(), ErrNoActiveSession)
		}
	})
}

func TestFullCycle(t *testing.T) {
	clock := newTestClock()
	state := NewFlowState(clock)

	// Session 1: start -> break -> resume
	err := state.StartSession("implement feature")
	if err != nil {
		t.Fatalf("start session 1: %v", err)
	}

	clock.Advance(40 * time.Minute) // 40 min flow

	err = state.TakeBreak()
	if err != nil {
		t.Fatalf("take break 1: %v", err)
	}

	clock.Advance(8 * time.Minute) // 8 min break

	resumed, err := state.Resume()
	if err != nil {
		t.Fatalf("resume 1: %v", err)
	}
	if !resumed {
		t.Error("expected resumedCurrent = true for first resume")
	}

	// Verify first completed session
	if len(state.CompletedSessions) != 1 {
		t.Fatalf("after resume 1: completed sessions = %d, want 1", len(state.CompletedSessions))
	}
	cs1 := state.CompletedSessions[0]
	if cs1.FlowDuration != 40*time.Minute {
		t.Errorf("session 1 flow = %v, want %v", cs1.FlowDuration, 40*time.Minute)
	}
	if cs1.BreakDuration == nil || *cs1.BreakDuration != 8*time.Minute {
		t.Errorf("session 1 break = %v, want %v", cs1.BreakDuration, 8*time.Minute)
	}

	// Session 2: (already resumed) -> break -> stop
	clock.Advance(20 * time.Minute) // 20 min flow

	err = state.TakeBreak()
	if err != nil {
		t.Fatalf("take break 2: %v", err)
	}

	clock.Advance(5 * time.Minute) // 5 min break

	completed, err := state.Stop()
	if err != nil {
		t.Fatalf("stop: %v", err)
	}

	// Verify second completed session
	if len(state.CompletedSessions) != 2 {
		t.Fatalf("after stop: completed sessions = %d, want 2", len(state.CompletedSessions))
	}
	if completed.FlowDuration != 20*time.Minute {
		t.Errorf("session 2 flow = %v, want %v", completed.FlowDuration, 20*time.Minute)
	}
	if completed.BreakDuration == nil || *completed.BreakDuration != 5*time.Minute {
		t.Errorf("session 2 break = %v, want %v", completed.BreakDuration, 5*time.Minute)
	}
	if completed.Task != "implement feature" {
		t.Errorf("session 2 task = %q, want %q", completed.Task, "implement feature")
	}

	// State should be idle
	if state.CurrentSession != nil {
		t.Error("expected no current session after stop")
	}
	if state.CurrentBreak != nil {
		t.Error("expected no current break after stop")
	}
}

func TestCancelSession(t *testing.T) {
	t.Run("from flow clears session", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		_ = state.StartSession("write code")
		clock.Advance(10 * time.Minute)

		err := state.CancelSession()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.CurrentSession != nil {
			t.Error("expected current session to be cleared")
		}
		if len(state.CompletedSessions) != 0 {
			t.Errorf("completed sessions = %d, want 0", len(state.CompletedSessions))
		}
	})

	t.Run("from break clears session and break", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		_ = state.StartSession("write code")
		clock.Advance(20 * time.Minute)
		_ = state.TakeBreak()

		err := state.CancelSession()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.CurrentSession != nil {
			t.Error("expected current session to be cleared")
		}
		if state.CurrentBreak != nil {
			t.Error("expected current break to be cleared")
		}
	})

	t.Run("errors when no session", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		err := state.CancelSession()
		if err == nil {
			t.Fatal("expected error when no session active")
		}
		if !errors.Is(err, ErrNoActiveSession) {
			t.Errorf("error = %q, want %q", err.Error(), ErrNoActiveSession)
		}
	})
}

func TestActiveSessions(t *testing.T) {
	t.Run("returns only non-deleted sessions", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		now := clock.Now()
		state.CompletedSessions = []CompletedSession{
			{Task: "task 1", FlowDuration: 10 * time.Minute, CompletedAt: now},
			{Task: "task 2", FlowDuration: 20 * time.Minute, CompletedAt: now, DeletedAt: &now},
			{Task: "task 3", FlowDuration: 30 * time.Minute, CompletedAt: now},
		}

		active := state.ActiveSessions()
		if len(active) != 2 {
			t.Fatalf("active sessions = %d, want 2", len(active))
		}
		if active[0].Task != "task 1" {
			t.Errorf("active[0].Task = %q, want %q", active[0].Task, "task 1")
		}
		if active[1].Task != "task 3" {
			t.Errorf("active[1].Task = %q, want %q", active[1].Task, "task 3")
		}
	})

	t.Run("returns empty slice when all deleted", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		now := clock.Now()
		state.CompletedSessions = []CompletedSession{
			{Task: "task 1", FlowDuration: 10 * time.Minute, CompletedAt: now, DeletedAt: &now},
		}

		active := state.ActiveSessions()
		if len(active) != 0 {
			t.Fatalf("active sessions = %d, want 0", len(active))
		}
	})

	t.Run("returns empty slice when no sessions", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		active := state.ActiveSessions()
		if len(active) != 0 {
			t.Fatalf("active sessions = %d, want 0", len(active))
		}
	})
}

func TestDeleteSession(t *testing.T) {
	t.Run("soft-deletes session at index", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		state.CompletedSessions = []CompletedSession{
			{Task: "task 1", FlowDuration: 10 * time.Minute, CompletedAt: clock.Now()},
			{Task: "task 2", FlowDuration: 20 * time.Minute, CompletedAt: clock.Now()},
		}

		clock.Advance(time.Hour)

		err := state.DeleteSession(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.CompletedSessions[0].DeletedAt == nil {
			t.Fatal("expected DeletedAt to be set")
		}
		if !state.CompletedSessions[0].DeletedAt.Equal(clock.Now()) {
			t.Errorf("DeletedAt = %v, want %v", state.CompletedSessions[0].DeletedAt, clock.Now())
		}
		// Second session should be untouched
		if state.CompletedSessions[1].DeletedAt != nil {
			t.Error("expected second session DeletedAt to be nil")
		}
	})

	t.Run("errors on out-of-range index", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		err := state.DeleteSession(0)
		if !errors.Is(err, ErrSessionNotFound) {
			t.Errorf("error = %v, want %v", err, ErrSessionNotFound)
		}

		err = state.DeleteSession(-1)
		if !errors.Is(err, ErrSessionNotFound) {
			t.Errorf("error = %v, want %v", err, ErrSessionNotFound)
		}
	})

	t.Run("errors on already-deleted session", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		now := clock.Now()
		state.CompletedSessions = []CompletedSession{
			{Task: "task 1", FlowDuration: 10 * time.Minute, CompletedAt: now, DeletedAt: &now},
		}

		err := state.DeleteSession(0)
		if !errors.Is(err, ErrSessionDeleted) {
			t.Errorf("error = %v, want %v", err, ErrSessionDeleted)
		}
	})
}

func TestDeleteAllSessions(t *testing.T) {
	t.Run("soft-deletes all active sessions", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		now := clock.Now()
		state.CompletedSessions = []CompletedSession{
			{Task: "task 1", FlowDuration: 10 * time.Minute, CompletedAt: now},
			{Task: "task 2", FlowDuration: 20 * time.Minute, CompletedAt: now, DeletedAt: &now},
			{Task: "task 3", FlowDuration: 30 * time.Minute, CompletedAt: now},
		}

		clock.Advance(time.Hour)

		err := state.DeleteAllSessions()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i, cs := range state.CompletedSessions {
			if cs.DeletedAt == nil {
				t.Errorf("session %d: expected DeletedAt to be set", i)
			}
		}

		// Already-deleted session should keep its original DeletedAt
		if !state.CompletedSessions[1].DeletedAt.Equal(now) {
			t.Errorf("previously deleted session: DeletedAt = %v, want %v", state.CompletedSessions[1].DeletedAt, now)
		}

		// Newly deleted sessions should use current clock time
		if !state.CompletedSessions[0].DeletedAt.Equal(clock.Now()) {
			t.Errorf("session 0: DeletedAt = %v, want %v", state.CompletedSessions[0].DeletedAt, clock.Now())
		}
	})

	t.Run("errors when no active sessions", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		err := state.DeleteAllSessions()
		if !errors.Is(err, ErrNoSessionsToDelete) {
			t.Errorf("error = %v, want %v", err, ErrNoSessionsToDelete)
		}
	})

	t.Run("errors when all already deleted", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		now := clock.Now()
		state.CompletedSessions = []CompletedSession{
			{Task: "task 1", FlowDuration: 10 * time.Minute, CompletedAt: now, DeletedAt: &now},
		}

		err := state.DeleteAllSessions()
		if !errors.Is(err, ErrNoSessionsToDelete) {
			t.Errorf("error = %v, want %v", err, ErrNoSessionsToDelete)
		}
	})
}

func TestResumeSkipsDeletedSessions(t *testing.T) {
	clock := newTestClock()
	state := NewFlowState(clock)

	now := clock.Now()
	deletedAt := now
	state.CompletedSessions = []CompletedSession{
		{Task: "old task", FlowDuration: 10 * time.Minute, CompletedAt: now},
		{Task: "deleted task", FlowDuration: 20 * time.Minute, CompletedAt: now, DeletedAt: &deletedAt},
	}

	clock.Advance(time.Hour)

	resumed, err := state.Resume()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resumed {
		t.Error("expected resumedCurrent = false")
	}
	if state.CurrentSession.Task != "old task" {
		t.Errorf("task = %q, want %q", state.CurrentSession.Task, "old task")
	}
}
