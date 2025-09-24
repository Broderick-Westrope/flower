package main

import (
	"fmt"
	"os"

	"github.com/Broderick-Westrope/flower/pkg/core"
	"github.com/Broderick-Westrope/flower/pkg/tui"
	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
)

type CLI struct {
	Start  StartCmd  `cmd:"" help:"Start flow, creating a new session if needed."`
	Break  BreakCmd  `cmd:"" help:"End flow, start break."`
	Resume ResumeCmd `cmd:"" help:"End break, resume the current or previous session in a flow state."`
	Stop   StopCmd   `cmd:"" help:"End current session."`
	Status StatusCmd `cmd:"" help:"Show current state."`
	Log    LogCmd    `cmd:"" help:"Show recent sessions."`
	Locate LocateCmd `cmd:"" help:"Show the state file path."`
}

func main() {
	// If no arguments provided, launch TUI directly
	if len(os.Args) == 1 {
		err := startTUI(nil)
		if err != nil {
			panic(err)
		}
		return
	}

	var cli CLI

	ctx := kong.Parse(&cli,
		kong.Name("flower"),
		kong.Description("A minimal Flowtime Technique CLI tool"),
	)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

func startTUI(state *core.AppState) error {
	m, err := tui.New(state)
	if err != nil {
		return fmt.Errorf("creating new model: %w", err)
	}

	_, err = tea.NewProgram(m).Run()
	if err != nil {
		return fmt.Errorf("running program: %w", err)
	}
	return nil
}
