// dtos_test.go
package models

import (
	"reflect"
	"sort"
	"testing"
)

func TestKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]bool
		expected []string
	}{
		{
			name:     "Empty map",
			input:    map[string]bool{},
			expected: []string{},
		},
		{
			name:     "Single key",
			input:    map[string]bool{"key1": true},
			expected: []string{"key1"},
		},
		{
			name:     "Multiple keys",
			input:    map[string]bool{"key1": true, "key2": false, "key3": true},
			expected: []string{"key1", "key2", "key3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := keys(tt.input)

			// Sort both slices to ensure consistent comparison
			if !reflect.DeepEqual(sorted(result), sorted(tt.expected)) {
				t.Errorf("keys(%v) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper function to sort string slices
func sorted(strs []string) []string {
	result := make([]string, len(strs))
	copy(result, strs)
	sort.Strings(result)
	return result
}

func TestValidateIntRange(t *testing.T) {
	tests := []struct {
		name         string
		min          *float64
		max          *float64
		expectMinOut int
		expectMaxOut int
		hasError     bool
	}{
		{
			name:         "Valid range",
			min:          floatPtr(10.0),
			max:          floatPtr(20.0),
			expectMinOut: 10,
			expectMaxOut: 20,
			hasError:     false,
		},
		{
			name:         "Min greater than max",
			min:          floatPtr(30.0),
			max:          floatPtr(20.0),
			expectMinOut: 30,
			expectMaxOut: 20,
			hasError:     true,
		},
		{
			name:         "Nil values",
			min:          nil,
			max:          nil,
			expectMinOut: 0,
			expectMaxOut: 0,
			hasError:     false,
		},
		// Add more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var minOut, maxOut int
			err := ValidateIntRange(tt.min, tt.max, &minOut, &maxOut)

			if tt.hasError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.hasError {
				if minOut/100 != tt.expectMinOut {
					t.Errorf("Expected minOut = %d; got %d", tt.expectMinOut, minOut)
				}

				if maxOut/100 != tt.expectMaxOut {
					t.Errorf("Expected maxOut = %d; got %d", tt.expectMaxOut, maxOut)
				}
			}
		})
	}
}

// Helper function to create float pointer
func floatPtr(v float64) *float64 {
	return &v
}
