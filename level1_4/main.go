package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func worker(ctx context.Context, id int, jobs <-chan int) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d завершает работу\n", id)
			return
		case job := <-jobs:
			fmt.Printf("Worker %d получил задачу: %d\n", id, job)
		}
	}
}

func main() {
	nWorkers := 5

	// Создаём контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для заданий
	jobs := make(chan int)

	// Запускаем воркеров
	for i := 1; i <= nWorkers; i++ {
		go worker(ctx, i, jobs)
	}

	// Канал для получения сигнала ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	// Запуск горутины, которая слушает SIGINT
	go func() {
		<-sigChan
		fmt.Println("\nПолучен сигнал завершения, останавливаем воркеров...")
		cancel() // отменяем контекст
		close(jobs)
	}()

	// Главный цикл
	counter := 1
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Главная горутина завершает работу")
			return
		default:
			jobs <- counter
			counter++
			time.Sleep(500 * time.Millisecond)
		}
	}
}
