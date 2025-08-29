package main

import (
	"reflect"
	"testing"
)

// TestFindAnagrams тестирует функцию FindAnagrams.
func TestFindAnagrams(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string][]string
	}{
		{
			name: "Basic case with anagrams",
			input: []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"},
			expected: map[string][]string{
				"листок": {"листок", "слиток", "столик"},
				"пятак":  {"пятак", "пятка", "тяпка"},
			},
		},
		{
			name:     "No anagrams",
			input:    []string{"стол", "кот", "дом"},
			expected: map[string][]string{},
		},
		{
			name:     "Empty input",
			input:    []string{},
			expected: map[string][]string{},
		},
		{
			name:     "Mixed case",
			input:    []string{"Пятак", "пЯткА", "ТяПкА", "СТОЛ", "столик"},
			expected: map[string][]string{
				"пятак": {"пятак", "пятка", "тяпка"},
			},
		},
		{
			name:     "Duplicate words",
			input:    []string{"пятак", "пятак", "пятка", "тяпка"},
			expected: map[string][]string{
				"пятак": {"пятак", "пятка", "тяпка"},
			},
		},
		{
			name:     "Empty string",
			input:    []string{"", "пятак", "пятка", ""},
			expected: map[string][]string{
				"пятак": {"пятак", "пятка"},
			},
		},
		{
			name:     "Spaces only",
			input:    []string{"  ", "пятак", "пятка", " "},
			expected: map[string][]string{
				"пятак": {"пятак", "пятка"},
			},
		},
		{
			name:     "Multiple duplicates",
			input:    []string{"пятак", "пятак", "пятак", "пятка", "тяпка"},
			expected: map[string][]string{
				"пятак": {"пятак", "пятка", "тяпка"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindAnagrams(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("FindAnagrams(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}