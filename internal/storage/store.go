package storage

import "github.com/Broderick-Westrope/flower/internal/flowtime"

// Store abstracts state persistence.
type Store interface {
	Load() (*flowtime.FlowState, error)
	Save(state *flowtime.FlowState) error
}
