package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
)

const StateVersion = 1

type SessionState string

const (
	StateWorking  SessionState = "working"
	StateBreaking SessionState = "breaking"
	StateIdle     SessionState = "idle"
)

type CurrentSession struct {
	ID        string       `json:"id"`
	Task      string       `json:"task"`
	StartTime time.Time    `json:"start_time"`
	State     SessionState `json:"state"`
}

type CurrentBreak struct {
	StartTime         time.Time `json:"start_time"`
	SuggestedDuration int       `json:"suggested_duration"`
	ActualDuration    *int      `json:"actual_duration"`
}

type CompletedSession struct {
	ID            string    `json:"id"`
	Task          string    `json:"task"`
	WorkDuration  int       `json:"work_duration"`
	BreakDuration *int      `json:"break_duration"`
	CompletedAt   time.Time `json:"completed_at"`
}

type AppState struct {
	Version           int                `json:"version"`
	CurrentSession    *CurrentSession    `json:"current_session"`
	CurrentBreak      *CurrentBreak      `json:"current_break"`
	CompletedSessions []CompletedSession `json:"completed_sessions"`
}

func getStateFilePath() (string, error) {
	flowerDir := filepath.Join(xdg.DataHome, "flower")
	if err := os.MkdirAll(flowerDir, 0755); err != nil {
		return "", fmt.Errorf("creating flower data directory %q: %w", flowerDir, err)
	}

	return filepath.Join(flowerDir, "state.json"), nil
}

func loadState() (*AppState, error) {
	stateFile, err := getStateFilePath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(stateFile)
	if os.IsNotExist(err) {
		return &AppState{
			Version:           StateVersion,
			CompletedSessions: make([]CompletedSession, 0),
		}, nil
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state AppState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	if state.CompletedSessions == nil {
		state.CompletedSessions = make([]CompletedSession, 0)
	}

	// Handle version migration
	if state.Version <= 0 {
		state.Version = StateVersion
	}

	return &state, nil
}

func saveState(state *AppState) error {
	stateFile, err := getStateFilePath()
	if err != nil {
		return err
	}

	// Ensure version is set
	state.Version = StateVersion

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	tempFile := stateFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp state file: %w", err)
	}

	if err := os.Rename(tempFile, stateFile); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to replace state file: %w", err)
	}

	return nil
}
