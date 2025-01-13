package env

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// TestParseGeneral tests the getKeyValue function with various valid and invalid key-value pairs.
//
// There is no "export KEY=VAL" support refer to https://forum.djangoproject.com/t/env-files-and-export/11059
func TestParseGeneral(t *testing.T) {
	validMatches := map[string]map[string]string{
		"FOO=bar":            {"FOO": "bar"},
		"FOO =bar":           {"FOO": "bar"},
		"FOO= bar":           {"FOO": "bar"},
		`FOO="bar"`:          {"FOO": "bar"},
		"FOO='bar'":          {"FOO": "bar"},
		`FOO="escaped\"bar"`: {"FOO": `escaped"bar`},
		`FOO="'d'"`:          {"FOO": "'d'"},
		"OPTION_A: 1":        {"OPTION_A": "1"},
		"OPTION_A: Foo=bar":  {"OPTION_A": "Foo=bar"},
		"OPTION_A=1:B":       {"OPTION_A": "1:B"},
		`FOO="bar\nbaz"`:     {"FOO": "bar\nbaz"},
		"FOO=foobar=":        {"FOO": "foobar="},
		"FOO=bar ":           {"FOO": "bar"},
		`KEY=value value`:    {"KEY": "value value"},
		// Comments are all #
		// A # included within " or ' (quotes) is part of the value even if there are spaces
		"FOO=bar # this is foo":        {"FOO": "bar"},
		`FOO="bar#baz" # comment`:      {"FOO": "bar#baz"},
		"FOO='bar#baz' # comment":      {"FOO": "bar#baz"},
		`FOO="bar#baz#bang" # comment`: {"FOO": "bar#baz#bang"},
		`FOO="ba#r"`:                   {"FOO": "ba#r"},
		"FOO='ba#r'":                   {"FOO": "ba#r"},
		`FOO="bar\n\ b\az"`:            {"FOO": "bar\n baz"},
		`FOO="bar\\\n\ b\az"`:          {"FOO": "bar\\\n baz"},
		`FOO="bar\\r\ b\az"`:           {"FOO": "bar\\r baz"},
		`FOO="bar\\\r\ b\az"`:          {"FOO": "bar\\\r baz"},
		" KEY =value":                  {"KEY": "value"},
		"   KEY=value":                 {"KEY": "value"},
		"\tKEY=value":                  {"KEY": "value"},
		"FOO.BAR=foobar":               {"FOO.BAR": "foobar"}, // While dots should not be allowed
	}

	invalidMatches := []string{
		`="value"`, // Must start with a letter
		"=value",
		"value", // No key
		"\n",    // Empty line
		"\r\n",
		"\t\t",
		"# Comment", // Comment only
		"\t # comment",
	}

	for src, expected := range validMatches {
		t.Run(fmt.Sprintf("Valid: %s", src), func(t *testing.T) {
			key, val, _, err := getKeyValue([]byte(src))
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			for k, v := range expected {
				if k != key && v == val {
					t.Errorf("Expected %s, got %s", k, key)
				}

				if k == key && v != val {
					t.Errorf("Expected %s, got %s", v, val)
				}
			}
		})
	}

	for _, src := range invalidMatches {
		t.Run(fmt.Sprintf("Invalid: %s", src), func(t *testing.T) {
			key, val, _, err := getKeyValue([]byte(src))
			if err == nil && key != "" && val != "" {
				t.Errorf("Expected error, got %s=%s", key, val)
			}
		})
	}
}

func createTempFile(t *testing.T, content string) string {
	file, err := os.CreateTemp("", "*.env")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err = file.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err = file.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	return file.Name()
}

func createTempFileBenchmark(b *testing.B, content string) string {
	file, err := os.CreateTemp("", "*.env")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err = file.WriteString(content); err != nil {
		b.Fatalf("Failed to write to temp file: %v", err)
	}

	if err = file.Close(); err != nil {
		b.Fatalf("Failed to close temp file: %v", err)
	}
	return file.Name()
}

