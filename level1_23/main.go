package main

import "fmt"

// removeAt удаляет элемент по индексу i из слайса без утечки памяти.
func removeAt(slice []int, i int) []int {
	if i < 0 || i >= len(slice) {
		return slice
	}

	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
}

func main() {
	numbers := []int{10, 20, 30, 40, 50}
	fmt.Println("Original slice:", numbers)

	indexToRemove := 2
	numbers = removeAt(numbers, indexToRemove)
	fmt.Println("After removing element at index", indexToRemove, ":", numbers)

	fmt.Printf("Length: %d, Capacity: %d\n", len(numbers), cap(numbers))
}
