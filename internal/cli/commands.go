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
	Status StatusCmd `cmd:"" help:"Show current state."`
	Log    LogCmd    `cmd:"" help:"Show recent sessions."`
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

	PrintLog(state.CompletedSessions, cmd.Page, cmd.Count, time.Now())
	return nil
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
