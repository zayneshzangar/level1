package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func Unpack(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	var b strings.Builder
	var prev rune
	var havePrev bool
	escaped := false

	for i := 0; i < len(s); {
		r, w := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && w == 1 {
			return "", errors.New("invalid UTF-8 encoding")
		}
		i += w

		if escaped {
			b.WriteRune(r)
			prev = r
			havePrev = true
			escaped = false
			continue
		}

		switch {
		case r == '\\':
			if i >= len(s) {
				return "", errors.New("dangling escape at end of string")
			}
			escaped = true

		case unicode.IsDigit(r):
			if !havePrev {
				return "", errors.New("digit cannot start or follow nothing")
			}
			count := int(r - '0')
			for k := 1; k < count; k++ {
				b.WriteRune(prev)
			}

		default:
			b.WriteRune(r)
			prev = r
			havePrev = true
		}
	}

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
	if result1 == expected1 {
		fmt.Println("test1 is good!")
	}

}
