package flowtime

import (
	"testing"
	"time"
)

func TestCalculateBreak(t *testing.T) {
	tests := []struct {
		name     string
		work     time.Duration
		expected time.Duration
	}{
		{"0 minutes", 0, 5 * time.Minute},
		{"15 minutes", 15 * time.Minute, 5 * time.Minute},
		{"25 minutes", 25 * time.Minute, 5 * time.Minute},
		{"26 minutes", 26 * time.Minute, 8 * time.Minute},
		{"50 minutes", 50 * time.Minute, 8 * time.Minute},
		{"51 minutes", 51 * time.Minute, 10 * time.Minute},
		{"90 minutes", 90 * time.Minute, 10 * time.Minute},
		{"91 minutes", 91 * time.Minute, 15 * time.Minute},
		{"180 minutes", 180 * time.Minute, 15 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateBreak(tt.work)
			if got != tt.expected {
				t.Errorf("CalculateBreak(%v) = %v, want %v", tt.work, got, tt.expected)
			}
		})
	}
}
