package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/Broderick-Westrope/flower/internal/flowtime"
	"github.com/Broderick-Westrope/flower/internal/storage"
)

// Context holds shared dependencies for CLI commands.
type Context struct {
	Store       storage.Store
	RunTUI      func(store storage.Store) error // injected by main.go to avoid circular import
	LocateStore func() (string, error)          // returns state file path
}

// CLI is the top-level Kong command structure.
type CLI struct {
	Start  StartCmd  `cmd:"" help:"Start flow, creating a new session if needed."`
	Break  BreakCmd  `cmd:"" help:"End flow, start break."`
	Resume ResumeCmd `cmd:"" help:"End break, resume the current or previous session."`
	Stop   StopCmd   `cmd:"" help:"End current session."`
	Cancel CancelCmd `cmd:"" help:"Cancel the current session without recording it."`
	Status StatusCmd `cmd:"" help:"Show current state."`
	Log    LogCmd    `cmd:"" help:"Show recent sessions."`
	Delete DeleteCmd `cmd:"" help:"Delete a completed session by index."`
	Clear  ClearCmd  `cmd:"" help:"Delete all completed sessions."`
	Locate LocateCmd `cmd:"" help:"Show the state file path."`
}

// StartCmd begins a new flow session.
type StartCmd struct {
	Task string `arg:"" help:"Task description"`

	Detach bool `short:"d"`
}

func (cmd *StartCmd) Run(ctx *Context) error {
	state, err := ctx.Store.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	if err := state.StartSession(cmd.Task); err != nil {
		return fmt.Errorf("starting session: %w", err)
	}

	if err := ctx.Store.Save(state); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	if cmd.Detach {
		fmt.Printf("Started: %s at %s\n", cmd.Task, state.CurrentSession.StartTime.Format("15:04"))
		return nil
	}

	return ctx.RunTUI(ctx.Store)
}

// BreakCmd ends the current flow and starts a break.
type BreakCmd struct {
	Detach bool `short:"d"`
}

func (cmd *BreakCmd) Run(ctx *Context) error {
	state, err := ctx.Store.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	if err := state.TakeBreak(); err != nil {
		return fmt.Errorf("taking break: %w", err)
	}

	if err := ctx.Store.Save(state); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	if cmd.Detach {
		fmt.Printf("Flow ended. Starting %s break.\n",
			flowtime.FormatDuration(state.CurrentBreak.SuggestedDuration))
		return nil
	}

	return ctx.RunTUI(ctx.Store)
}

// ResumeCmd ends a break and resumes the current or previous session.
type ResumeCmd struct {
	Detach bool `short:"d"`
}

func (cmd *ResumeCmd) Run(ctx *Context) error {
	state, err := ctx.Store.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	// Already flowing — no state change needed
	if state.CurrentSession != nil && state.CurrentBreak == nil {
		if cmd.Detach {
			return errors.New("already flowing")
		}
		return ctx.RunTUI(ctx.Store)
	}

	resumedCurrent, err := state.Resume()
	if err != nil {
		return fmt.Errorf("resuming session: %w", err)
	}

	if err := ctx.Store.Save(state); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	if cmd.Detach {
		kind := "Previous"
		if resumedCurrent {
			kind = "Current"
		}
		fmt.Printf("Break ended. %s session resumed.\n", kind)
		return nil
	}

	return ctx.RunTUI(ctx.Store)
}

// StopCmd ends the current session.
type StopCmd struct{}

func (cmd *StopCmd) Run(ctx *Context) error {
	state, err := ctx.Store.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	if _, err := state.Stop(); err != nil {
		return fmt.Errorf("stopping session: %w", err)
	}

	if err := ctx.Store.Save(state); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	fmt.Println("Session ended.")
	return nil
}

// StatusCmd shows the current flow state.
type StatusCmd struct{}

func (cmd *StatusCmd) Run(ctx *Context) error {
	state, err := ctx.Store.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	PrintStatus(state, time.Now())
	return nil
}

// LogCmd shows recent completed sessions.
type LogCmd struct {
	Count int `default:"10" help:"Entries per page"`
	Page  int `default:"1" help:"Page to display"`
}

