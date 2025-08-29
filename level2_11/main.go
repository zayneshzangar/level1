package main

import (
	"fmt"
	"sort"
	"strings"
)

// FindAnagrams находит все множества анаграмм в словаре.
func FindAnagrams(words []string) map[string][]string {
	// Карта для группировки анаграмм: ключ — отсортированные буквы, значение — список слов
	groups := make(map[string][]string)
	// Карта для отслеживания уникальных слов в каждой группе
	seen := make(map[string]map[string]bool)

	// Шаг 1: Группируем слова по отсортированным буквам, исключая дубликаты и пустые строки
	for _, word := range words {
		// Приводим слово к нижнему регистру
		lowerWord := strings.ToLower(word)
		// Пропускаем пустые строки
		if lowerWord == "" {
			continue
		}
		// Преобразуем слово в срез рун для корректной работы с Unicode
		runes := []rune(lowerWord)
		// Сортируем руны
		sort.Slice(runes, func(i, j int) bool {
			return runes[i] < runes[j]
		})
		// Преобразуем отсортированные руны обратно в строку — это ключ
		sortedKey := string(runes)
		// Инициализируем карту для текущего ключа, если она ещё не создана
		if _, exists := seen[sortedKey]; !exists {
			seen[sortedKey] = make(map[string]bool)
		}
		// Добавляем слово, только если его ещё нет в группе
		if !seen[sortedKey][lowerWord] {
			groups[sortedKey] = append(groups[sortedKey], lowerWord)
			seen[sortedKey][lowerWord] = true
		}
	}

	// Шаг 2: Формируем результат, исключая множества с одним словом
	result := make(map[string][]string)
	for _, group := range groups {
		if len(group) > 1 { // Пропускаем множества с одним словом
			// Сортируем группу слов по возрастанию
			sort.Strings(group)
			// Ключом будет первое слово множества
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