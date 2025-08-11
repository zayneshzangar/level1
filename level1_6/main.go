package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"
)

func stopByCondition() {
	stop := false
	go func() {
		for {
			if stop {
				fmt.Println("Горутина завершена (по условию)")
				return
			}
			fmt.Println("Работаю...")
			time.Sleep(500 * time.Millisecond)
		}
	}()
	time.Sleep(2 * time.Second)
	stop = true
	time.Sleep(1 * time.Second)
}

func stopByChannel() {
	stopChan := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopChan:
				fmt.Println("Горутина завершена (через канал)")
				return
			default:
				fmt.Println("Работаю...")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
	time.Sleep(2 * time.Second)
	close(stopChan)
	time.Sleep(1 * time.Second)
}

func stopByContextCancel() {
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Горутина завершена (через context.WithCancel)")
				return
			default:
				fmt.Println("Работаю...")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}(ctx)
	time.Sleep(2 * time.Second)
	cancel()
	time.Sleep(1 * time.Second)
}

func stopByContextTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Горутина завершена (через context.WithTimeout)")
				return
			default:
				fmt.Println("Работаю...")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}(ctx)
	time.Sleep(3 * time.Second)
}

func stopByContextDeadline() {
	deadline := time.Now().Add(2 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Горутина завершена (через context.WithDeadline)")
				return
			default:
				fmt.Println("Работаю...")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}(ctx)
	time.Sleep(3 * time.Second)
}

func stopByGoexit() {
	go func() {
		fmt.Println("Начало работы горутины")
		time.Sleep(1 * time.Second)
		fmt.Println("Завершение через runtime.Goexit()")
		runtime.Goexit()
		fmt.Println("Эта строка не будет выполнена")
	}()
	time.Sleep(2 * time.Second)
}

func stopByTimeAfter() {
	go func() {
		timer := time.After(2 * time.Second)
		for {
			select {
			case <-timer:
				fmt.Println("Горутина завершена (через time.After)")
				return
			default:
				fmt.Println("Работаю...")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
	time.Sleep(3 * time.Second)
}

func stopByPanic() {
	go func() {
		defer fmt.Println("defer: горутина завершена (panic)")
		panic("Что-то пошло не так")
	}()
	time.Sleep(1 * time.Second)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <method_number>")
		fmt.Println("1 - По условию")
		fmt.Println("2 - Через канал")
		fmt.Println("3 - context.WithCancel")
		fmt.Println("4 - context.WithTimeout")
		fmt.Println("5 - context.WithDeadline")
		fmt.Println("6 - runtime.Goexit")
		fmt.Println("7 - time.After")
		fmt.Println("8 - panic (демо)")
		return
	}

	method, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid method number")
		return
	}

	switch method {
	case 1:
		stopByCondition()
	case 2:
		stopByChannel()
	case 3:
		stopByContextCancel()
	case 4:
		stopByContextTimeout()
	case 5:
		stopByContextDeadline()
	case 6:
		stopByGoexit()
	case 7:
		stopByTimeAfter()
	case 8:
		stopByPanic()
	default:
		fmt.Println("Unknown method")
	}
}
