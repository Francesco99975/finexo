// limiters_test.go
package middlewares

import (
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		array    []string
		str      string
		expected bool
	}{
		{"Found in array", []string{"apple", "banana", "cherry"}, "banana", true},
		{"Not found in array", []string{"apple", "banana", "cherry"}, "grape", false},
		{"Empty array", []string{}, "anything", false},
		{"Empty string", []string{"apple", "banana", "cherry"}, "", false},
		{"Case sensitive", []string{"Apple", "Banana", "Cherry"}, "apple", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.array, tt.str)
			if result != tt.expected {
				t.Errorf("contains(%v, %q) = %v; want %v", tt.array, tt.str, result, tt.expected)
			}
		})
	}
}
