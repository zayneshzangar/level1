package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func worker(id int, jobs <-chan int) {
	for job := range jobs {
		fmt.Printf("Worker %d got job: %d\n", id, job)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <num_workers>")
		return
	}

	nWorkers, err := strconv.Atoi(os.Args[1])
	if err != nil || nWorkers <= 0 {
		fmt.Println("Invalid number of workers")
		return
	}

	jobs := make(chan int)

	for i := 1; i <= nWorkers; i++ {
		go worker(i, jobs)
	}

	counter := 1
	for {
		jobs <- counter
		counter++
		time.Sleep(500 * time.Millisecond)
	}
	close(jobs)
}
