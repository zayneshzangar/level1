package main

import (
	"fmt"
)

/*
func main() {
	s := "snow dog sun"
	runes := strings.Split(s, " ")
	slices.Reverse(runes)
	s = strings.Join(runes, " ")
	fmt.Println(s)
}
*/

// ReverseWords переворачивает порядок слов в строке "на месте".
func ReverseWords(s string) string {
	// Преобразуем строку в срез рун для работы с символами
	runes := []rune(s)
	length := len(runes)

	// Шаг 1: Переворачиваем всю строку
	for i := 0; i < length/2; i++ {
		runes[i], runes[length-1-i] = runes[length-1-i], runes[i]
	}

	// Шаг 2: Переворачиваем каждое слово
	start := 0
	for i := 0; i <= length; i++ {
		if i == length || runes[i] == ' ' {
			// Переворачиваем слово от start до i-1
			for j := start; j < (start+i)/2; j++ {
				runes[j], runes[start+i-1-j] = runes[start+i-1-j], runes[j]
			}
			start = i + 1
		}
	}

	return string(runes)
}

func main() {
	input := "snow blur sun"
	reversed := ReverseWords(input)
	fmt.Printf("Original: %s\nReversed: %s\n", input, reversed)

	input = "Mammy world go"
	reversed = ReverseWords(input)
	fmt.Printf("Original: %s\nReversed: %s\n", input, reversed)
}
