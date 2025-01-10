package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/Broderick-Westrope/flower/gen/model"
	"github.com/Broderick-Westrope/flower/internal/data"
	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3"
)

type CLI struct {
	Root  RootCmd       `cmd:"" default:"1" hidden:""`
	Start StartTimerCmd `cmd:""`
	Stop  StopTimerCmd  `cmd:""`

	Task struct {
		Add    AddTaskCmd    `cmd:"" default:"1"`
		Remove RemoveTaskCmd `cmd:"" aliases:"rm"`
		List   ListTasksCmd  `cmd:""`
		Clear  ClearTasksCmd `cmd:""`
		// TODO(feat): archive/activate, info/stats (for specific task)
	} `cmd:""`
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

	deps, deferFunc, err := setupGlobalDependencies()
	if err != nil {
		log.Fatal(err)
	}
	defer deferFunc()

	ctx.FatalIfErrorf(ctx.Run(deps))
}

func setupGlobalDependencies() (*GlobalDependencies, func(), error) {
	dataDir := filepath.Join(xdg.DataHome, "flower")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("creating data directory: %w", err)
	}

	dsn := "file://" + filepath.Join(dataDir, "sqlite.db")
	db, err := setupDatabase(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup database: %w", err)
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

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
  			id integer NOT NULL,
  			name text NOT NULL,
  			description text NOT NULL,
  			PRIMARY KEY (id)
		);
    `)
	if err != nil {
		return nil, fmt.Errorf("creating tasks table: %w", err)
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
	panic(errors.New("unimplemented: start timer for task"))
}

type StopTimerCmd struct{}

func (c *StopTimerCmd) Run(deps *GlobalDependencies) error {
	panic(errors.New("unimplemented: stop timer for task"))
}

type AddTaskCmd struct {
	Name        string `help:"task name"`
	Description string `aliases:"desc" help:"task description"`
}

func (c *AddTaskCmd) Run(deps *GlobalDependencies) error {
	var task *model.Tasks
	var err error
	if len(c.Name) == 0 {
		fmt.Println("Creating a new task...")
		task, err = promptForNewTask()
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				fmt.Println("Task creation canceled.")
				return nil
			}
			return err
		}
	} else {
		task = &model.Tasks{
			Name:        c.Name,
			Description: c.Description,
		}
	}

	if task == nil {
		return errors.New("task object was nil")
	}

	fmt.Println("Adding task...")
	task, err = deps.Repo.CreateTask(context.Background(), task.Name, task.Description)
	if err != nil {
		return err
	}

	fmt.Printf("New task added! 🎉\n\n%s\n",
		lipgloss.JoinHorizontal(lipgloss.Top,
			"  ",
			stringifyTask(*task),
		),
	)
	return nil
}

type RemoveTaskCmd struct {
	TaskID int `arg:"" name:"task-id" help:"ID of the task to remove"`
}

func (c *RemoveTaskCmd) Run(deps *GlobalDependencies) error {
	err := deps.Repo.DeleteTask(context.Background(), c.TaskID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			fmt.Printf("Task with ID %d not found.\n", c.TaskID)
			return nil
		}
		return err
	}
	return nil
}

type ListTasksCmd struct {
	AsJSON bool `name:"json" help:"marshal the list as JSON"`
}

func (c *ListTasksCmd) Run(deps *GlobalDependencies) error {
	tasks, err := deps.Repo.ListTasks(context.Background())
	if err != nil {
		return err
	}

	if c.AsJSON {
		jsonBytes, err := json.MarshalIndent(tasks, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", jsonBytes)
		return nil
	}

	for _, task := range tasks {
		fmt.Printf("%s\n", stringifyTask(task))
	}
	return nil
}

type ClearTasksCmd struct {
	Force bool `help:"proceed without confirmation"`
}

func (c *ClearTasksCmd) Run(deps *GlobalDependencies) error {
	if !c.Force {
		var proceed bool
		err := huh.NewConfirm().Value(&proceed).
			Title("Are you sure you want to clear all tasks?").
			Description("This is destructive and cannot be reversed.").
			Run()
		if err != nil {
			return err
		}

		if !proceed {
			return nil
		}
	}
	return deps.Repo.DeleteAllTasks(context.Background())
}

// Other ------------------------------------------------------------------------

func promptForNewTask() (*model.Tasks, error) {
	var taskNameSuggestions = []string{
		"Write documentation",
		"Weekly review",
		"Project planning",
		"Client meeting",
		"Code refactoring",
		"Declutter email inbox",
		"Research task",
		"Bug fixes",

		// Fun ones
		"Fight procrastination dragon",
		"Tame the inbox monster",
		"World domination plans",
		"Invent time machine",
		"Teach office plant kung-fu",
		"Debug platypus genes",
		"Coffee-driven development",
		"Learn interpretive coding",
		"Teach cat SOLID principles",
		"Quantum bug fixing",
		"Knowledge mining",
		"Mind gardening",
		"Quest for flow state",
	}

	var task model.Tasks
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Value(&task.Name).
				Title("What's the name of the task?").
				Placeholder(taskNameSuggestions[rand.Intn(len(taskNameSuggestions))]).
				Validate(func(s string) error {
					if len(s) == 0 {
						return errors.New("please enter a name")
					}
					return nil
				}),
			huh.NewInput().Value(&task.Description).
				Title("Describe the task.").
				Description("Leave blank to skip."),
		),
	).Run()

	if err != nil {
		return nil, err
	}
	return &task, nil
}

func stringifyTask(task model.Tasks) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Name: %s\n", task.Name))

	if len(task.Description) > 0 {
		sb.WriteString(fmt.Sprintf("Description: %s\n", task.Description))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		fmt.Sprintf("%d. ", task.ID),
		sb.String(),
	)
}
