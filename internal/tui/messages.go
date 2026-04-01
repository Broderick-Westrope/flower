package tui

// Message types and Tick are defined in the msgs sub-package to avoid
// import cycles between tui and tui/views. This file re-exports them
// so callers of the tui package can use them directly.

import "github.com/Broderick-Westrope/flower/internal/tui/msgs"

type (
	TickMsg         = msgs.TickMsg
	StartSessionMsg = msgs.StartSessionMsg
	ShowLogMsg      = msgs.ShowLogMsg
	BackMsg         = msgs.BackMsg
	ErrorMsg        = msgs.ErrorMsg
)

var Tick = msgs.Tick
