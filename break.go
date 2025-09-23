package main

import (
	"fmt"
	"time"
)

func calculateBreakMinutes(workDuration time.Duration) int {
	workMinutes := int(workDuration.Minutes())
	
	if workMinutes <= 25 {
		return 5
	}
	if workMinutes <= 50 {
		return 8
	}
	if workMinutes <= 90 {
		return 10
	}
	return 15
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func formatDurationShort(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}