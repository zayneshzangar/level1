// package main

// import (
// 	"bufio"
// 	"fmt"
// 	"os"
// 	"strconv"
// 	"strings"
// 	"unicode/utf8"
// )

// func main() {
// 	r := bufio.NewReader(os.Stdin)
// 	var n string
// 	fmt.Fscan(r, &n)
// 	var symvol rune

// 	for index, value := range n {
// 		if value > 96 && value < 123 && symvol != 0 {
// 			fmt.Print(string(symvol))
// 		}
// 		if value > 96 && value < 123 {
// 			if index == utf8.RuneCountInString(n)-1 {
// 				fmt.Println(string(value))
// 				break
// 			}
// 			symvol = value
// 		}
// 		if symvol != 0 && value > 47 && value < 58 {
// 			size, err := strconv.Atoi(string(value))
// 			if err != nil {
// 				fmt.Errorf("failed to Atoi func: %w", err)
// 				return
// 			}
// 			fmt.Print(strings.Repeat(string(symvol), size))
// 			symvol = 0
// 		}
// 	}
// 	fmt.Println()

// }

package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Unpack выполняет примитивную распаковку строки.
// Правила:
//   - Буква/руна без цифры -> один раз.
//   - Цифра d после руны -> повторить предыдущую руну так,
//     чтобы общее количество стало d (т.е. добавить d-1 копий).
//   - Обратный слеш '\' экранирует следующий символ (и цифры, и сам слеш).
//   - Пустая строка -> пустая строка.
//
// Ошибки:
//   - Строка начинается с неэкранированной цифры (например, "45").
//   - Висячий слеш в конце (например, "ab\").
//   - Цифра без предыдущей руны (например, "3a").
func Unpack(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	var b strings.Builder
	var prev rune
	var havePrev bool // есть ли уже выведенная предыдущая руна
	escaped := false  // следующий символ трактуем буквально

	for i := 0; i < len(s); {
		r, w := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && w == 1 {
			return "", errors.New("invalid UTF-8 encoding")
		}
		i += w

		if escaped {
			// Трактуем r буквально.
			b.WriteRune(r)
			prev = r
			havePrev = true
			escaped = false
			continue
		}

		switch {
		case r == '\\':
			// Экранируем следующий символ
			if i >= len(s) { // висячий слеш в конце
				return "", errors.New("dangling escape at end of string")
			}
			escaped = true

		case unicode.IsDigit(r):
			// Цифра без предыдущей руны — ошибка
			if !havePrev {
				return "", errors.New("digit cannot start or follow nothing")
			}
			count := int(r - '0')
			// Добавляем ещё (count-1) копий prev (одна уже записана ранее)
			for k := 1; k < count; k++ {
				b.WriteRune(prev)
			}

		default:
			// Обычная руна: сразу выводим
			b.WriteRune(r)
			prev = r
			havePrev = true
		}
	}

	// Если цикл завершился с ожидаемым экранированием — это уже обработано выше.
	return b.String(), nil
}

func main() {
	test1 := "qwe\\45"
	result1, err := Unpack(test1)
	if err != nil {
		fmt.Printf("failed: %v\n", err)
		return
	}
	expected1 := "qwe44444"
	if result1 == expected1{
		fmt.Println("test1 is good!")
	}

}
