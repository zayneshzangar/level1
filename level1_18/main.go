package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Counter представляет структуру-счётчик с атомарным инкрементом.
type Counter struct {
	value int32
}

// NewCounter создаёт новый экземпляр счётчика.
func NewCounter() *Counter {
	return &Counter{value: 0}
}

// Increment безопасно увеличивает значение счётчика атомарно.
func (c *Counter) Increment() {
	atomic.AddInt32(&c.value, 1)
}

// GetValue возвращает текущее значение счётчика.
func (c *Counter) GetValue() int32 {
	return atomic.LoadInt32(&c.value)
}

func main() {
	counter := NewCounter()
	var wg sync.WaitGroup

	// Запускаем 100 горутин, каждая увеличивает счётчик 1000 раз
	numGoroutines := 5
	incrementsPerGoroutine := 1000
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				counter.Increment()
			}
		}()
	}

	wg.Wait()
	fmt.Printf("Final counter value: %d\n", counter.GetValue())
	expected := int32(numGoroutines * incrementsPerGoroutine)
	fmt.Printf("Expected value: %d\n", expected)
}
