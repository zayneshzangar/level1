package main

import "fmt"

func main() {
	ch1 := make(chan int)
	ch2 := make(chan int)
	arr := []int{1, 2, 3, 4, 5}
	go func(ch <-chan int) {
		for n := range ch {
			ch2 <- n * 2
		}
		close(ch2)
	}(ch1)

	go func() {
		for value := range ch2 {
			fmt.Println(value)
		}
	}()

	for _, value := range arr {
		ch1 <- value
	}
	close(ch1)
}
