package main

import (
	"fmt"
	"os"

	"github.com/Broderick-Westrope/flower/internal/cli"
	"github.com/Broderick-Westrope/flower/internal/flowtime"
	"github.com/Broderick-Westrope/flower/internal/storage"
	"github.com/Broderick-Westrope/flower/internal/tui"
	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	clock := flowtime.RealClock{}
	jsonStore := storage.NewJSONStore(clock)

	if len(os.Args) == 1 {
		if err := runTUI(jsonStore); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	ctx := &cli.Context{
		Store:       jsonStore,
		RunTUI:      runTUI,
		LocateStore: jsonStore.GetFilePath,
	}
	var c cli.CLI
	kongCtx := kong.Parse(&c,
		kong.Name("flower"),
		kong.Description("A minimal Flowtime Technique CLI tool"),
	)
	err := kongCtx.Run(ctx)
	kongCtx.FatalIfErrorf(err)
}

func runTUI(store storage.Store) error {
	m, err := tui.New(store)
	if err != nil {
		return fmt.Errorf("creating TUI model: %w", err)
	}
	_, err = tea.NewProgram(m).Run()
	if err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}
	return nil
}
