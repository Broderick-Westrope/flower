package flowtime

import (
	"fmt"
	"strings"
	"time"
)

// FormatDuration formats a duration as a human-readable string like "1h 23m 45s".
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	var output string
	if hours > 0 {
		output += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 {
		output += fmt.Sprintf("%dm ", minutes)
	}
	if seconds > 0 || len(output) == 0 {
		output += fmt.Sprintf("%ds", seconds)
	}
	return strings.TrimSuffix(output, " ")
}

// FormatHumanDateTime formats a timestamp for display relative to now.
// Today shows time only, yesterday shows "Yesterday HH:MM", same year shows "Mon D HH:MM",
// different year shows "Mon D, YYYY HH:MM".
func FormatHumanDateTime(t time.Time, now time.Time) string {
	if t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day() {
		return t.Format("15:04")
	}

	yesterday := now.AddDate(0, 0, -1)
	if t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day() {
		return fmt.Sprintf("Yesterday %s", t.Format("15:04"))
	}

	if t.Year() == now.Year() {
		return t.Format("Jan 2 15:04")
	}

	return t.Format("Jan 2, 2006 15:04")
}
