package main

import (
	"github.com/alecthomas/kong"
)

type CLI struct {
	Start  StartCmd  `cmd:"" help:"Start flow, creating a new session if needed."`
	Break  BreakCmd  `cmd:"" help:"End flow, start break."`
	Resume ResumeCmd `cmd:"" help:"End break, resume the previous flow."`
	Stop   StopCmd   `cmd:"" help:"End current session"`
	Status StatusCmd `cmd:"" help:"Show current state"`
	Log    LogCmd    `cmd:"" help:"Show recent sessions"`
}

func main() {
	var cli CLI

	ctx := kong.Parse(&cli,
		kong.Name("flower"),
		kong.Description("A minimal Flowtime Technique CLI tool"),
	)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
