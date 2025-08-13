package main

import "fmt"

func intersection(a []string) []string {
	set := make(map[string]bool)
	var result []string

	// Запоминаем элементы первого множества
	for _, v := range a {
		if set[v] {
			continue
		}
		result = append(result, v)
		set[v] = true
	}

	return result
}

func main() {
	A := []string{"cat", "cat", "dog", "cat", "tree"}

	fmt.Println(intersection(A)) // [2 3]
}
