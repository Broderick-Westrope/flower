package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Broderick-Westrope/flower/internal/data"
	"github.com/alecthomas/kong"
	_ "github.com/mattn/go-sqlite3"
)

type CLI struct {
	Root  RootCmd       `cmd:"" default:"1" hidden:""`
	Start StartTimerCmd `cmd:""`
	Stop  StopTimerCmd  `cmd:""`

	Task struct {
		Add    AddTaskCmd `cmd:"" default:"1"`
		Remove AddTaskCmd `cmd:"" aliases:"rm"`
		// TODO(feat): archive/activate, list, info/stats (for specific task)
	} `cmd:"" help:"Start in the menu"`
}

type GlobalDependencies struct {
	Repo data.Respository
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("flower"),
		kong.Description("A tool for using the flowtime time-management method."),
		kong.UsageOnError(),
	)

	deps := setupGlobalDependencies()
	ctx.FatalIfErrorf(ctx.Run(deps))
}

func setupGlobalDependencies() *GlobalDependencies {
	dsn := os.Getenv("DSN")
	db, err := setupDatabase(dsn)
	if err != nil {
		log.Fatalf("failed to setup database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	return &GlobalDependencies{
		Repo: data.NewRepository(db),
	}
}

func setupDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("pinging: %w", err)
	}
	return db, nil
}

// Subcommands ---------------------------------------------------

type RootCmd struct{}

func (c *RootCmd) Run(deps *GlobalDependencies) error {
	panic(errors.New("unimplemented: TUI"))
}

type StartTimerCmd struct{}

func (c *StartTimerCmd) Run(deps *GlobalDependencies) error {
	deps.Repo.CreateTask(context.Background(), "SOME", "ANOTHER")

	return nil
}

type StopTimerCmd struct{}

func (c *StopTimerCmd) Run(deps *GlobalDependencies) error {
	panic(errors.New("unimplemented: stop timer for task"))
}

type AddTaskCmd struct{}

func (c *AddTaskCmd) Run(deps *GlobalDependencies) error {
	panic(errors.New("unimplemented: add new task"))
}

type RemoveTaskCmd struct{}

func (c *RemoveTaskCmd) Run(deps *GlobalDependencies) error {
	panic(errors.New("unimplemented: remove existing task"))
}
