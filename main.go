package main

import (
	"errors"

	"github.com/alecthomas/kong"
)

type CLI struct {
	GlobalVars

	Root  RootCmd       `cmd:"" default:"1" hidden:""`
	Start StartTimerCmd `cmd:""`
	Stop  StopTimerCmd  `cmd:""`

	Task struct {
		Add    AddTaskCmd `cmd:"" default:"1"`
		Remove AddTaskCmd `cmd:"" aliases:"rm"`
		// TODO(feat): archive/activate, list, info/stats (for specific task)
	} `cmd:"" help:"Start in the menu"`
}

type GlobalVars struct{}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("flower"),
		kong.Description("A tool for using the flowtime time-management method."),
		kong.UsageOnError(),
	)

	if err := handleDefaultGlobals(&cli.GlobalVars); err != nil {
		ctx.FatalIfErrorf(err)
	}

	err := ctx.Run(&cli.GlobalVars)
	ctx.FatalIfErrorf(err)
}

func handleDefaultGlobals(_ *GlobalVars) error {
	return nil
}

// Subcommands ---------------------------------------------------

type RootCmd struct{}

func (c *RootCmd) Run(globals *GlobalVars) error {
	panic(errors.New("unimplemented: TUI"))
}

type StartTimerCmd struct{}

func (c *StartTimerCmd) Run(globals *GlobalVars) error {
	panic(errors.New("unimplemented: start timer for task"))
}

type StopTimerCmd struct{}

func (c *StopTimerCmd) Run(globals *GlobalVars) error {
	panic(errors.New("unimplemented: stop timer for task"))
}

type AddTaskCmd struct{}

func (c *AddTaskCmd) Run(globals *GlobalVars) error {
	panic(errors.New("unimplemented: add new task"))
}

type RemoveTaskCmd struct{}

func (c *RemoveTaskCmd) Run(globals *GlobalVars) error {
	panic(errors.New("unimplemented: remove existing task"))
}
