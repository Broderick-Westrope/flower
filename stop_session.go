package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Broderick-Westrope/flower/internal/data"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
)

type StopSessionCmd struct {
	Latest  bool  `help:"Stop the latest open session."`
	All     bool  `help:"Stop all open sessions."`
	TaskIDs []int `name:"task-ids" help:"Stop all sessions associated with the given task IDs."`
	Force   bool  `help:"Performs the operation without confirming. This must be paired with one of the other flags."`
}

func (c *StopSessionCmd) Run(deps *GlobalDependencies) error {
	ctx := context.Background()

	if !areMutuallyExclusive(c.Latest, c.All, len(c.TaskIDs) > 0) {
		sb := strings.Builder{}
		sb.WriteString("You have provided the following mutually exclusive values. Please provide only one:\n")
		if c.Latest {
			sb.WriteString(fmt.Sprintln(" - latest"))
		}
		if c.All {
			sb.WriteString(fmt.Sprintln(" - all"))
		}
		if len(c.TaskIDs) > 0 {
			sb.WriteString(fmt.Sprintln(" - task-ids"))
		}
		return errors.New(sb.String())
	}

	sessions, err := deps.Repo.GetOpenSessions(ctx)
	if err != nil {
		return fmt.Errorf("retrieving open sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No open sessions found.")
		return nil
	}

	switch {
	case c.Latest:
		fmt.Println("Stopping the latest open session.")
		return c.stopLatestOpenSession(ctx, deps.Repo, sessions)
	case c.All:
		fmt.Println("Stopping all open sessions.")
		return c.stopAllOpenSessions(ctx, deps.Repo, sessions)
	case len(c.TaskIDs) > 0:
		fmt.Printf("Stopping any open sessions associated with task IDs %v.\n", c.TaskIDs)
		return c.stopOpenSessionsAssociatedWithTaskIDs(ctx, deps.Repo, c.TaskIDs, sessions)
	}

	return c.stopOpenSessionsInteractively(ctx, deps.Repo, sessions)
}

func (c *StopSessionCmd) stopLatestOpenSession(ctx context.Context, repo *data.Respository, sessions []data.Session) error {
	session := &sessions[0]
	var err error

	if !c.Force {
		stopSessions := true
		err = huh.NewConfirm().Value(&stopSessions).
			Title("Stop the latest session?").
			Description(fmt.Sprintf("Task: %q, Duration: %s",
				session.Task.Name, time.Since(session.StartedAt).String())).
			Run()
		if err != nil {
			return err
		}
	}

	session, err = repo.StopSession(ctx, session.ID)
	if err != nil {
		return err
	}
	fmt.Printf(color.GreenString(
		"Latest session stopped. The suggested break is %s.\n"),
		calculateBreak(session.EndedAt.Sub(session.StartedAt)).String(),
	)
	return nil
}

func (c *StopSessionCmd) stopAllOpenSessions(ctx context.Context, repo *data.Respository, sessions []data.Session) error {
	var err error
	if !c.Force {
		msg := fmt.Sprintf("Stop all %d open sessions?", len(sessions))
		if len(sessions) == 1 {
			msg = "Stop the one open session?"
		}

		stopSessions := true
		err = huh.NewConfirm().Value(&stopSessions).
			Title(msg).
			Run()
		if err != nil {
			return err
		}
	}

	return stopSessionsAndAnnounce(ctx, repo, sessions)
}

func (c *StopSessionCmd) stopOpenSessionsAssociatedWithTaskIDs(ctx context.Context, repo *data.Respository, taskIDs []int, sessions []data.Session) error {
	filteredSessions := make([]data.Session, 0, len(sessions))
	var belongs bool
	for _, session := range sessions {
		belongs = false
		for _, id := range taskIDs {
			if session.Task.ID == id {
				belongs = true
				break
			}
		}
		if belongs {
			filteredSessions = append(filteredSessions, session)
		}
	}
	sessions = filteredSessions

	if len(sessions) == 0 {
		fmt.Println("No open sessions found that are associated with the given task IDs.")
		return nil
	}

	var err error
	if !c.Force {
		msg := fmt.Sprintf("Stop %d open sessions?", len(sessions))
		if len(sessions) == 1 {
			msg = "Stop the one open session?"
		}

		stopSessions := true
		err = huh.NewConfirm().Value(&stopSessions).
			Title(msg).
			Run()
		if err != nil {
			return err
		}
	}

	return stopSessionsAndAnnounce(ctx, repo, sessions)
}

