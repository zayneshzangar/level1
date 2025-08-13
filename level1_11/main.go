package main

import "fmt"

func intersection(a, b []int) []int {
	set := make(map[int]bool)
	var result []int

	// Запоминаем элементы первого множества
	for _, v := range a {
		set[v] = true
	}

	// Проверяем элементы второго множества
	for _, v := range b {
		if set[v] {
			result = append(result, v)
			delete(set, v) // чтобы избежать дубликатов в результате
		}
	}

	return result
}

func main() {
	A := []int{1, 2, 3}
	B := []int{2, 3, 4}

	fmt.Println(intersection(A, B)) // [2 3]
}