func TestParseFromFile(t *testing.T) {
	tests := []struct {
		name      string
		callback  func(key, value string) error
		content   string
		expectErr bool
	}{
		{
			name: "Single file with valid key-value pairs",
			callback: func(key, value string) error {
				return nil
			},
			content:   "KEY=value\nANOTHER_KEY=another_value",
			expectErr: false,
		},
		{
			name: "Single file with invalid key",
			callback: func(key, value string) error {
				return nil
			},
			content:   "key=value",
			expectErr: true,
		},
		{
			name: "Multiple files with valid key-value pairs",
			callback: func(key, value string) error {
				return nil
			},
			content:   "KEY=value\nANOTHER_KEY=another_value",
			expectErr: false,
		},
		{
			name: "No filenames provided",
			callback: func(key, value string) error {
				return nil
			},
			content:   "",
			expectErr: true,
		},
		{
			name: "Callback function returns error",
			callback: func(key, value string) error {
				return errors.New("callback error")
			},
			content:   "KEY=value",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var err error

			if tt.content != "" {
				filename := createTempFile(t, tt.content)
				defer os.Remove(filename)

				err = ParseFromFile(tt.callback, filename)
			} else {
				err = ParseFromFile(tt.callback)
			}
			if tt.expectErr && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestParseFromFileIntoStruct(t *testing.T) {
	type testStruct struct {
		String         string  `env:"STRING"`
		Int            int     `env:"INT"`
		Float          float64 `env:"FLOAT"`
		RequiredString string  `env:"REQUIRED_STRING,required"`
	}

	tests := []struct {
		name       string
		content    string
		assertFunc func(t *testing.T, filename string) error
	}{
		{
			name: "Valid file parses into struct",
			content: `STRING=string
INT=1
FLOAT=1.1
REQUIRED_STRING=required
OPTIONAL_STRING=optional`,
			assertFunc: func(t *testing.T, filename string) error {
				var test testStruct
				err := ParseFromFileIntoStruct(&test, filename)
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
					return err
				}

				if test.String != "string" || test.Int != 1 || test.Float != 1.1 || test.RequiredString != "required" {
					t.Errorf("Unexpected struct values: %+v", test)
				}
				return nil
			},
		},
		{
			name:    "Missing file should return error",
			content: "",
			assertFunc: func(t *testing.T, filename string) error {
				var test testStruct
				err := ParseFromFileIntoStruct(&test)
				if err == nil {
					t.Errorf("Expected error, got nil")
					return errors.New("expected error, got nil")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := createTempFile(t, tt.content)
			defer os.Remove(filename)

			if err := tt.assertFunc(t, filename); err != nil {
				t.Errorf("Test %s failed: %v", tt.name, err)
			}
		})
	}
}

func mockFileOpenerSuccess(content string) FileOpener {
	return func(filename string) (*os.File, error) {
		tmpFile, err := os.CreateTemp("", filename)
		if err != nil {
			return nil, err
		}

		if _, err := tmpFile.WriteString(content); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return nil, err
		}

		// Rewind the file for reading
		_, err = tmpFile.Seek(0, 0)
		if err != nil {
			return nil, err
		}
		return tmpFile, nil
	}
}

func TestParseFile(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		opener    FileOpener
		expected  map[string]string
		expectErr bool
	}{
		{
			name:     "Valid file contents",
			filename: "test.env",
			opener:   mockFileOpenerSuccess("KEY=value\nANOTHER_KEY=another_value"),
			expected: map[string]string{
				"KEY":         "value",
				"ANOTHER_KEY": "another_value",
			},
			expectErr: false,
		},
		{
			name:     "File not found",
			filename: "nonexistent.env",
			opener: func(s string) (*os.File, error) {
				return nil, os.ErrNotExist
			},
			expected:  nil,
			expectErr: true,
		},
		{
			name:     "Invalid file contents",
			filename: "invalid-ssadsad",
			opener: func(s string) (*os.File, error) {
				tmpFile, err := os.CreateTemp("", s)
				if err != nil {
					return nil, err
				}

				tmpFile.Close()

				return tmpFile, err
			},
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFile(tt.filename, tt.opener)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d keys, got %d", len(tt.expected), len(result))
			}

			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("Expected %s=%s, got %s=%s", k, v, k, result[k])
				}
			}
		})
	}
}

func TestReadWithIO(t *testing.T) {
	tests := []struct {
		name      string
		r         io.Reader
		expected  map[string]string
		expectErr bool
	}{
		{
			name: "Valid input with Unix line endings",
			r:    strings.NewReader("KEY=value\nANOTHER_KEY=another_value"),
			expected: map[string]string{
				"KEY":         "value",
				"ANOTHER_KEY": "another_value",
			},
			expectErr: false,
		},
		{
			name: "Valid input with Windows line endings",
			r:    strings.NewReader("KEY=value\r\nANOTHER_KEY=another_value"),
			expected: map[string]string{
				"KEY":         "value",
				"ANOTHER_KEY": "another_value",
			},
			expectErr: false,
		},
		{
			name:      "Empty input",
			r:         strings.NewReader(""),
			expected:  map[string]string{},
			expectErr: true,
		},
		{
			name:      "Invalid input",
			r:         strings.NewReader("KEY value"),
			expected:  map[string]string{},
			expectErr: true,
		},
		{
			name: "Invalid reader",
			r: func() io.Reader {
				r, w := io.Pipe()
				w.CloseWithError(errors.New("test error"))
				return r
			}(),
			expected:  map[string]string{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := readWithIO(tt.r)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d keys, got %d", len(tt.expected), len(result))
			} else {
				for k, v := range tt.expected {
					if result[k] != v {
						t.Errorf("Expected %s=%s, got %s=%s", k, v, k, result[k])
					}
				}
			}
		})
	}
}