func (c *StopSessionCmd) stopOpenSessionsInteractively(ctx context.Context, repo *data.Respository, sessions []data.Session) error {
	if len(sessions) == 1 {
		session := sessions[0]
		fmt.Println("Found 1 open session.")
		closeIt := false
		err := huh.NewConfirm().Value(&closeIt).
			Description(fmt.Sprintf("Task: %q, Duration: %s",
				session.Task.Name, time.Since(session.EndedAt).String())).
			Description("").
			Run()
		if err != nil {
			return err
		}
		if !closeIt {
			return nil
		}

		return stopSessionsAndAnnounce(ctx, repo, sessions)
	}

	fmt.Printf("Found %d open sessions.\n", len(sessions))
	return promptToStopOpenSessions(ctx, repo, sessions)
}

func stopSessionsAndAnnounce(ctx context.Context, repo *data.Respository, sessions []data.Session) error {
	var totalBreak time.Duration
	for _, s := range sessions {
		_, err := repo.StopSession(ctx, s.ID)
		if err != nil {
			return err
		}
		totalBreak += calculateBreak(s.EndedAt.Sub(s.StartedAt))
	}

	msg := fmt.Sprintf(
		"%d sessions were stopped. The total break is %s.",
		len(sessions), totalBreak.String(),
	)
	if len(sessions) == 1 {
		msg = fmt.Sprintf(
			"1 session was stopped. The suggested break is %s.", totalBreak.String(),
		)
	}
	fmt.Println(color.GreenString(msg))
	return nil
}

func calculateBreak(sessionDuration time.Duration) time.Duration {
	switch {
	case sessionDuration < time.Minute*25:
		return time.Minute * 5
	case sessionDuration < time.Minute*50:
		return time.Minute * 8
	case sessionDuration < time.Minute*90:
		return time.Minute * 10
	default:
		return time.Minute * 15
	}
}

func promptToStopOpenSessions(ctx context.Context, repo *data.Respository, sessions []data.Session) error {
	closeAll := true
	err := huh.NewConfirm().Value(&closeAll).
		Title("Would you like to stop all of the sessions?").
		Description("If not, you'll be prompted to decide for each session individually.").
		Run()
	if err != nil {
		return err
	}

	if closeAll {
		return stopSessionsAndAnnounce(ctx, repo, sessions)
	}

	stoppedCount := 0
	var totalBreak time.Duration
	var stop bool
	for _, s := range sessions {
		stop = false
		err = huh.NewConfirm().Value(&stop).
			Title("Would you like to stop this session?").
			Description(fmt.Sprintf("Task: %q, Duration: %s", s.Task.Name, time.Since(s.EndedAt).String())).
			Run()
		if err != nil {
			return err
		}

		if !stop {
			continue
		}

		_, err := repo.StopSession(ctx, s.ID)
		if err != nil {
			return err
		}
		stoppedCount++
		totalBreak += calculateBreak(s.EndedAt.Sub(s.StartedAt))
	}
	fmt.Printf("Stopped %d out of %d sessions.\n", stoppedCount, len(sessions))

	msg := fmt.Sprintf(
		"Stopped %d out of %d sessions. The total break is %s.",
		stoppedCount, len(sessions), totalBreak.String(),
	)
	if len(sessions) == 1 {
		msg = fmt.Sprintf(
			"Stopped 1 session. The suggested break is %s.", totalBreak.String(),
		)
	}
	fmt.Println(color.GreenString(msg))

	return nil
}
