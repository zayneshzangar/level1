package main

import "fmt"

// SwapNumbers меняет местами два числа a и b с помощью XOR.
func SwapNumbers(a, b *int) {
	*a = *a ^ *b
	*b = *a ^ *b
	*a = *a ^ *b
}

func main() {
	a, b := 5, 10
	fmt.Printf("Before: a=%d, b=%d\n", a, b)
	SwapNumbers(&a, &b)
	fmt.Printf("After: a=%d, b=%d\n", a, b)
}
