// enums_test.go
package models

import (
	"testing"
)

func TestParseFrequency(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Frequency
		hasError bool
	}{
		{"Weekly", "weekly", FrequencyWeekly, false},
		{"Biweekly", "biweekly", FrequencyBiweekly, false},
		{"Monthly", "monthly", FrequencyMonthly, false},
		{"Quarterly", "quarterly", FrequencyQuarterly, false},
		{"Semi-Annual", "semi-annual", FrequencySemi, false},
		{"Annual", "annual", FrequencyYearly, false},
		{"Case insensitive", "Monthly", FrequencyMonthly, false},
		{"Invalid", "invalid", FrequencyUnknown, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFrequency(tt.input)

			if tt.hasError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.hasError && result != tt.expected {
				t.Errorf("ParseFrequency(%s) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}
