package main

import "fmt"

// binarySearch ищет элемент в отсортированном срезе и возвращает его индекс или -1, если элемент не найден.
func binarySearch(arr []int, target int) int {
	left := 0
	right := len(arr) - 1

	for left <= right {
		mid := left + (right-left)/2 // Избегаем переполнения

		if arr[mid] == target {
			return mid // Элемент найден
		} else if arr[mid] < target {
			left = mid + 1 // Ищем в правой половине
		} else {
			right = mid - 1 // Ищем в левой половине
		}
	}

	return -1 // Элемент не найден
}

func main() {
	arr := []int{1, 3, 5, 7, 9, 11, 13, 15}
	target := 7
	index := binarySearch(arr, target)
	if index != -1 {
		fmt.Printf("Element %d found at index %d\n", target, index)
	} else {
		fmt.Printf("Element %d not found\n", target)
	}

	target = 6
	index = binarySearch(arr, target)
	if index != -1 {
		fmt.Printf("Element %d found at index %d\n", target, index)
	} else {
		fmt.Printf("Element %d not found\n", target)
	}
}


/*
// binarySearch выполняет рекурсивный бинарный поиск элемента в отсортированном срезе.
func binarySearch(arr []int, target int) int {
    return binarySearchRecursive(arr, target, 0, len(arr)-1)
}

// binarySearchRecursive рекурсивно ищет элемент в заданном диапазоне.
func binarySearchRecursive(arr []int, target, left, right int) int {
    if left > right {
        return -1 // Элемент не найден
    }

    mid := left + (right-left)/2 // Избегаем переполнения

    if arr[mid] == target {
        return mid // Элемент найден
    } else if arr[mid] < target {
        return binarySearchRecursive(arr, target, mid+1, right) // Ищем в правой половине
    } else {
        return binarySearchRecursive(arr, target, left, mid-1) // Ищем в левой половине
    }
}

*/
