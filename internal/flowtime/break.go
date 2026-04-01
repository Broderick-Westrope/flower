package flowtime

import "time"

// CalculateBreak returns a suggested break duration based on how long the work session lasted.
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
