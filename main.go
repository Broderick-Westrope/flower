package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Broderick-Westrope/flower/internal/cli"
	"github.com/Broderick-Westrope/flower/internal/data"
	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"
	"github.com/charmbracelet/huh"
	_ "github.com/mattn/go-sqlite3"
)

type CLI struct {
	Root  cli.RootCmd         `cmd:"" name:"tui" default:"1" help:"Interactive Terminal-UI."`
	Start cli.StartSessionCmd `cmd:"" help:"Start a new session."`
	Stop  cli.StopSessionCmd  `cmd:"" help:"Stop the open sessions."`
	List  cli.ListSessionsCmd `cmd:"" help:"List information about all sessions."`

	Task struct {
		Add    cli.AddTaskCmd    `cmd:"" help:"Add a new task."`
		Get    cli.GetTaskCmd    `cmd:"" help:"Get information about a task."`
		List   cli.ListTasksCmd  `cmd:"" help:"List information about all tasks."`
		Remove cli.RemoveTaskCmd `cmd:"" aliases:"rm" help:"Remove a task and associated sessions. (destructive)"`
		Clear  cli.ClearTasksCmd `cmd:"" help:"Remove all tasks and associated sessions. (destructive)"`
		// TODO(feat): archive/activate
	} `cmd:""`
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("flower"),
		kong.Description("A tool for using the flowtime time-management method."),
		kong.UsageOnError(),
	)

	deps, deferFunc, err := setupGlobalDependencies()
	if err != nil {
		log.Fatal(err)
	}
	defer deferFunc()

	err = ctx.Run(deps)
	if errors.Is(err, huh.ErrUserAborted) {
		fmt.Println("User canceled.")
		return
	}
	ctx.FatalIfErrorf(err)
}

func setupGlobalDependencies() (*cli.GlobalDependencies, func(), error) {
	dataDir := filepath.Join(xdg.DataHome, "flower")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("creating data directory: %w", err)
	}

	dsn := "file://" + filepath.Join(dataDir, "sqlite.db?_fk=1")
	db, err := setupDatabase(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("setting up database: %w", err)
	}

	deps := &cli.GlobalDependencies{
		Repo: data.NewRepository(db),
	}

	deferFunc := func() {
		db.Close()
	}
	return deps, deferFunc, nil
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
