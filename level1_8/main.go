package main

import (
	"fmt"
)

func setBit(num int64, pos uint, value int) int64 {
	if value == 1 {
		// Установить бит в 1
		num |= (1 << pos)
	} else {
		// Установить бит в 0
		num &^= (1 << pos)
	}
	return num
}

func main() {
	var num int64 = 5 // 0101₂

	fmt.Printf("Исходное число: %d (%04b)\n", num, num)

	// Устанавливаем 1-й бит в 0
	num = setBit(num, 1, 0)
	fmt.Printf("После установки бита 1 в 0: %d (%04b)\n", num, num)

	// Устанавливаем 2-й бит в 1
	num = setBit(num, 2, 1)
	fmt.Printf("После установки бита 2 в 1: %d (%04b)\n", num, num)
}
