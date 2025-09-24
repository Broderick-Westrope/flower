package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
)

const stateVersion = 1

type CurrentSession struct {
	Task      string    `json:"task"`
	StartTime time.Time `json:"start_time"`
}

type CurrentBreak struct {
	StartTime         time.Time     `json:"start_time"`
	SuggestedDuration time.Duration `json:"suggested_duration"`
}

type CompletedSession struct {
	Task          string         `json:"task"`
	FlowDuration  time.Duration  `json:"flow_duration"`
	BreakDuration *time.Duration `json:"break_duration"`
	CompletedAt   time.Time      `json:"completed_at"`
}

type AppState struct {
	Version           int                `json:"version"`
	CurrentSession    *CurrentSession    `json:"current_session"`
	CurrentBreak      *CurrentBreak      `json:"current_break"`
	CompletedSessions []CompletedSession `json:"completed_sessions"`
}

func GetStateFilePath() (string, error) {
	flowerDir := filepath.Join(xdg.DataHome, "flower")
	if err := os.MkdirAll(flowerDir, 0755); err != nil {
		return "", fmt.Errorf("creating flower data directory %q: %w", flowerDir, err)
	}

	return filepath.Join(flowerDir, "state.json"), nil
}

func LoadState() (*AppState, error) {
	stateFile, err := GetStateFilePath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(stateFile)
	if os.IsNotExist(err) {
		return &AppState{
			Version:           stateVersion,
			CompletedSessions: make([]CompletedSession, 0),
		}, nil
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var state AppState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}

	if state.CompletedSessions == nil {
		state.CompletedSessions = make([]CompletedSession, 0)
	}

	// Handle version migration
	if state.Version <= 0 {
		state.Version = stateVersion
	}

	return &state, nil
}

func SaveState(state *AppState) error {
	stateFile, err := GetStateFilePath()
	if err != nil {
		return err
	}

	// Ensure version is set
	state.Version = stateVersion

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshalling state: %w", err)
	}

	tempFile := stateFile + ".tmp"
	err = os.WriteFile(tempFile, data, 0644)
	if err != nil {
		return fmt.Errorf("writing temp state file: %w", err)
	}

	err = os.Rename(tempFile, stateFile)
	if err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("replacing state file: %w", err)
	}

	return nil
}