func TestParseEnvFileBytes(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		expected  map[string]string
		expectErr bool
	}{
		{
			name:      "Empty input",
			input:     []byte(""),
			expected:  map[string]string{},
			expectErr: true,
		},
		{
			name:  "Valid input with Unix line endings",
			input: []byte("KEY=value\nANOTHER_KEY=another_value"),
			expected: map[string]string{
				"KEY":         "value",
				"ANOTHER_KEY": "another_value",
			},
			expectErr: false,
		},
		{
			name:  "Valid input with Windows line endings",
			input: []byte("KEY=value\r\nANOTHER_KEY=another_value"),
			expected: map[string]string{
				"KEY":         "value",
				"ANOTHER_KEY": "another_value",
			},
			expectErr: false,
		},
		{
			name:      "Invalid input",
			input:     []byte("KEY value"),
			expected:  map[string]string{},
			expectErr: true,
		},
		{
			name:      "Invalid input2",
			input:     []byte("KEY='value"),
			expected:  map[string]string{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseEnvFileBytes(tt.input)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d keys, got %d", len(tt.expected), len(result))
			} else {
				for k, v := range tt.expected {
					if result[k] != v {
						t.Errorf("Expected %s=%s, got %s=%s", k, v, k, result[k])
					}
				}
			}
		})
	}
}

func TestGetStart(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "Empty input",
			input:    []byte(""),
			expected: nil,
		},
		{
			name:     "Whitespace only",
			input:    []byte("   \t\n"),
			expected: nil,
		},
		{
			name:     "Comment line",
			input:    []byte("   # comment\nKEY=value"),
			expected: []byte("KEY=value"),
		},
		{
			name:     "Key-value pair",
			input:    []byte("KEY=value"),
			expected: []byte("KEY=value"),
		},
		{
			name:     "Key-value pair with leading spaces",
			input:    []byte("   KEY=value"),
			expected: []byte("KEY=value"),
		},
		{
			name:     "Multiple lines with comments",
			input:    []byte("   # comment\n   # another comment\nKEY=value"),
			expected: []byte("KEY=value"),
		},
		{
			name:     "No key-value pair",
			input:    []byte("   # comment"),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStart(tt.input)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetKeyValue(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		expectedKey   string
		expectedValue string
		expectErr     bool
	}{
		{
			name:          "Valid key-value pair",
			input:         []byte("KEY=value"),
			expectedKey:   "KEY",
			expectedValue: "value",
			expectErr:     false,
		},
		{
			name:          "Key with spaces",
			input:         []byte("   KEY=value"),
			expectedKey:   "KEY",
			expectedValue: "value",
			expectErr:     false,
		},
		{
			name:          "Value with spaces",
			input:         []byte("KEY=   value"),
			expectedKey:   "KEY",
			expectedValue: "value",
			expectErr:     false,
		},
		{
			name:          "Key with invalid characters",
			input:         []byte("key=value"),
			expectedKey:   "",
			expectedValue: "",
			expectErr:     true,
		},
		{
			name:          "Missing value",
			input:         []byte("KEY="),
			expectedKey:   "KEY",
			expectedValue: "",
			expectErr:     false,
		},
		{
			name:          "Invalid format",
			input:         []byte("KEY value"),
			expectedKey:   "",
			expectedValue: "",
			expectErr:     true,
		},
		{
			name:          "Empty input",
			input:         []byte(""),
			expectedKey:   "",
			expectedValue: "",
			expectErr:     true,
		},
		{
			name:          "Bad quotes",
			input:         []byte(`KEY="value`),
			expectedKey:   "",
			expectedValue: "",
			expectErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, _, err := getKeyValue(tt.input)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if key != tt.expectedKey {
				t.Errorf("Expected key %s, got %s", tt.expectedKey, key)
			}
			if value != tt.expectedValue {
				t.Errorf("Expected value %s, got %s", tt.expectedValue, value)
			}
		})
	}
}

