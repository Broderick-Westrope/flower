package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Broderick-Westrope/flower/internal/data"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
)

type StartSessionCmd struct {
	TaskID string `arg:"" optional:"" help:"Task ID to associate with this session."`
	Skip   bool   `help:"Skip checking for open sessions."`
}

func (c *StartSessionCmd) Run(deps *GlobalDependencies) error {
	ctx := context.Background()

	if !c.Skip {
		err := c.checkForOpenSessions(ctx, deps.Repo)
		if err != nil {
			return fmt.Errorf("checking for open sessions: %w", err)
		}
	}

	var taskID int
	var err error
	if len(c.TaskID) == 0 {
		addTaskCmd := AddTaskCmd{}
		err = addTaskCmd.Run(deps)
		if err != nil {
			return fmt.Errorf("running add task command: %w", err)
		}
		taskID = addTaskCmd.taskID
	} else {
		taskID, err = strconv.Atoi(c.TaskID)
		if err != nil {
			return errors.New("the provided task ID must be an integer (number)")
		}

		_, err = deps.Repo.GetTask(ctx, taskID)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				fmt.Printf(color.RedString(
					"Task with ID %d not found. Please check the ID and try again.\n",
				), taskID)
				return nil
			}
			return fmt.Errorf("check if task exists: %w", err)
		}
	}

	session, err := deps.Repo.StartSession(ctx, taskID)
	if err != nil {
		return nil
	}
	fmt.Printf(color.GreenString("Session started for task %q.\n"), session.Task.Name)
	return nil
}

func (c *StartSessionCmd) checkForOpenSessions(ctx context.Context, repo *data.Respository) error {
	sessions, err := repo.GetOpenSessions(ctx)
	if err != nil {
		return fmt.Errorf("retrieving open sessions: %w", err)
	}

	switch len(sessions) {
	case 0:
		return nil

	case 1:
		fmt.Println(color.YellowString(
			"There is still an open session. This should be stopped before starting a new session."),
		)

		stopNow := true
		err := huh.NewConfirm().Value(&stopNow).
			Title("Would you like to stop it now?").
			Run()
		if err != nil {
			return fmt.Errorf("prompting to stop open session: %w", err)
		}

		if stopNow {
			_, err = repo.StopSession(ctx, sessions[0].ID)
			if err != nil {
				return err
			}
		}
		return nil

	default:
		fmt.Printf(color.YellowString(
			"There are still %d open sessions. These should be stopped before starting a new session.\n"),
			len(sessions),
		)

		leaveOpen := true
		err := huh.NewConfirm().Value(&leaveOpen).
			Title("Would you like to leave these sessions open?").
			Run()
		if err != nil {
			return err
		}

		if leaveOpen {
			return nil
		}

		err = promptToStopOpenSessions(ctx, repo, sessions)
		if err != nil {
			return fmt.Errorf("prompting to stop open sessions: %w", err)
		}
		return nil
	}
}
