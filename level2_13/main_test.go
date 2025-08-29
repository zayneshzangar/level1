package main

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

// TestCut тестирует функцию cut с различными комбинациями флагов.
func TestCut(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Basic -f 1",
			config:   &Config{fields: "1", delimiter: "\t", separated: false},
			input:    "field1\tfield2\tfield3\nfield4\tfield5\tfield6\n",
			expected: "field1\nfield4\n",
			wantErr:  false,
		},
		{
			name:     "Range -f 2-3",
			config:   &Config{fields: "2-3", delimiter: "\t", separated: false},
			input:    "field1\tfield2\tfield3\nfield4\tfield5\tfield6\n",
			expected: "field2\tfield3\nfield5\tfield6\n",
			wantErr:  false,
		},
		{
			name:     "Mixed -f 1,3-4",
			config:   &Config{fields: "1,3-4", delimiter: "\t", separated: false},
			input:    "field1\tfield2\tfield3\tfield4\n",
			expected: "field1\tfield3\tfield4\n",
			wantErr:  false,
		},
		{
			name:     "Fields out of range",
			config:   &Config{fields: "1,5", delimiter: "\t", separated: false},
			input:    "field1\tfield2\tfield3\n",
			expected: "field1\n",
			wantErr:  false,
		},
		{
			name:     "With -s",
			config:   &Config{fields: "1", delimiter: "\t", separated: true},
			input:    "field1\nfield2\tfield3\nfield4\n",
			expected: "field2\n",
			wantErr:  false,
		},
		{
			name:     "Custom delimiter -d ,",
			config:   &Config{fields: "2", delimiter: ",", separated: false},
			input:    "field1,field2,field3\n",
			expected: "field2\n",
			wantErr:  false,
		},
		{
			name:     "Empty input",
			config:   &Config{fields: "1", delimiter: "\t", separated: false},
			input:    "",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "No fields - error",
			config:   &Config{fields: "", delimiter: "\t", separated: false},
			input:    "field1\tfield2\n",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Invalid fields",
			config:   &Config{fields: "invalid", delimiter: "\t", separated: false},
			input:    "field1\tfield2\n",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Negative range - error",
			config:   &Config{fields: "1- -2", delimiter: "\t", separated: false},
			input:    "field1\tfield2\n",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := bytes.NewReader([]byte(tt.input))
			var output bytes.Buffer
			err := cut(tt.config, input, &output)
			if (err != nil) != tt.wantErr {
				t.Errorf("cut() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := output.String(); got != tt.expected {
				t.Errorf("cut() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestParseFields тестирует функцию parseFields.
func TestParseFields(t *testing.T) {
	tests := []struct {
		name      string
		fieldsStr string
		expected  []int
		wantErr   bool
	}{
		{
			name:      "Single field",
			fieldsStr: "1",
			expected:  []int{1},
			wantErr:   false,
		},
		{
			name:      "Multiple fields",
			fieldsStr: "1,3,5",
			expected:  []int{1, 3, 5},
			wantErr:   false,
		},
		{
			name:      "Range",
			fieldsStr: "2-4",
			expected:  []int{2, 3, 4},
			wantErr:   false,
		},
		{
			name:      "Mixed",
			fieldsStr: "1,3-5,7",
			expected:  []int{1, 3, 4, 5, 7},
			wantErr:   false,
		},
		{
			name:      "Duplicates",
			fieldsStr: "1,1,2-3,3",
			expected:  []int{1, 2, 3},
			wantErr:   false,
		},
		{
			name:      "Invalid field",
			fieldsStr: "invalid",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "Negative field",
			fieldsStr: "-1",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "Invalid range",
			fieldsStr: "5-2",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "Empty",
			fieldsStr: "",
			expected:  nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFields(tt.fieldsStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseFields() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestMainFile тестирует утилиту с чтением из файла.
func TestMainFile(t *testing.T) {
	content := "field1\tfield2\tfield3\nfield4\tfield5\tfield6\n"
	tmpfile, err := os.CreateTemp("", "cut_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	config := &Config{fields: "1-2", delimiter: "\t", separated: false}
	file, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	var output bytes.Buffer
	err = cut(config, file, &output)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "field1\tfield2\nfield4\tfield5\n"
	if got := output.String(); got != expected {
		t.Errorf("Expected output:\n%q\nGot:\n%q", expected, got)
	}
}