func TestGetValueWithinQuotes(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		quote     byte
		expected  string
		remaining []byte
		expectErr bool
	}{
		{
			name:      "Valid double-quoted value",
			input:     []byte(`"value"`),
			quote:     '"',
			expected:  "value",
			remaining: []byte{},
			expectErr: false,
		},
		{
			name:      "Valid single-quoted value",
			input:     []byte(`'value'`),
			quote:     '\'',
			expected:  "value",
			remaining: []byte{},
			expectErr: false,
		},
		{
			name:      "Escaped double quote",
			input:     []byte(`"value\"with\"quotes"`),
			quote:     '"',
			expected:  `value"with"quotes`,
			remaining: []byte{},
			expectErr: false,
		},
		{
			name:      "Unterminated double quote",
			input:     []byte(`"value`),
			quote:     '"',
			expected:  "",
			remaining: nil,
			expectErr: true,
		},
		{
			name:      "Unterminated single quote",
			input:     []byte(`'value`),
			quote:     '\'',
			expected:  "",
			remaining: nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, remaining, err := getValueWithinQuotes(tt.input, tt.quote)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected value %s, got %s", tt.expected, result)
			}
			if !bytes.Equal(remaining, tt.remaining) {
				t.Errorf("Expected remaining %s, got %s", tt.remaining, remaining)
			}
		})
	}
}

func TestUnescapeQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No escape sequences",
			input:    "simple string",
			expected: "simple string",
		},
		{
			name:     "Newline escape sequence",
			input:    "Line1\\nLine2",
			expected: "Line1\nLine2",
		},
		{
			name:     "Carriage return escape sequence",
			input:    "Line1\\rLine2",
			expected: "Line1\rLine2",
		},
		{
			name:     "Mixed escape sequences",
			input:    "Line1\\nLine2\\rLine3",
			expected: "Line1\nLine2\rLine3",
		},
		{
			name:     "Escaped backslash",
			input:    "Path\\\\to\\\\file",
			expected: "Path\\to\\file",
		},
		{
			name:     "No escape sequences with backslashes",
			input:    "Path\\to\\file",
			expected: "Pathtofile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unescapeQuotes(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetValueWithoutQuotes(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		expected  string
		remaining []byte
		expectErr bool
	}{
		{
			name:      "Valid value without quotes",
			input:     []byte("value\n"),
			expected:  "value",
			remaining: []byte("\n"),
			expectErr: false,
		},
		{
			name:      "Value with spaces",
			input:     []byte("  value  \n"),
			expected:  "value",
			remaining: []byte("\n"),
			expectErr: false,
		},
		{
			name:      "Value with comment",
			input:     []byte("value # comment\n"),
			expected:  "value",
			remaining: []byte("\n"),
			expectErr: false,
		},
		{
			name:      "Empty value",
			input:     []byte("\n"),
			expected:  "",
			remaining: []byte(""),
			expectErr: false,
		},
		{
			name:      "End of file",
			input:     []byte("value"),
			expected:  "value",
			remaining: []byte{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, remaining, err := getValueWithoutQuotes(tt.input)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected value %s, got %s", tt.expected, result)
			}
			if !bytes.Equal(remaining, tt.remaining) {
				t.Errorf("Expected remaining %s, got %s", tt.remaining, remaining)
			}
		})
	}
}

func TestGetKey(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectedKey string
		expectedRem []byte
		expectErr   bool
	}{
		{
			name:        "Valid key-value pair",
			input:       []byte("KEY=value"),
			expectedKey: "KEY",
			expectedRem: []byte("value"),
			expectErr:   false,
		},
		{
			name:        "Key with spaces",
			input:       []byte("   KEY=value"),
			expectedKey: "KEY",
			expectedRem: []byte("value"),
			expectErr:   false,
		},
		{
			name:        "Invalid key",
			input:       []byte("key=value"),
			expectedKey: "",
			expectedRem: []byte("value"),
			expectErr:   true,
		},
		{
			name:        "Empty input",
			input:       []byte(""),
			expectedKey: "",
			expectedRem: nil,
			expectErr:   true,
		},
		{
			name:        "No key-value separator",
			input:       []byte("KEY value"),
			expectedKey: "",
			expectedRem: []byte("KEY value"),
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, remaining, err := getKey(tt.input)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if key != tt.expectedKey {
				t.Errorf("Expected key %s, got %s", tt.expectedKey, key)
			}
			if !bytes.Equal(remaining, tt.expectedRem) {
				t.Errorf("Expected remaining %s, got %s", tt.expectedRem, remaining)
			}
		})
	}
}

func BenchmarkUnescapeQuotes(b *testing.B) {
	// Input string for the benchmark
	input := `Line1\nLine2\rLine3\\nLine4\\rEnd`
	for i := 0; i < b.N; i++ {
		unescapeQuotes(input)
	}
}

func BenchmarkParseFromFile(b *testing.B) {
	content := `KEY1=value1
KEY2=value2
KEY3=value3
KEY4=value4
KEY5=value5`

	filename := createTempFileBenchmark(b, content)
	defer os.Remove(filename)

	callback := func(key, value string) error {
		if key == "" || value == "" {
			return errors.New("key or value is empty")
		} else if key == "KEY1" && value != "value1" {
			return errors.New("unexpected value for KEY1")
		} else {
			return nil
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ParseFromFile(callback, filename)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}
