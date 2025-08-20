package main

import (
	"fmt"
	"sync"
	"time"
)

// sleep приостанавливает выполнение текущей горутины на заданное время.
func sleep(duration time.Duration) {
	done := make(chan struct{})
	go func() {
		// Имитация задержки через цикл с проверкой времени
		start := time.Now()
		for {
			if time.Since(start) >= duration {
				close(done)
				break
			}
		}
	}()
	<-done // Блокируем текущую горутину до закрытия канала
}

func main() {
	fmt.Println("Start:", time.Now().Format(time.RFC3339))
	sleep(2 * time.Second)
	fmt.Println("End:", time.Now().Format(time.RFC3339))

	// Дополнительный тест с разным временем
	fmt.Println("Start second test:", time.Now().Format(time.RFC3339))
	sleep(1 * time.Second)
	fmt.Println("End second test:", time.Now().Format(time.RFC3339))

	// Проверка с несколькими горутинами
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Goroutine %d starting at %s\n", id, time.Now().Format(time.RFC3339))
			sleep(time.Duration(id+1) * time.Second)
			fmt.Printf("Goroutine %d ending at %s\n", id, time.Now().Format(time.RFC3339))
		}(i)
	}
	wg.Wait()
}
