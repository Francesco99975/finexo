// parsers_test.go
package tools

import (
	"testing"
)

func TestIsAnEmptyString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Empty string", "", true},
		{"Whitespace only", "   ", true},
		{"N/A value", "N/A", true},
		{"n/a lowercase", "n/a", true},
		{"Dash", "-", true},
		{"Double dash", "--", true},
		{"Normal text", "Hello", false},
		{"Text with spaces", "  Hello  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAnEmptyString(tt.input)
			if result != tt.expected {
				t.Errorf("isAnEmptyString(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractPercentage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"With percentage", "Price $100 (+5%)", "+5%"},
		{"No parentheses", "Price $100", ""},
		{"Empty parentheses", "Price $()", ""},
		{"Multiple parentheses", "Price $100 (+5%) (yesterday)", "+5%"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPercentage(tt.input)
			if result != tt.expected {
				t.Errorf("extractPercentage(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTickerExtractor(t *testing.T) {
	tests := []struct {
		name             string
		seed             string
		expectedTicker   string
		expectedExchange string
		hasError         bool
	}{
		{"Colon format", "NYSE:AAPL", "AAPL", "NYSE", false},
		{"Dot format", "AAPL.NYSE", "AAPL", "NYSE", false},
		{"Just ticker", "AAPL", "AAPL", "", false},
		{"With slash", "NYSE/AAPL", "AAPL", "NYSE", false},
		{"Lowercase", "nyse:aapl", "AAPL", "NYSE", false},
		{"Empty", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticker, exchange, err := tickerExtractor(tt.seed)

			if tt.hasError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.hasError {
				if ticker != tt.expectedTicker {
					t.Errorf("Expected ticker = %q; got %q", tt.expectedTicker, ticker)
				}

				if exchange != tt.expectedExchange {
					t.Errorf("Expected exchange = %q; got %q", tt.expectedExchange, exchange)
				}
			}
		})
	}
}
