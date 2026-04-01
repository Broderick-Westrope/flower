package flowtime

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero", 0, "0s"},
		{"seconds only", 45 * time.Second, "45s"},
		{"minutes and seconds", 5*time.Minute + 30*time.Second, "5m 30s"},
		{"hours minutes seconds", 1*time.Hour + 23*time.Minute + 45*time.Second, "1h 23m 45s"},
		{"hours only", 1 * time.Hour, "1h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.duration)
			if got != tt.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, got, tt.expected)
			}
		})
	}
}

func TestFormatHumanDateTime(t *testing.T) {
	now := time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC)

	t.Run("today shows time only", func(t *testing.T) {
		target := time.Date(2025, 6, 15, 9, 45, 0, 0, time.UTC)
		got := FormatHumanDateTime(target, now)
		expected := "09:45"
		if got != expected {
			t.Errorf("FormatHumanDateTime = %q, want %q", got, expected)
		}
	})

	t.Run("yesterday shows Yesterday prefix", func(t *testing.T) {
		target := time.Date(2025, 6, 14, 18, 30, 0, 0, time.UTC)
		got := FormatHumanDateTime(target, now)
		expected := "Yesterday 18:30"
		if got != expected {
			t.Errorf("FormatHumanDateTime = %q, want %q", got, expected)
		}
	})

	t.Run("this year shows month day time", func(t *testing.T) {
		target := time.Date(2025, 3, 10, 12, 0, 0, 0, time.UTC)
		got := FormatHumanDateTime(target, now)
		expected := "Mar 10 12:00"
		if got != expected {
			t.Errorf("FormatHumanDateTime = %q, want %q", got, expected)
		}
	})

	t.Run("different year shows full date", func(t *testing.T) {
		target := time.Date(2023, 12, 25, 8, 0, 0, 0, time.UTC)
		got := FormatHumanDateTime(target, now)
		expected := "Dec 25, 2023 08:00"
		if got != expected {
			t.Errorf("FormatHumanDateTime = %q, want %q", got, expected)
		}
	})
}
