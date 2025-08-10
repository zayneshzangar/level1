package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

func main() {
	arg := os.Args
	if len(arg) != 2 {
		fmt.Println("Не правильный ввод аргумента!")
		fmt.Println("Пример: go run main.go 1")
		os.Exit(0)
	}

	num, err := strconv.Atoi(arg[1])
	var wg sync.WaitGroup
	ch := make(chan int, num)

	if err != nil {
		log.Fatalf("Function Atoi: %v", err)
	}
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func() {
			wg.Done()
			fmt.Println(<-ch)
		}()
	}

	for i := 0; i < 100; i++ {
		ch <- i
	}
	wg.Wait()
}
