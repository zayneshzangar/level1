package main

import (
	"fmt"
	"sort"
	"strings"
)

// FindAnagrams находит все множества анаграмм в словаре.
func FindAnagrams(words []string) map[string][]string {
	groups := make(map[string][]string)
	seen := make(map[string]map[string]bool)

	for _, word := range words {
		lowerWord := strings.ToLower(word)
		if lowerWord == "" {
			continue
		}
		runes := []rune(lowerWord)
		sort.Slice(runes, func(i, j int) bool {
			return runes[i] < runes[j]
		})

		sortedKey := string(runes)
		if _, exists := seen[sortedKey]; !exists {
			seen[sortedKey] = make(map[string]bool)
		}

		if !seen[sortedKey][lowerWord] {
			groups[sortedKey] = append(groups[sortedKey], lowerWord)
			seen[sortedKey][lowerWord] = true
		}
	}

	result := make(map[string][]string)
	for _, group := range groups {
		if len(group) > 1 {
			sort.Strings(group)
			result[group[0]] = group
		}
	}

	return result
}

func main() {
	words := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	result := FindAnagrams(words)
	for key, group := range result {
		fmt.Printf("%s: %v\n", key, group)
	}
}