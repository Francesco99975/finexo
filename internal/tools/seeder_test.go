// seeder_test.go
package tools

import (
	"testing"
)

func TestNormalizeSeed(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Uppercase", "aapl", "AAPL"},
		{"Trim spaces", " AAPL ", "AAPL"},
		{"Remove spaces", "A A P L", "AAPL"},
		{"Toronto exchange", "RY-TO", "RY.TO"},
		{"NEO exchange", "AAPL-NE", "AAPL.NE"},
		{"London exchange", "BP-L", "BP.L"},
		{"Vancouver exchange", "ABC-V", "ABC.V"},
		{"Tokyo exchange", "7203-T", "7203.T"},
		{"Milan exchange", "ENI-MI", "ENI.MI"},
		{"Frankfurt exchange", "BMW-F", "BMW.F"},
		{"Period to dash", "AAPL.US", "AAPL-US"},
		{"Multiple conversions", "eni-mi-f", "ENI.MI.F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeSeed(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeSeed(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
