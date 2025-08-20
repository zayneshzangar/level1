package main

import (
	"fmt"
	"unicode"
)

// hasUniqueChars проверяет, что все символы в строке уникальны с учётом регистра.
func hasUniqueChars(s string) bool {
	seen := make(map[rune]bool)
	for _, char := range s {
		lowerChar := unicode.ToLower(char)
		if seen[lowerChar] {
			return false
		}
		seen[lowerChar] = true
	}
	return true
}

func main() {
	// Тестовые случаи
	testCases := []struct {
		input    string
		expected bool
	}{
		{"abcd", true},
		{"abCdefAaf", false},
		{"aabcd", false},
	}

	for _, tc := range testCases {
		result := hasUniqueChars(tc.input)
		fmt.Printf("String: %q, Unique: %v, Expected: %v\n", tc.input, result, tc.expected)
	}
}
