package main

import "fmt"

// DetectType возвращает строковое представление типа переменной v.
func DetectType(v interface{}) string {
	switch t := v.(type) {
	case int:
		return "int"
	case string:
		return "string"
	case bool:
		return "bool"
	case chan int:
		return "chan int"
	case chan string:
		return "chan string"
	case chan bool:
		return "chan bool"
	default:
		return fmt.Sprintf("unknown type: %T", t)
	}
}

func main() {
	// Примеры переменных разных типов
	var i int = 42
	var s string = "hello"
	var b bool = true
	var ci chan int = make(chan int)
	var cs chan string = make(chan string)
	var cb chan bool = make(chan bool)
	var f float64 = 3.14 // Неподдерживаемый тип для теста

	fmt.Println("Type of i:", DetectType(i))
	fmt.Println("Type of s:", DetectType(s))
	fmt.Println("Type of b:", DetectType(b))
	fmt.Println("Type of ci:", DetectType(ci))
	fmt.Println("Type of cs:", DetectType(cs))
	fmt.Println("Type of cb:", DetectType(cb))
	fmt.Println("Type of f:", DetectType(f))
}
