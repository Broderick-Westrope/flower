package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/Broderick-Westrope/flower/internal/data"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
)

type GetTaskCmd struct {
	TaskID int  `arg:"" help:"Task ID."`
	AsJSON bool `name:"json" help:"Display the list as JSON"`
}

func (c *GetTaskCmd) Run(deps *GlobalDependencies) error {
	task, err := deps.Repo.GetTask(context.Background(), c.TaskID)
	if err != nil {
		return err
	}

	if c.AsJSON {
		jsonBytes, err := json.MarshalIndent(task, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", jsonBytes)
		return nil
	}

	fmt.Printf("\n%s\n", stringifyTask(task))
	return nil
}

type AddTaskCmd struct {
	Name        string `help:"Task name."`
	Description string `aliases:"desc" help:"Task description."`
	ParentID    string `name:"parent-id" help:"Task ID of the parent task."`

	// This is used by other commands that want to interactively create a task.
	taskID int
}

func (c *AddTaskCmd) Run(deps *GlobalDependencies) error {
	ctx := context.Background()
	var task *data.Task
	var err error
	if len(c.Name) == 0 {
		fmt.Println("Creating a new task...")
		task, err = promptForNewTask(ctx, deps.Repo)
		if err != nil {
			return fmt.Errorf("prompting for new task: %w", err)
		}
	} else {
		task = &data.Task{
			Name:        c.Name,
			Description: c.Description,
		}
	}

	if task == nil {
		return errors.New("task object was nil")
	}

	task, err = deps.Repo.CreateTask(ctx, task)
	if err != nil {
		return err
	}

	fmt.Printf(color.GreenString("New task added with ID %d.\n"), task.ID)
	c.taskID = task.ID
	return nil
}

type RemoveTaskCmd struct {
	TaskID int  `arg:"" help:"ID of the task to remove."`
	Force  bool `help:"Confirm deletion without prompting."`
}

func (c *RemoveTaskCmd) Run(deps *GlobalDependencies) error {
	ctx := context.Background()
	err := deps.Repo.DeleteTask(ctx, c.TaskID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			fmt.Printf(color.RedString("Task with ID %d not found.\n"), c.TaskID)
			return nil
		}
		return fmt.Errorf("deleting task: %w", err)
	}
	fmt.Println(color.GreenString("Task removed."))
	return nil
}

type ListTasksCmd struct {
	AsJSON bool `name:"json" help:"Display the list as JSON."`
}

func (c *ListTasksCmd) Run(deps *GlobalDependencies) error {
	ctx := context.Background()
	tasks, err := deps.Repo.ListTasks(ctx)
	if err != nil {
		return fmt.Errorf("retrieving tasks: %w", err)
	}

	if c.AsJSON {
		jsonBytes, err := json.MarshalIndent(tasks, "", "  ")
		if err != nil {
			return fmt.Errorf("marshalling json: %w", err)
		}
		fmt.Printf("%s\n", jsonBytes)
		return nil
	}

	for _, task := range tasks {
		fmt.Println(stringifyTask(&task))
	}
	return nil
}

type ClearTasksCmd struct {
	Force bool `help:"Confirm deletion without prompting."`
}

func (c *ClearTasksCmd) Run(deps *GlobalDependencies) error {
	if !c.Force {
		var proceed bool
		err := huh.NewConfirm().Value(&proceed).
			Title("Are you sure you want to clear all tasks?").
			Description("This is destructive and cannot be reversed.").
			Run()
		if err != nil {
			return fmt.Errorf("confirming clearing tasks: %w", err)
		}

		if !proceed {
			return nil
		}
	}

	err := deps.Repo.DeleteAllTasks(context.Background())
	if err != nil {
		return fmt.Errorf("deleting all tasks: %w", err)
	}
	fmt.Println(color.GreenString("Tasks cleared."))
	return nil
}

func promptForNewTask(ctx context.Context, repo *data.Repository) (*data.Task, error) {
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
		"Refactor spaghetti monster",
		"Train AI to make tea",
		"Conquer the laundry mountain",
		"Reverse-engineer toaster intelligence",
		"Pixel-perfect unicorn designs",
		"Optimise sandwich algorithms",
		"Deploy rocket-powered ducks",
		"Master the art of WiFi summoning",
		"Paint happy little bugs",
		"Solve Rubik's cube of life",
		"Document the dark matter API",
		"Assemble IKEA time machine",
		"Fix time-travel paradoxes",
		"Write documentation in haiku",
		"Brew potion of productivity",
		"Unlock the secrets of YAML",
		"Build empathy-driven robots",
		"Complete 10,000-hour procrastination course",
		"Outsource chores to squirrels",
		"Encrypt the secret of happiness",
		"Launch the procrastination-free startup",
		"Train dog to debug pipelines",
		"Publish manifesto on snack-driven development",
		"Overclock the office coffee machine",
		"Host team-building dragon hunt",
		"Invent wireless hugs",
	}

	var task data.Task
	linkParent := false
	parentIDStr := ""
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Value(&task.Name).
				Title("What's the name of the task?").
				Placeholder(taskNameSuggestions[rand.Intn(len(taskNameSuggestions))]).
				Validate(func(s string) error {
					if len(s) == 0 {
						return errors.New("Name is required.")
					}
					return nil
				}),
			huh.NewInput().Value(&task.Description).
				Title("Describe the task.").
				Description("Leave blank to skip."),
			huh.NewConfirm().Value(&linkParent).
				Title("Would you like to link a parent task?").
				Description("This allows you to create a logical heirachy of tasks."),
		),
		huh.NewGroup(
			huh.NewInput().Value(&parentIDStr).
				Title("What's the ID of the parent task?").
				Validate(func(s string) error {
					id, err := strconv.Atoi(s)
					if err != nil {
						return errors.New("Parent ID must be an integer.")
					}
					_, err = repo.GetTask(ctx, id)
					if err != nil {
						if errors.Is(err, data.ErrNotFound) {
							return fmt.Errorf("No task with ID %d found.", id)
						}
						return errors.New("Failed to verify task exists, please try again.")
					}
					cyclic, err := repo.DetectParentTaskCycle(ctx, id)
					if err != nil {
						return errors.New("Failed to check cycles, please try again.")
					}
					if cyclic {
						return errors.New("This would create a cyclic relationship between tasks.")
					}
					return nil
				}),
		).WithHideFunc(func() bool { return !linkParent }),
	).Run()

	if err != nil {
		return nil, err
	}

	if linkParent && len(parentIDStr) > 0 {
		parentID, err := strconv.Atoi(parentIDStr)
		if err != nil {
			return nil, err
		}
		task.Parent = &data.Task{
			ID: parentID,
		}
	}
	return &task, nil
}

func stringifyTask(task *data.Task) string {
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
