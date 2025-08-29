package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Config хранит параметры утилиты.
type Config struct {
	fields    string // -f: поля для вывода
	delimiter string // -d: разделитель
	separated bool   // -s: только строки с разделителем
}

// parseFields парсит строку с полями и возвращает отсортированный уникальный список номеров полей (начиная с 1).
func parseFields(fieldsStr string) ([]int, error) {
	if fieldsStr == "" {
		return nil, fmt.Errorf("fields are required")
	}

	var fields []int
	parts := strings.Split(fieldsStr, ",")
	for _, part := range parts {
		if strings.Contains(part, "-") {
			// Диапазон, например "3-5"
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range: %s", part)
			}
			start, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid start of range: %s", rangeParts[0])
			}
			end, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid end of range: %s", rangeParts[1])
			}
			if start > end || start < 1 {
				return nil, fmt.Errorf("invalid range: %s", part)
			}
			for i := start; i <= end; i++ {
				fields = append(fields, i)
			}
		} else {
			// Одиночное поле
			num, err := strconv.Atoi(part)
			if err != nil || num < 1 {
				return nil, fmt.Errorf("invalid field: %s", part)
			}
			fields = append(fields, num)
		}
	}

	// Удаляем дубликаты и сортируем
	sort.Ints(fields)
	uniqueFields := fields[:0]
	for i, f := range fields {
		if i == 0 || fields[i-1] != f {
			uniqueFields = append(uniqueFields, f)
		}
	}

	return uniqueFields, nil
}

// cut обрабатывает входной поток согласно конфигурации.
func cut(config *Config, input io.Reader, output io.Writer) error {
	fields, err := parseFields(config.fields)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, config.delimiter)

		if len(parts) == 1 && config.separated {
			continue // Пропускаем строки без разделителя, если -s
		}

		var selected []string
		for _, f := range fields {
			if f-1 < len(parts) {
				selected = append(selected, parts[f-1])
			}
		}

		if len(selected) > 0 {
			fmt.Fprintln(output, strings.Join(selected, config.delimiter))
		} else if !config.separated {
			fmt.Fprintln(output, line) // Выводим оригинальную строку, если нет полей, но не -s
		}
	}

	return scanner.Err()
}

func main() {
	config := &Config{}
	flag.StringVar(&config.fields, "f", "", "select fields (columns)")
	flag.StringVar(&config.delimiter, "d", "\t", "use delimiter instead of TAB")
	flag.BoolVar(&config.separated, "s", false, "only print lines containing delimiter")
	flag.Parse()

	if config.fields == "" {
		fmt.Fprintln(os.Stderr, "Error: -f is required")
		os.Exit(1)
	}

	var input io.Reader = os.Stdin
	if flag.NArg() > 0 {
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		input = file
	}

	if err := cut(config, input, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
