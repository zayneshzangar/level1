package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Config хранит настройки флагов и шаблон.
type Config struct {
	after      int    // -A: строки после
	before     int    // -B: строки до
	context    int    // -C: строки до и после
	countOnly  bool   // -c: только количество совпадений
	ignoreCase bool   // -i: игнорировать регистр
	invert     bool   // -v: инвертировать фильтр
	fixed      bool   // -F: точное совпадение
	lineNumber bool   // -n: выводить номера строк
	pattern    string // шаблон для поиска
}

// Line хранит информацию о строке: текст и номер.
type Line struct {
	text   string
	number int
}

// parseFlags парсит аргументы командной строки и возвращает конфигурацию.
func parseFlags(args []string) (*Config, error) {
	config := &Config{}
	flagSet := flag.NewFlagSet("grep", flag.ExitOnError)
	flagSet.IntVar(&config.after, "A", 0, "print N lines after match")
	flagSet.IntVar(&config.before, "B", 0, "print N lines before match")
	flagSet.IntVar(&config.context, "C", 0, "print N lines of context")
	flagSet.BoolVar(&config.countOnly, "c", false, "print only count of matches")
	flagSet.BoolVar(&config.ignoreCase, "i", false, "ignore case")
	flagSet.BoolVar(&config.invert, "v", false, "invert match")
	flagSet.BoolVar(&config.fixed, "F", false, "fixed string match")
	flagSet.BoolVar(&config.lineNumber, "n", false, "print line numbers")

	if err := flagSet.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %v", err)
	}

	// Проверяем, что шаблон указан
	if flagSet.NArg() < 1 {
		return nil, fmt.Errorf("pattern is required")
	}
	config.pattern = flagSet.Arg(0)

	// Если указан -C, устанавливаем -A и -B равными значению -C
	if config.context > 0 {
		config.after = config.context
		config.before = config.context
	}

	// Проверяем, что значения -A, -B, -C неотрицательные
	if config.after < 0 || config.before < 0 || config.context < 0 {
		return nil, fmt.Errorf("negative values for -A, -B, or -C are not allowed")
	}

	return config, nil
}

// grep выполняет фильтрацию строк согласно конфигурации.
func grep(config *Config, input io.Reader, output io.Writer) error {
	var matcher func(string) bool
	if config.fixed {
		// Для -F используем точное совпадение подстроки
		pattern := config.pattern
		if config.ignoreCase {
			pattern = strings.ToLower(pattern)
		}
		matcher = func(line string) bool {
			if config.ignoreCase {
				return strings.Contains(strings.ToLower(line), pattern)
			}
			return strings.Contains(line, pattern)
		}
	} else {
		// Для регулярных выражений
		pattern := config.pattern
		if config.ignoreCase {
			pattern = "(?i)" + pattern
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %v", err)
		}
		matcher = func(line string) bool {
			return re.MatchString(line)
		}
	}

	// Если -v, инвертируем matcher
	if config.invert {
		origMatcher := matcher
		matcher = func(line string) bool {
			return !origMatcher(line)
		}
	}

	// Читаем строки и сохраняем их для контекста
	scanner := bufio.NewScanner(input)
	lines := []Line{}
	for i := 1; scanner.Scan(); i++ {
		lines = append(lines, Line{text: scanner.Text(), number: i})
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %v", err)
	}

	// Если -c, подсчитываем совпадения и выводим только число
	if config.countOnly {
		count := 0
		for _, line := range lines {
			if matcher(line.text) {
				count++
			}
		}
		fmt.Fprintf(output, "%d\n", count)
		return nil
	}

	// Храним строки для вывода и их номера
	toPrint := make(map[int]bool) // Множество индексов строк для вывода
	for i, line := range lines {
		if matcher(line.text) {
			// Добавляем совпадающую строку
			toPrint[i] = true
			// Добавляем контекст до
			for j := i - 1; j >= i-config.before && j >= 0; j-- {
				toPrint[j] = true
			}
			// Добавляем контекст после
			for j := i + 1; j <= i+config.after && j < len(lines); j++ {
				toPrint[j] = true
			}
		}
	}

	// Выводим строки, избегая дублирования
	lastPrinted := -1
	for i := 0; i < len(lines); i++ {
		if !toPrint[i] {
			continue
		}
		// Если строка не следующая за предыдущей, добавляем разделитель
		if config.after > 0 && lastPrinted >= 0 && i > lastPrinted+1 {
			fmt.Fprintln(output, "--")
		}
		line := lines[i]
		if config.lineNumber {
			fmt.Fprintf(output, "%d:", line.number)
		}
		fmt.Fprintln(output, line.text)
		lastPrinted = i
	}

	return nil
}

func main() {
	config, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	input := os.Stdin
	if flagSet := flag.NewFlagSet("grep", flag.ExitOnError); flagSet.NArg() > 1 {
		file, err := os.Open(flagSet.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		input = file
	}

	if err := grep(config, input, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
