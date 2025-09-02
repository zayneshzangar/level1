package main

import (
	"fmt"
	"time"
)

func or(channels ...<-chan interface{}) <-chan interface{} {
	out := make(chan interface{})

	if len(channels) == 0 {
		close(out)
		return out
	}
	if len(channels) == 1 {
		return channels[0]
	}

	go func() {
		defer close(out)
		select {
		case <-channels[0]:
		case <-or(channels[1:]...):
		}
	}()

	return out
}

func sig(after time.Duration) <-chan interface{} {
	c := make(chan interface{})
	go func() {
		defer close(c)
		time.Sleep(after)
	}()
	return c
}

func main() {
	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v\n", time.Since(start))
}
