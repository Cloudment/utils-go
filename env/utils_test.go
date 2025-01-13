package env

import (
	"errors"
	"reflect"
	"testing"
)

func TestIsSliceOfStructs(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.StructField
		expected bool
	}{
		{
			name: "Slice of structs",
			field: reflect.TypeOf(struct {
				Field []struct{}
			}{}).Field(0),
			expected: true,
		},
		{
			name: "Slice of non-structs",
			field: reflect.TypeOf(struct {
				Field []int
			}{}).Field(0),
			expected: false,
		},
		{
			name: "Non-slice type",
			field: reflect.TypeOf(struct {
				Field int
			}{}).Field(0),
			expected: false,
		},
		{
			name: "Pointer to slice of structs",
			field: reflect.TypeOf(struct {
				Field *[]struct{}
			}{}).Field(0),
			expected: true,
		},
		{
			name: "Pointer to non-slice type",
			field: reflect.TypeOf(struct {
				Field *int
			}{}).Field(0),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isSliceOfStructs(tc.field)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

type validTextUnmarshaler struct {
	value string
}

func (v *validTextUnmarshaler) UnmarshalText(text []byte) error {
	v.value = string(text)
	return nil
}

type invalidTextUnmarshaler struct{}

func (v *invalidTextUnmarshaler) UnmarshalText(text []byte) error {
	return errors.New("unmarshal error")
}

func TestAsTextUnmarshaler(t *testing.T) {
	tests := []struct {
		name     string
		v        reflect.Value
		expected bool
	}{
		{
			name:     "Valid TextUnmarshaler",
			v:        reflect.ValueOf(&validTextUnmarshaler{}),
			expected: true,
		},
		{
			name:     "Invalid TextUnmarshaler",
			v:        reflect.ValueOf(&invalidTextUnmarshaler{}),
			expected: true,
		},
		{
			name:     "Non-TextUnmarshaler",
			v:        reflect.ValueOf(&struct{}{}),
			expected: false,
		},
		{
			name:     "Nil value",
			v:        reflect.ValueOf(nil),
			expected: false,
		},
		{
			name:     "Non-pointer but addressable",
			v:        reflect.ValueOf(validTextUnmarshaler{}),
			expected: false,
		},
		{
			name:     "Pointer",
			v:        reflect.ValueOf(&validTextUnmarshaler{}).Elem(),
			expected: true,
		},
		{
			name:     "Nil pointer",
			v:        reflect.ValueOf((*validTextUnmarshaler)(nil)).Elem(),
			expected: false,
		},
		{
			name:     "Int pointer, with invalid value",
			v:        reflect.New(reflect.TypeOf((*int)(nil))).Elem(),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := asTextUnmarshaler(tc.v)
			if (result != nil) != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result != nil)
			}
		})
	}
}

func TestInitialisePointer(t *testing.T) {
	tests := []struct {
		name     string
		v        reflect.Value
		expected interface{}
	}{
		{
			name:     "Non-nil pointer to struct",
			v:        reflect.ValueOf(&struct{}{}),
			expected: &struct{}{},
		},
		{
			name:     "Non-pointer value",
			v:        reflect.ValueOf(struct{}{}),
			expected: struct{}{},
		},
		{
			name:     "Int pointer, with invalid value",
			v:        reflect.New(reflect.TypeOf((*int)(nil))).Elem(),
			expected: func() *int { i := 0; return &i }(), // It's a pointer value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialisePointer(tt.v)
			if !reflect.DeepEqual(tt.v.Interface(), tt.expected) {
				t.Errorf("initialisePointer() = %v, expected %v", tt.v.Interface(), tt.expected)
			}
		})
	}
}

func TestUpdateReference(t *testing.T) {
	tests := []struct {
		name     string
		ref      reflect.Value
		result   reflect.Value
		expected interface{}
	}{
		{
			name:     "Update non pointer to int",
			ref:      reflect.ValueOf(new(int)).Elem(),
			result:   reflect.ValueOf(42),
			expected: 42,
		},
		{
			name:     "Update pointer to struct",
			ref:      reflect.ValueOf(new(struct{})),
			result:   reflect.ValueOf(struct{}{}),
			expected: &struct{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateReference(tt.ref, tt.result)
			if !reflect.DeepEqual(tt.ref.Interface(), tt.expected) {
				t.Errorf("updateReference() = %v, expected %v", tt.ref.Interface(), tt.expected)
			}
		})
	}
}

