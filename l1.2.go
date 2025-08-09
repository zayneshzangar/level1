package main

import (
	"fmt"
	"sync"
)

func level_1_2() {
	numbers := []int{2, 4, 6, 8, 10}

	var wg sync.WaitGroup

	for _, num := range numbers {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			fmt.Println(n * n)
		}(num)
	}

	wg.Wait()
}
