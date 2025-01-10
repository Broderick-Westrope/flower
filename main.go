package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Broderick-Westrope/flower/internal/data"
	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"
	"github.com/charmbracelet/huh"
	_ "github.com/mattn/go-sqlite3"
)

type CLI struct {
	Root  RootCmd         `cmd:"" name:"tui" default:"1" help:"Interactive Terminal-UI."`
	Start StartSessionCmd `cmd:"" help:"Start a new session."`
	Stop  StopSessionCmd  `cmd:"" help:"Stop the open sessions."`
	List  ListSessionsCmd `cmd:"" help:"List information about all sessions."`

	Task struct {
		Add    AddTaskCmd    `cmd:"" help:"Add a new task."`
		Get    GetTaskCmd    `cmd:"" help:"Get information about a task."`
		List   ListTasksCmd  `cmd:"" help:"List information about all tasks."`
		Remove RemoveTaskCmd `cmd:"" aliases:"rm" help:"Remove a task and associated sessions. (destructive)"`
		Clear  ClearTasksCmd `cmd:"" help:"Remove all tasks and associated sessions. (destructive)"`
		// TODO(feat): archive/activate, info/stats (for specific task)
	} `cmd:""`
}

type GlobalDependencies struct {
	Repo *data.Respository
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

func setupGlobalDependencies() (*GlobalDependencies, func(), error) {
	dataDir := filepath.Join(xdg.DataHome, "flower")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("creating data directory: %w", err)
	}

	dsn := "file://" + filepath.Join(dataDir, "sqlite.db?_fk=1")
	db, err := setupDatabase(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("setting up database: %w", err)
	}

	deps := &GlobalDependencies{
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

type RootCmd struct{}

func (c *RootCmd) Run(deps *GlobalDependencies) error {
	panic(errors.New("unimplemented: TUI"))
}