func TestResolvePointer(t *testing.T) {
	tests := []struct {
		name     string
		v        reflect.Value
		sfType   reflect.Type
		expected reflect.Value
	}{
		{
			name:     "Non-pointer value",
			v:        reflect.ValueOf(42),
			sfType:   reflect.TypeOf(42),
			expected: reflect.ValueOf(42),
		},
		{
			name:     "Pointer to value",
			v:        reflect.ValueOf(new(int)).Elem(),
			sfType:   reflect.TypeOf(new(int)).Elem(),
			expected: reflect.ValueOf(new(int)).Elem(),
		},
		{
			name:     "Pointer to struct",
			v:        reflect.ValueOf(&struct{ Field int }{Field: 42}),
			sfType:   reflect.TypeOf(&struct{ Field int }{}),
			expected: reflect.ValueOf(struct{ Field int }{Field: 42}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := resolvePointer(tt.v, tt.sfType)
			if !reflect.DeepEqual(result.Interface(), tt.expected.Interface()) {
				t.Errorf("resolvePointer() = %v, expected %v", result.Interface(), tt.expected.Interface())
			}
		})
	}
}

func TestEnsureTrailingUnderscore(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		expected string
	}{
		{
			name:     "Empty string",
			prefix:   "",
			expected: "",
		},
		{
			name:     "String without underscore",
			prefix:   "PREFIX",
			expected: "PREFIX_",
		},
		{
			name:     "String with trailing underscore",
			prefix:   "PREFIX_",
			expected: "PREFIX_",
		},
		{
			name:     "String with multiple underscores",
			prefix:   "PREFIX__",
			expected: "PREFIX__",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ensureTrailingUnderscore(tt.prefix)
			if result != tt.expected {
				t.Errorf("ensureTrailingUnderscore(%q) = %q, expected %q", tt.prefix, result, tt.expected)
			}
		})
	}
}

func TestFindMaxIndex(t *testing.T) {
	tests := []struct {
		name           string
		prefixedEnvMap map[int]bool
		expected       int
	}{
		{
			name:           "Empty map",
			prefixedEnvMap: map[int]bool{},
			expected:       0,
		},
		{
			name:           "Single element map",
			prefixedEnvMap: map[int]bool{1: true},
			expected:       1,
		},
		{
			name:           "Multiple elements map",
			prefixedEnvMap: map[int]bool{1: true, 3: true, 2: true},
			expected:       3,
		},
		{
			name:           "Non-sequential elements map",
			prefixedEnvMap: map[int]bool{10: true, 5: true, 7: true},
			expected:       10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findMaxIndex(tt.prefixedEnvMap)
			if result != tt.expected {
				t.Errorf("findMaxIndex() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetSeparators(t *testing.T) {
	tests := []struct {
		name              string
		field             reflect.StructField
		expectedSep       string
		expectedKeyValSep string
	}{
		{
			name: "Default separators",
			field: reflect.TypeOf(struct {
				Field string ``
			}{}).Field(0),
			expectedSep:       ",",
			expectedKeyValSep: ":",
		},
		{
			name: "Custom separators",
			field: reflect.TypeOf(struct {
				Field string `envSeparator:";" envKeyValSeparator:"="`
			}{}).Field(0),
			expectedSep:       ";",
			expectedKeyValSep: "=",
		},
		{
			name: "Custom separator only",
			field: reflect.TypeOf(struct {
				Field string `envSeparator:"|"`
			}{}).Field(0),
			expectedSep:       "|",
			expectedKeyValSep: ":",
		},
		{
			name: "Custom key-value separator only",
			field: reflect.TypeOf(struct {
				Field string `envKeyValSeparator:"-"`
			}{}).Field(0),
			expectedSep:       ",",
			expectedKeyValSep: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sep, keyValSep := getSeparators(tt.field)
			if sep != tt.expectedSep {
				t.Errorf("getSeparators() envSeparator = %v, expected %v", sep, tt.expectedSep)
			}
			if keyValSep != tt.expectedKeyValSep {
				t.Errorf("getSeparators() envKeyValSeparator = %v, expected %v", keyValSep, tt.expectedKeyValSep)
			}
		})
	}
}

func TestHasQuotePrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected byte
		hasQuote bool
	}{
		{
			name:     "Empty input",
			input:    []byte(""),
			expected: 0,
			hasQuote: false,
		},
		{
			name:     "No quote prefix",
			input:    []byte("value"),
			expected: 0,
			hasQuote: false,
		},
		{
			name:     "Double quote prefix",
			input:    []byte(`"value"`),
			expected: '"',
			hasQuote: true,
		},
		{
			name:     "Single quote prefix",
			input:    []byte(`'value'`),
			expected: '\'',
			hasQuote: true,
		},
		{
			name:     "Whitespace before quote",
			input:    []byte(`  "value"`),
			expected: 0,
			hasQuote: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quote, hasQuote := hasQuotePrefix(tt.input)
			if quote != tt.expected || hasQuote != tt.hasQuote {
				t.Errorf("Expected (%c, %v), got (%c, %v)", tt.expected, tt.hasQuote, quote, hasQuote)
			}
		})
	}
}

