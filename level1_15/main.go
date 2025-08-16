/*
1. v — это строка, но justString = v[:100] сохраняет только первые 100 символов.
   Сборщик мусора не может освободить память под массив байтов, пока justString существует,
   так как она удерживает ссылку на весь массив. даже если v выходит из области видимости (scope) в someFunc,
   массив байтов, на который ссылается justString, не освобождается сборщиком мусора, так как justString
   (глобальная переменная) продолжает ссылаться на него

2. Если len(v) < 100, v[:100] вызовет панику (panic: slice bounds out of range)

3. Без реализации createHugeString длина v неизвестна, что делает поведение кода непредсказуемым
*/

package main

import (
	"fmt"
	"strings"
)

// createHugeString создаёт строку заданной длины, заполненную символом 'a'.
func createHugeString(size int) string {
	return strings.Repeat("a", size)
}

// someFunc возвращает подстроку из первых 100 символов или всю строку, если она короче.
func someFunc(size int) string {
	v := createHugeString(size)
	if len(v) == 0 {
		return ""
	}
	maxLength := 100
	if len(v) < maxLength {
		return string([]byte(v)) // Копируем всю строку
	}
	return string([]byte(v[:maxLength])) // Копируем первые 100 байтов
}

func main() {
	justString := someFunc(150) // Пример размера больше 100
	fmt.Println("justString:", justString)
	fmt.Println("Length of justString:", len(justString))
	justString = someFunc(50) // Пример размера меньше 100
	fmt.Println("justString:", justString)
	fmt.Println("Length of justString:", len(justString))
}
