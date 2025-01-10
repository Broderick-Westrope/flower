package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Broderick-Westrope/flower/internal"
	"github.com/Broderick-Westrope/flower/internal/data"
	"github.com/fatih/color"
)

type ListSessionsCmd struct {
	Open   bool `help:"Filter only open sessions."`
	Closed bool `help:"Filter only closed sessions."`
	AsJSON bool `name:"json" help:"Display the list as JSON."`
}

func (c *ListSessionsCmd) Run(deps *GlobalDependencies) error {
	ctx := context.Background()

	if !internal.AreMutuallyExclusive(c.Open, c.Closed) {
		fmt.Println(color.RedString("You have provided the following mutually exclusive flags. Please provide only one:"))
		if c.Open {
			fmt.Println(color.RedString(" - open"))
		}
		if c.Closed {
			fmt.Println(color.RedString(" - closed"))
		}
		return nil
	}

	var sessions []data.Session
	var err error
	switch {
	case c.Open:
		sessions, err = deps.Repo.ListOpenSessions(ctx)
	case c.Closed:
		sessions, err = deps.Repo.ListClosedSessions(ctx)
	default:
		sessions, err = deps.Repo.ListSessions(ctx)
	}
	if err != nil {
		return fmt.Errorf("retrieving sessions: %w", err)
	}

	if c.AsJSON {
		jsonBytes, err := json.MarshalIndent(sessions, "", "  ")
		if err != nil {
			return fmt.Errorf("marshalling json: %w", err)
		}
		fmt.Printf("%s\n", jsonBytes)
		return nil
	}

	msg := fmt.Sprintf("%d sessions found:", len(sessions))
	switch len(sessions) {
	case 0:
		fmt.Println("No sessions found.")
		return nil
	case 1:
		msg = "1 session found:"
	}
	fmt.Println(msg)

	for _, s := range sessions {
		state := "Open"
		duration := time.Since(s.StartedAt)
		if s.EndedAt != nil {
			state = "Closed"
			duration = s.EndedAt.Sub(s.StartedAt)
		}
		fmt.Printf("  - Task: %q,\n    Duration: %s,\n    State: %s\n",
			s.Task.Name, duration, state)
	}
	return nil
}
