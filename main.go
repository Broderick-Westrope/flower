package main

import (
	"github.com/alecthomas/kong"
)

type CLI struct {
	Work   WorkCmd   `cmd:"" help:"Start work session"`
	Break  BreakCmd  `cmd:"" help:"End work, start break"`
	Resume ResumeCmd `cmd:"" help:"End break, return to work"`
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
