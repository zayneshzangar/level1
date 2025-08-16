package main

import "fmt"

// quickSort сортирует срез целых чисел с помощью алгоритма быстрой сортировки.
// Сложность: в среднем O(n log n), в худшем O(n²) (если массив почти отсортирован).
func quickSort(arr []int) []int {
	if len(arr) <= 1 {
		return arr
	}

	pivot := arr[0] // Опорный элемент — первый
	var left, right []int

	for i := 1; i < len(arr); i++ {
		if arr[i] <= pivot {
			left = append(left, arr[i])
		} else {
			right = append(right, arr[i])
		}
	}

	left = quickSort(left)
	right = quickSort(right)

	// Объединяем: left + [pivot] + right
	result := append(left, pivot)
	result = append(result, right...)
	return result
}

func main() {
	arr := []int{64, 34, 25, 12, 22, 11, 90}
	sorted := quickSort(arr)
	fmt.Println("Sorted array:", sorted)
}
