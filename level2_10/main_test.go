package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestSort тестирует различные сценарии сортировки через запуск программы.
func TestSort(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		args      []string
		expected  string
		wantErr   bool
		useStderr bool // Проверяем stderr вместо stdout
	}{
		{
			name:      "Sort by first column (-k 0)",
			input:     "r\ta\tn\nc\td\tm\nj\tq\te\n1\t3\t2\n",
			args:      []string{"-k", "0", "test_input.txt"},
			expected:  "1\t3\t2\nc\td\tm\nj\tq\te\nr\ta\tn\n",
			wantErr:   false,
			useStderr: false,
		},
		{
			name:      "Reverse sort (-r)",
			input:     "r\ta\tn\nc\td\tm\nj\tq\te\n",
			args:      []string{"-r", "test_input.txt"},
			expected:  "r\ta\tn\nj\tq\te\nc\td\tm\n",
			wantErr:   false,
			useStderr: false,
		},
		{
			name:      "Numeric sort by first column (-k 0 -n)",
			input:     "10\ta\tn\n5\td\tm\n20\tq\te\n",
			args:      []string{"-k", "0", "-n", "test_input.txt"},
			expected:  "5\td\tm\n10\ta\tn\n20\tq\te\n",
			wantErr:   false,
			useStderr: false,
		},
		{
			name:      "Unique lines (-u)",
			input:     "c\ta\tn\nc\td\tm\nc\ta\tn\nj\tq\te\n",
			args:      []string{"-u", "test_input.txt"},
			expected:  "c\ta\tn\nc\td\tm\nj\tq\te\n",
			wantErr:   false,
			useStderr: false,
		},
		{
			name:      "Human-readable sizes (-k 0 -h)",
			input:     "10K\ta\tn\n2M\td\tm\n500\tq\te\n",
			args:      []string{"-k", "0", "-h", "test_input.txt"},
			expected:  "500\tq\te\n10K\ta\tn\n2M\td\tm\n",
			wantErr:   false,
			useStderr: false,
		},
		{
			name:      "Month sort (-k 0 -M)",
			input:     "Feb\ta\tn\nDec\td\tm\nJan\tq\te\n",
			args:      []string{"-k", "0", "-M", "test_input.txt"},
			expected:  "Jan\tq\te\nFeb\ta\tn\nDec\td\tm\n",
			wantErr:   false,
			useStderr: false,
		},
		{
			name:      "Ignore trailing blanks (-k 0 -b)",
			input:     "r\ta\tn\t \nc\td\tm\nj\tq\te\n",
			args:      []string{"-k", "0", "-b", "test_input.txt"},
			expected:  "c\td\tm\nj\tq\te\nr\ta\tn\t \n",
			wantErr:   false,
			useStderr: false,
		},
		{
			name:      "Check sorted input (-c)",
			input:     "1\ta\tn\nc\td\tm\nj\tq\te\n",
			args:      []string{"-k", "0", "-c", "test_input.txt"},
			expected:  "Input is sorted\n",
			wantErr:   false,
			useStderr: false,
		},
		{
			name:      "Check unsorted input (-c)",
			input:     "r\ta\tn\nc\td\tm\nj\tq\te\n",
			args:      []string{"-k", "0", "-c", "test_input.txt"},
			expected:  "Input is not sorted\n",
			wantErr:   true,
			useStderr: true,
		},
		{
			name:      "Sort by file input",
			input:     "r\ta\tn\nc\td\tm\nj\tq\te\n1\t3\t2\n",
			args:      []string{"-k", "0", "test_input.txt"},
			expected:  "1\t3\t2\nc\td\tm\nj\tq\te\nr\ta\tn\n",
			wantErr:   false,
			useStderr: false,
		},
		{
			name:      "Invalid column (-k 10)",
			input:     "r\ta\tn\nc\td\tm\n",
			args:      []string{"-k", "10", "test_input.txt"},
			expected:  "r\ta\tn\nc\td\tm\n",
			wantErr:   false,
			useStderr: false,
		},
		{
			name:      "Empty input",
			input:     "",
			args:      []string{"-k", "0", "test_input.txt"},
			expected:  "",
			wantErr:   false,
			useStderr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем временный файл для всех тестов
			inputFile := "test_input.txt"
			err := os.WriteFile(inputFile, []byte(tt.input), 0644)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(inputFile)

			// Запускаем команду
			cmd := exec.Command("go", append([]string{"run", "main.go"}, tt.args...)...)
			var out bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr

			err = cmd.Run()
			if (err != nil) != tt.wantErr {
				t.Errorf("Expected error: %v, got: %v, stderr: %s", tt.wantErr, err, stderr.String())
			}

			// Проверяем вывод (stdout или stderr в зависимости от useStderr)
			got := out.String()
			if tt.useStderr {
				got = stderr.String()
				// Для теста Check_unsorted_input_(-c) проверяем, что ожидаемая строка содержится в выводе
				if !strings.Contains(got, tt.expected) {
					t.Errorf("Expected output to contain:\n%s\nGot:\n%s", tt.expected, got)
				}
			} else if got != tt.expected {
				t.Errorf("Expected output:\n%s\nGot:\n%s", tt.expected, got)
			}
		})
	}
}

// TestParseHumanSize тестирует функцию parseHumanSize.
func TestParseHumanSize(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"10", 10},
		{"10K", 10 * 1024},
		{"2M", 2 * 1024 * 1024},
		{"invalid", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("parseHumanSize(%s)", tt.input), func(t *testing.T) {
			got := parseHumanSize(tt.input)
			if got != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, got)
			}
		})
	}
}

// TestGetKey тестирует функцию getKey.
func TestGetKey(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		line     string
		expected string
	}{
		{
			name:     "Whole line",
			config:   Config{column: -1},
			line:     "r\ta\tn",
			expected: "r\ta\tn",
		},
		{
			name:     "Column 0",
			config:   Config{column: 0},
			line:     "r\ta\tn",
			expected: "r",
		},
		{
			name:     "Numeric sort",
			config:   Config{column: 0, numeric: true},
			line:     "10\ta\tn",
			expected: fmt.Sprintf("%020f", 10.0),
		},
		{
			name:     "Human-readable size",
			config:   Config{column: 0, human: true},
			line:     "10K\ta\tn",
			expected: fmt.Sprintf("%020f", 10.0*1024),
		},
		{
			name:     "Month sort",
			config:   Config{column: 0, month: true},
			line:     "Feb\ta\tn",
			expected: "02",
		},
		{
			name:     "Ignore blanks",
			config:   Config{column: -1, ignoreBlanks: true},
			line:     "r\ta\tn\t ",
			expected: "r\ta\tn",
		},
		{
			name:     "Invalid column",
			config:   Config{column: 10},
			line:     "r\ta\tn",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.getKey(tt.line)
			if got != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}
		})
	}
}