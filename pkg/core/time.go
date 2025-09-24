package core

import (
	"fmt"
	"strings"
	"time"
)

func CalculateBreak(workDuration time.Duration) time.Duration {
	workMinutes := int(workDuration.Minutes())

	if workMinutes <= 25 {
		return 5 * time.Minute
	}
	if workMinutes <= 50 {
		return 8 * time.Minute
	}
	if workMinutes <= 90 {
		return 10 * time.Minute
	}
	return 15 * time.Minute
}

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

// FormatHumanDateTime formats a timestamp for display, showing date only if not today.
func FormatHumanDateTime(t time.Time) string {
	now := time.Now()

	// Check if the date is today by comparing year, month, and day.
	if t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day() {
		return t.Format("15:04")
	}

	// Check if the date is yesterday.
	yesterday := now.AddDate(0, 0, -1)
	if t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day() {
		return fmt.Sprintf("Yesterday %s", t.Format("15:04"))
	}

	// Check if the date is within the current year.
	if t.Year() == now.Year() {
		return t.Format("Jan 2 15:04")
	}

	// Full date for different years.
	return t.Format("Jan 2, 2006 15:04")
}
