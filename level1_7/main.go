package main

import (
	"fmt"
	"sync"
)

type SafeMap struct {
	mu sync.Mutex
	m  map[string]int
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		m: make(map[string]int),
	}
}

func (s *SafeMap) Set(key string, value int) {
	s.mu.Lock()
	s.m[key] = value
	s.mu.Unlock()
}

func main() {
	safeMap := NewSafeMap()
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000 ; j++ {
				safeMap.Set("test-key", j)
			}
		}()
	}

	wg.Wait()

	// Читаем значения
	for k, v := range safeMap.m {
		fmt.Println(k, v)
	}
}
