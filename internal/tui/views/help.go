package views

import (
	"strings"

	"github.com/Broderick-Westrope/flower/internal/tui/styles"
)

// KeyBinding pairs a key label with a short description for the help bar.
type KeyBinding struct {
	Key         string
	Description string
}

// RenderHelpBar formats a slice of key bindings as a single help-bar line.
// Output looks like: [key] desc · [key] desc · [key] desc
func RenderHelpBar(bindings []KeyBinding) string {
	parts := make([]string, len(bindings))
	for i, b := range bindings {
		parts[i] = "[" + b.Key + "] " + b.Description
	}
	return styles.HelpBar.Render(strings.Join(parts, " · "))
}
