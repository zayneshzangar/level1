package main

import "fmt"

// Reverse переворачивает входную строку, поддерживая Unicode-символы.
func Reverse(s string) string {
	runes := []rune(s)
	length := len(runes)
	for i := 0; i < length/2; i++ {
		runes[i], runes[length-1-i] = runes[length-1-i], runes[i]
	}
	return string(runes)
}

func main() {
	input := "главрыба"
	reversed := Reverse(input)
	fmt.Printf("Original: %s\nReversed: %s\n", input, reversed)

	input = "hello 🌍"
	reversed = Reverse(input)
	fmt.Printf("Original: %s\nReversed: %s\n", input, reversed)
}
