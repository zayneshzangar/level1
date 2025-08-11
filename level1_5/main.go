package main

import (
	"fmt"
	"time"
)

func main() {
	t := 5
	ch := make(chan int)

	go func() {
		for val := range ch {
			fmt.Println("Получено:", val)
		}
	}()

	timer := time.After(time.Duration(t) * time.Second)

	counter := 1
	for {
		select {
		case <-timer:
			fmt.Println("Время вышло, завершаем...")
			close(ch)
			return
		default:
			ch <- counter
			counter++
			time.Sleep(500 * time.Millisecond)
		}
	}
}