func TestIndexOfChar(t *testing.T) {
	tests := []struct {
		name     string
		src      []byte
		c        rune
		expected int
	}{
		{
			name:     "Character present",
			src:      []byte("hello"),
			c:        'e',
			expected: 1,
		},
		{
			name:     "Character not present",
			src:      []byte("hello"),
			c:        'x',
			expected: -1,
		},
		{
			name:     "Empty source",
			src:      []byte(""),
			c:        'a',
			expected: -1,
		},
		{
			name:     "Character at start",
			src:      []byte("abc"),
			c:        'a',
			expected: 0,
		},
		{
			name:     "Character at end",
			src:      []byte("abc"),
			c:        'c',
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indexOfChar(tt.src, tt.c)
			if result != tt.expected {
				t.Errorf("indexOfChar() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIndexOfChars(t *testing.T) {
	tests := []struct {
		name     string
		src      []byte
		chars    []rune
		expected int
	}{
		{
			name:     "Character present",
			src:      []byte("hello"),
			chars:    []rune{'e', 'o'},
			expected: 1,
		},
		{
			name:     "Character not present",
			src:      []byte("hello"),
			chars:    []rune{'x', 'y'},
			expected: -1,
		},
		{
			name:     "Empty source",
			src:      []byte(""),
			chars:    []rune{'a', 'b'},
			expected: -1,
		},
		{
			name:     "Character at start",
			src:      []byte("abc"),
			chars:    []rune{'a', 'b'},
			expected: 0,
		},
		{
			name:     "Character at end",
			src:      []byte("abc"),
			chars:    []rune{'c', 'd'},
			expected: 2,
		},
		{
			name:     "Multiple characters present",
			src:      []byte("abcdef"),
			chars:    []rune{'d', 'e', 'f'},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indexOfChars(tt.src, tt.chars...)
			if result != tt.expected {
				t.Errorf("indexOfChars() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIndexOfNonSpaceChar(t *testing.T) {
	tests := []struct {
		name     string
		src      []byte
		expected int
	}{
		{
			name:     "Empty input",
			src:      []byte(""),
			expected: -1,
		},
		{
			name:     "No non-space character",
			src:      []byte(" \t\n"),
			expected: -1,
		},
		{
			name:     "Non-space character at start",
			src:      []byte("a \t\n"),
			expected: 0,
		},
		{
			name:     "Non-space character in middle",
			src:      []byte(" \t\na"),
			expected: 3,
		},
		{
			name:     "Non-space character at end",
			src:      []byte(" \t\n a"),
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indexOfNonSpaceChar(tt.src)
			if result != tt.expected {
				t.Errorf("indexOfNonSpaceChar() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsSpace(t *testing.T) {
	tests := []struct {
		name     string
		input    rune
		expected bool
	}{
		{
			name:     "Tab character",
			input:    '\t',
			expected: true,
		},
		{
			name:     "Vertical tab character",
			input:    '\v',
			expected: true,
		},
		{
			name:     "Form feed character",
			input:    '\f',
			expected: true,
		},
		{
			name:     "Carriage return character",
			input:    '\r',
			expected: true,
		},
		{
			name:     "Space character",
			input:    ' ',
			expected: true,
		},
		{
			name:     "Next line character",
			input:    0x85,
			expected: true,
		},
		{
			name:     "Non-breaking space character",
			input:    0xA0,
			expected: true,
		},
		{
			name:     "Newline character",
			input:    '\n',
			expected: false,
		},
		{
			name:     "Alphabet character",
			input:    'a',
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSpace(tt.input)
			if result != tt.expected {
				t.Errorf("isSpace(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
