// calculate_test.go
package helpers

import (
	"strings"
	"testing"
)

func TestCompoundingPeriodsPerYear(t *testing.T) {
	tests := []struct {
		name     string
		freq     string
		expected int
	}{
		{"Daily", "daily", 365},
		{"Monthly", "monthly", 12},
		{"Quarterly", "quarterly", 4},
		{"Semi-Annual", "semi-annual", 2},
		{"Annual", "annual", 1},
		{"Default", "unknown", 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compoundingPeriodsPerYear(tt.freq)
			if result != tt.expected {
				t.Errorf("compoundingPeriodsPerYear(%s) = %d; want %d", tt.freq, result, tt.expected)
			}
		})
	}
}

func TestFrequencyToMonths(t *testing.T) {
	tests := []struct {
		name     string
		freq     string
		expected int
	}{
		{"Monthly", "monthly", 1},
		{"Quarterly", "quarterly", 3},
		{"Semi-Annual", "semi-annual", 6},
		{"Annual", "annual", 12},
		{"Default", "unknown", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := frequencyToMonths(tt.freq)
			if result != tt.expected {
				t.Errorf("frequencyToMonths(%s) = %d; want %d", tt.freq, result, tt.expected)
			}
		})
	}
}

func TestCalculateHISAInvestment(t *testing.T) {
	tests := []struct {
		name               string
		principal          float64
		contribution       float64
		contributionFreq   string
		compoundingFreq    string
		annualInterestRate float64
		compoundingYears   int
		currency           string
		expectError        bool
	}{
		{
			name:               "Basic calculation",
			principal:          1000,
			contribution:       100,
			contributionFreq:   "monthly",
			compoundingFreq:    "monthly",
			annualInterestRate: 0.05,
			compoundingYears:   10,
			currency:           "USD",
			expectError:        false,
		},
		{
			name:               "Zero principal",
			principal:          0,
			contribution:       100,
			contributionFreq:   "monthly",
			compoundingFreq:    "monthly",
			annualInterestRate: 0.05,
			compoundingYears:   10,
			currency:           "USD",
			expectError:        false,
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateHISAInvestment(
				tt.principal,
				tt.contribution,
				tt.contributionFreq,
				tt.compoundingFreq,
				tt.annualInterestRate,
				tt.compoundingYears,
				tt.currency,
			)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				// Verify the result has valid values
				if strings.Contains(result.FinalBalance, "-") {
					t.Errorf("Expected positive final balance but got %s", result.FinalBalance)
				}

				// Add more specific validations based on expected outputs
			}
		})
	}
}
