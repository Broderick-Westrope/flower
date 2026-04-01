package flowtime

import (
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
		if !strings.Contains(err.Error(), "session already active") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "session already active")
		}
	})

	t.Run("errors on empty task", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		err := state.StartSession("")
		if err == nil {
			t.Fatal("expected error on empty task")
		}
		if !strings.Contains(err.Error(), "must not be empty") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "must not be empty")
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
		if !strings.Contains(err.Error(), "100 characters or less") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "100 characters or less")
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
		if !strings.Contains(err.Error(), "no active session") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "no active session")
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
		if !strings.Contains(err.Error(), "already on break") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "already on break")
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
		if !strings.Contains(err.Error(), "already in flow state") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "already in flow state")
		}
	})

	t.Run("errors with no session and no history", func(t *testing.T) {
		clock := newTestClock()
		state := NewFlowState(clock)

		_, err := state.Resume()
		if err == nil {
			t.Fatal("expected error with no session and no history")
		}
		if !strings.Contains(err.Error(), "no session to resume") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "no session to resume")
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
		if !strings.Contains(err.Error(), "no active session") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "no active session")
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
