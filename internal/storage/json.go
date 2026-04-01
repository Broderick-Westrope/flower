package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Broderick-Westrope/flower/internal/flowtime"
	"github.com/adrg/xdg"
)

const stateVersion = 1

// JSONStore persists FlowState as JSON to the filesystem.
type JSONStore struct {
	clock flowtime.Clock
}

// NewJSONStore creates a new JSONStore that uses the given clock for constructing FlowState.
func NewJSONStore(clock flowtime.Clock) *JSONStore {
	return &JSONStore{clock: clock}
}

// GetFilePath returns the path to the state file, creating the parent directory if needed.
func (s *JSONStore) GetFilePath() (string, error) {
	flowerDir := filepath.Join(xdg.DataHome, "flower")
	if err := os.MkdirAll(flowerDir, 0755); err != nil {
		return "", fmt.Errorf("creating flower data directory %q: %w", flowerDir, err)
	}
	return filepath.Join(flowerDir, "state.json"), nil
}

// JSON serialization types

type jsonSession struct {
	Task      string    `json:"task"`
	StartTime time.Time `json:"start_time"`
}

type jsonBreak struct {
	StartTime         time.Time     `json:"start_time"`
	SuggestedDuration time.Duration `json:"suggested_duration"`
}

type jsonCompletedSession struct {
	Task          string         `json:"task"`
	FlowDuration  time.Duration  `json:"flow_duration"`
	BreakDuration *time.Duration `json:"break_duration"`
	CompletedAt   time.Time      `json:"completed_at"`
}

type jsonState struct {
	Version           int                    `json:"version"`
	CurrentSession    *jsonSession           `json:"current_session"`
	CurrentBreak      *jsonBreak             `json:"current_break"`
	CompletedSessions []jsonCompletedSession `json:"completed_sessions"`
}

// Load reads the state file and returns a FlowState. If the file does not exist, returns a new empty FlowState.
func (s *JSONStore) Load() (*flowtime.FlowState, error) {
	stateFile, err := s.GetFilePath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(stateFile)
	if os.IsNotExist(err) {
		return flowtime.NewFlowState(s.clock), nil
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var js jsonState
	if err := json.Unmarshal(data, &js); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}

	state := flowtime.NewFlowState(s.clock)

	if js.CurrentSession != nil {
		state.CurrentSession = &flowtime.Session{
			Task:      js.CurrentSession.Task,
			StartTime: js.CurrentSession.StartTime,
		}
	}

	if js.CurrentBreak != nil {
		state.CurrentBreak = &flowtime.Break{
			StartTime:         js.CurrentBreak.StartTime,
			SuggestedDuration: js.CurrentBreak.SuggestedDuration,
		}
	}

	for _, cs := range js.CompletedSessions {
		completed := flowtime.CompletedSession{
			Task:         cs.Task,
			FlowDuration: cs.FlowDuration,
			CompletedAt:  cs.CompletedAt,
		}
		if cs.BreakDuration != nil {
			bd := *cs.BreakDuration
			completed.BreakDuration = &bd
		}
		state.CompletedSessions = append(state.CompletedSessions, completed)
	}

	return state, nil
}

// Save writes the FlowState to the state file using atomic temp-file-then-rename.
func (s *JSONStore) Save(state *flowtime.FlowState) error {
	stateFile, err := s.GetFilePath()
	if err != nil {
		return err
	}

	js := jsonState{
		Version:           stateVersion,
		CompletedSessions: make([]jsonCompletedSession, 0, len(state.CompletedSessions)),
	}

	if state.CurrentSession != nil {
		js.CurrentSession = &jsonSession{
			Task:      state.CurrentSession.Task,
			StartTime: state.CurrentSession.StartTime,
		}
	}

	if state.CurrentBreak != nil {
		js.CurrentBreak = &jsonBreak{
			StartTime:         state.CurrentBreak.StartTime,
			SuggestedDuration: state.CurrentBreak.SuggestedDuration,
		}
	}

	for _, cs := range state.CompletedSessions {
		jcs := jsonCompletedSession{
			Task:         cs.Task,
			FlowDuration: cs.FlowDuration,
			CompletedAt:  cs.CompletedAt,
		}
		if cs.BreakDuration != nil {
			bd := *cs.BreakDuration
			jcs.BreakDuration = &bd
		}
		js.CompletedSessions = append(js.CompletedSessions, jcs)
	}

	data, err := json.Marshal(js)
	if err != nil {
		return fmt.Errorf("marshalling state: %w", err)
	}

	tempFile := stateFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("writing temp state file: %w", err)
	}

	if err := os.Rename(tempFile, stateFile); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("replacing state file: %w", err)
	}

	return nil
}