func (cmd *LogCmd) Run(ctx *Context) error {
	if cmd.Count <= 0 {
		return errors.New("count must be greater than zero")
	}
	if cmd.Page <= 0 {
		return errors.New("page must be greater than zero")
	}

	state, err := ctx.Store.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	PrintLog(state.ActiveSessions(), cmd.Page, cmd.Count, time.Now())
	return nil
}

// CancelCmd discards the current session without recording it.
type CancelCmd struct {
	Yes bool `short:"y" help:"Skip confirmation prompt."`
}

func (cmd *CancelCmd) Run(ctx *Context) error {
	state, err := ctx.Store.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	if state.CurrentSession == nil {
		return errors.New("no active session to cancel")
	}

	if !cmd.Yes {
		ok, err := confirm(fmt.Sprintf("Cancel session %q?", state.CurrentSession.Task))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := state.CancelSession(); err != nil {
		return fmt.Errorf("cancelling session: %w", err)
	}

	if err := ctx.Store.Save(state); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	fmt.Println("Session cancelled.")
	return nil
}

// DeleteCmd soft-deletes a completed session by its display index (1-based, newest first).
type DeleteCmd struct {
	Index int  `arg:"" help:"Session number (1 = most recent)."`
	Yes   bool `short:"y" help:"Skip confirmation prompt."`
}

func (cmd *DeleteCmd) Run(ctx *Context) error {
	if cmd.Index <= 0 {
		return errors.New("index must be a positive number")
	}

	state, err := ctx.Store.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	active := state.ActiveSessions()
	if len(active) == 0 {
		return errors.New("no sessions to delete")
	}

	// Convert 1-based newest-first display index to the underlying slice index.
	displayIndex := cmd.Index - 1
	if displayIndex >= len(active) {
		return fmt.Errorf("index %d out of range (have %d sessions)", cmd.Index, len(active))
	}

	// Map from active-list index (newest-first) to full CompletedSessions index.
	// Active list preserves chronological order; display index 0 = last active session.
	activeReversed := len(active) - 1 - displayIndex
	target := active[activeReversed]

	// Find the matching entry in the full slice.
	fullIndex := -1
	for i, cs := range state.CompletedSessions {
		if cs.CompletedAt.Equal(target.CompletedAt) && cs.Task == target.Task && cs.DeletedAt == nil {
			fullIndex = i
			break
		}
	}
	if fullIndex == -1 {
		return errors.New("session not found")
	}

	if !cmd.Yes {
		ok, err := confirm(fmt.Sprintf("Delete session %q (%s)?",
			target.Task, flowtime.FormatDuration(target.FlowDuration)))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := state.DeleteSession(fullIndex); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	if err := ctx.Store.Save(state); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	fmt.Println("Session deleted.")
	return nil
}

// ClearCmd soft-deletes all completed sessions.
type ClearCmd struct {
	Yes bool `short:"y" help:"Skip confirmation prompt."`
}

func (cmd *ClearCmd) Run(ctx *Context) error {
	state, err := ctx.Store.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	active := state.ActiveSessions()
	if len(active) == 0 {
		return errors.New("no sessions to delete")
	}

	if !cmd.Yes {
		ok, err := confirm(fmt.Sprintf("Delete all %d sessions?", len(active)))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := state.DeleteAllSessions(); err != nil {
		return fmt.Errorf("deleting sessions: %w", err)
	}

	if err := ctx.Store.Save(state); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	fmt.Printf("Deleted %d sessions.\n", len(active))
	return nil
}

// confirm prompts the user with the given message and reads y/n from stdin.
func confirm(prompt string) (bool, error) {
	fmt.Printf("%s [y/N] ", prompt)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// EOF or empty input = no
		return false, nil
	}
	return response == "y" || response == "Y", nil
}

// LocateCmd shows the state file path.
type LocateCmd struct{}

func (cmd *LocateCmd) Run(ctx *Context) error {
	fp, err := ctx.LocateStore()
	if err != nil {
		return fmt.Errorf("getting state file path: %w", err)
	}

	fmt.Println(fp)
	return nil
}
