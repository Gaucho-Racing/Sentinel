package utils

import (
	"testing"
	"time"
)

func TestWithPrecision(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "Round to microsecond precision",
			input:    time.Date(2023, 4, 15, 12, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 4, 15, 12, 30, 45, 123457000, time.UTC),
		},
		{
			name:     "Already at microsecond precision",
			input:    time.Date(2023, 4, 15, 12, 30, 45, 123000000, time.UTC),
			expected: time.Date(2023, 4, 15, 12, 30, 45, 123000000, time.UTC),
		},
		{
			name:     "Round up to next microsecond",
			input:    time.Date(2023, 4, 15, 12, 30, 45, 123999999, time.UTC),
			expected: time.Date(2023, 4, 15, 12, 30, 45, 124000000, time.UTC),
		},
		{
			name:     "Round down to previous microsecond",
			input:    time.Date(2023, 4, 15, 12, 30, 45, 123000001, time.UTC),
			expected: time.Date(2023, 4, 15, 12, 30, 45, 123000000, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WithPrecision(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("WithPrecision(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
