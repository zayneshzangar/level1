package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Config хранит параметры сортировки.
type Config struct {
	column      int    // Номер столбца для сортировки (-k, 0-based)
	numeric     bool   // Числовая сортировка (-n)
	reverse     bool   // Обратный порядок (-r)
	unique      bool   // Уникальные строки (-u)
	month       bool   // Сортировка по месяцам (-M)
	ignoreBlanks bool   // Игнорировать хвостовые пробелы (-b)
	check       bool   // Проверка отсортированности (-c)
	human       bool   // Человекочитаемые размеры (-h)
}

// parseHumanSize преобразует строку с суффиксами (K, M) в число.
func parseHumanSize(s string) float64 {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0
	}
	multiplier := 1.0
	if strings.HasSuffix(s, "K") {
		multiplier = 1024
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "M") {
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	}
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0 // Если не число, возвращаем 0
	}
	return num * multiplier
}

// byColumn реализует sort.Interface для кастомной сортировки.
type byColumn struct {
	lines  []string
	config Config
	getKey func(string) string
}

func (bc *byColumn) Len() int           { return len(bc.lines) }
func (bc *byColumn) Swap(i, j int)      { bc.lines[i], bc.lines[j] = bc.lines[j], bc.lines[i] }
func (bc *byColumn) Less(i, j int) bool {
	a := bc.getKey(bc.lines[i])
	b := bc.getKey(bc.lines[j])
	if bc.config.reverse {
		return a > b
	}
	return a < b
}

// getKey возвращает ключ сортировки в зависимости от флагов.
func (c Config) getKey(line string) string {
	if c.ignoreBlanks {
		line = strings.TrimRight(line, " \t")
	}
	if c.column >= 0 {
		columns := strings.Split(line, "\t")
		if len(columns) <= c.column {
			return ""
		}
		line = columns[c.column]
	}
	if c.numeric {
		num, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
		if err != nil {
			return line // Если не число, сортируем как строку
		}
		return fmt.Sprintf("%020f", num) // Паддинг для корректной сортировки
	}
	if c.human {
		return fmt.Sprintf("%020f", parseHumanSize(line))
	}
	if c.month {
		month := strings.ToLower(strings.TrimSpace(line))
		months := []string{"jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"}
		for i, m := range months {
			if m == month {
				return fmt.Sprintf("%02d", i+1)
			}
		}
		return "00"
	}
	return line
}

func main() {
	// Парсинг флагов
	config := Config{}
	flag.IntVar(&config.column, "k", -1, "sort by column number (0-based, tab-separated)")
	flag.BoolVar(&config.numeric, "n", false, "sort numerically")
	flag.BoolVar(&config.reverse, "r", false, "reverse sort order")
	flag.BoolVar(&config.unique, "u", false, "output only unique lines")
	flag.BoolVar(&config.month, "M", false, "sort by month name")
	flag.BoolVar(&config.ignoreBlanks, "b", false, "ignore trailing blanks")
	flag.BoolVar(&config.check, "c", false, "check if already sorted")
	flag.BoolVar(&config.human, "h", false, "sort by human-readable sizes (K, M)")
	flag.Parse()

	// Чтение данных
	var lines []string
	input := os.Stdin
	args := flag.Args()
	if len(args) > 0 {
		f, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		input = f
	}

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" { // Пропускаем пустые строки
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	if len(lines) == 0 {
		return
	}

	// Проверка отсортированности
	if config.check {
		sorted := true
		for i := 1; i < len(lines); i++ {
			if config.getKey(lines[i]) < config.getKey(lines[i-1]) {
				sorted = false
				break
			}
		}
		if !sorted {
			fmt.Fprintf(os.Stderr, "Input is not sorted\n")
			os.Exit(1)
		}
		fmt.Println("Input is sorted")
		return
	}

	// Сортировка
	sorter := byColumn{lines: lines, config: config, getKey: config.getKey}
	sort.Sort(&sorter)

	// Удаление дубликатов
	if config.unique {
		uniqueLines := []string{lines[0]}
		for i := 1; i < len(lines); i++ {
			if lines[i] != lines[i-1] { // Сравниваем полные строки для уникальности
				uniqueLines = append(uniqueLines, lines[i])
			}
		}
		lines = uniqueLines
	}

	// Вывод
	for _, line := range lines {
		fmt.Println(line)
	}
}