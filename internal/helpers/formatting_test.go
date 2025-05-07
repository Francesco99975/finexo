// formatting_test.go
package helpers

import (
	"testing"
)

func TestParseNumberString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		hasError bool
	}{
		{"Simple number", "123", 123, false},
		{"K suffix", "5K", 5000, false},
		{"M suffix", "2.5M", 2500000, false},
		{"B suffix", "1.2B", 1200000000, false},
		{"T suffix", "3T", 3000000000000, false},
		{"Decimal", "123.45", 123, false},
		{"Invalid input", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNumberString(tt.input)

			if tt.hasError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.hasError && result != tt.expected {
				t.Errorf("ParseNumberString(%s) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}
