package env

import (
	"encoding"
	"errors"
	"fmt"
	"github.com/cloudment/utils-go/utils"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestParsers(t *testing.T) {
	tests := []struct {
		kind   reflect.Kind
		input  string
		output interface{}
		hasErr bool
	}{
		{reflect.Bool, "true", true, false},
		{reflect.Int, "123", int(123), false},
		{reflect.Int8, "123", int8(123), false},
		{reflect.Int16, "123", int16(123), false},
		{reflect.Int32, "123", int32(123), false},
		{reflect.Int64, "123", int64(123), false},
		{reflect.Uint, "123", uint(123), false},
		{reflect.Uint8, "123", uint8(123), false},
		{reflect.Uint16, "123", uint16(123), false},
		{reflect.Uint32, "123", uint32(123), false},
		{reflect.Uint64, "123", uint64(123), false},
		{reflect.Float32, "123.45", float32(123.45), false},
		{reflect.Float64, "123.45", float64(123.45), false},
		{reflect.String, "test", "test", false},
		{reflect.Int, "invalid", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			parser, ok := parsers[tt.kind]
			if !ok {
				t.Fatalf("No parser found for kind %v", tt.kind)
			}

			result, err := parser(tt.input)
			if (err != nil) != tt.hasErr {
				t.Errorf("Expected error: %v, got: %v", tt.hasErr, err)
			}

			if !tt.hasErr && !reflect.DeepEqual(result, tt.output) {
				t.Errorf("Expected output: %v, got: %v", tt.output, result)
			}
		})
	}
}

func TestTypeParsers(t *testing.T) {
	// bad duration
	_, err := typeParsers[reflect.TypeOf(time.Nanosecond)]("bad")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// good duration
	d, err := typeParsers[reflect.TypeOf(time.Nanosecond)]("1s")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	} else if d != time.Second {
		t.Errorf("Expected 1s, got %v", d)
	}

	// bad location
	_, err = typeParsers[reflect.TypeOf(time.Location{})]("bad")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// good location
	loc, err := typeParsers[reflect.TypeOf(time.Location{})]("UTC")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	goodLoc, _ := time.LoadLocation("UTC")
	if !utils.IsEqual(loc, *goodLoc) {
		t.Errorf("Expected %v, got %v", *goodLoc, loc)
	}
}

func TestHandleSpecialTypes(t *testing.T) {
	tests := []struct {
		name string
		v    reflect.Value
		val  string
		sf   reflect.StructField
		err  bool
	}{
		{
			"Slice",
			reflect.ValueOf(&[]int{}).Elem(),
			"1,2,3",
			reflect.StructField{Type: reflect.TypeOf([]int{})},
			false,
		},
		{
			"Map",
			reflect.ValueOf(&map[string]int{}).Elem(),
			"key1:1,key2:2",
			reflect.StructField{Type: reflect.TypeOf(map[string]int{})},
			false,
		},
		{
			"Unsupported type",
			reflect.ValueOf(42),
			"42",
			reflect.StructField{Type: reflect.TypeOf(42)},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := handleSpecialTypes(tc.v, tc.val, tc.sf)
			if err != nil && !tc.err {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.err {
				t.Errorf("Expected error, got nil")
			}
		})
	}
}

func TestParseSliceOfStructs(t *testing.T) {
	tests := []struct {
		name string
		v    reflect.Value
		opts Options
		err  bool
	}{
		{
			"Empty slice, no env",
			reflect.ValueOf(&[]struct{}{}).Elem(),
			Options{},
			false, // Could not find any matching fields
		},
		{
			"Slice of structs, with env",
			reflect.ValueOf(&[]struct{ Foo string }{}).Elem(),
			Options{Env: map[string]string{"PREFIX_0_FOO": "foo_value"}, Prefix: "PREFIX"},
			false,
		},
		{
			"Slice of complex structs, with invalid env",
			reflect.ValueOf(&[]string{}).Elem(), // Not a struct
			Options{Env: map[string]string{"PREFIX_0_FOO": "invalid_int", "PREFIX_0_BAR": "bar_value"}, Prefix: "PREFIX"},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := parseSliceOfStructs(tc.v, tc.opts)
			if err != nil && !tc.err {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.err {
				t.Errorf("Expected error, got nil")
			}
		})
	}
}

func TestInitialiseSlice(t *testing.T) {
	tests := []struct {
		name     string
		v        reflect.Value
		maxIndex int
		expected int
	}{
		{
			"Slice of integers",
			reflect.ValueOf(&[]int{}).Elem(),
			5,
			6,
		},
		{
			"Slice of structs",
			reflect.ValueOf(&[]struct{ Foo string }{}).Elem(),
			3,
			4,
		},
		{
			"Slice of pointers",
			reflect.ValueOf(&[]*int{}).Elem(),
			2,
			3,
		},
		{
			"Pointer to slice of integers with initialised length greater than new length",
			reflect.ValueOf(&[]int{1, 2, 3, 4, 5}),
			2,
			3,
		},
		{
			"Pointer to slice of integers with initialised length less than new length",
			reflect.ValueOf(&[]int{1, 2, 3, 4, 5}),
			4,
			5,
		},
		{
			"Pointer to slice of integers with initialised length greater than maxIndex",
			reflect.ValueOf(&[]int{1, 2, 3, 4, 5}).Elem(),
			2,
			5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, capacity := initialiseSlice(tc.v, tc.maxIndex)
			if result.Len() != tc.expected {
				t.Errorf("Expected length %d, got %d", tc.expected, result.Len())
			}
			if capacity != tc.expected {
				t.Errorf("Expected capacity %d, got %d", tc.expected, capacity)
			}
		})
	}
}

func TestPopulateSlice(t *testing.T) {
	tests := []struct {
		name          string
		v             reflect.Value
		prefixedEnv   map[int]bool
		capacity      int
		opts          Options
		expectedError bool
	}{
		{
			"Slice of integers",
			reflect.ValueOf(&[]int{
				1, 2, 3, 4, 5,
			}).Elem(),
			map[int]bool{0: true},
			5,
			Options{Prefix: "PREFIX"},
			true, // Expected a struct got int
		},
		{
			"Slice of structs",
			reflect.ValueOf(&[]struct{ Foo string }{
				{Foo: "foo"},
			}).Elem(),
			map[int]bool{0: true},
			1,
			Options{Env: map[string]string{"PREFIX_0_FOO": "foo_value"}, Prefix: "PREFIX"},
			false,
		},
		{
			"No prefixed env",
			reflect.ValueOf(&[]struct{ Foo string }{
				{Foo: "foo"},
			}).Elem(),
			map[int]bool{
				0: false,
			},
			1,
			Options{Env: map[string]string{"PREFIX_0_FOO": "foo_value"}, Prefix: "PREFIX"},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			slice := reflect.MakeSlice(tc.v.Type(), tc.capacity, tc.capacity)
			err := populateSlice(slice, tc.v, tc.prefixedEnv, tc.capacity, tc.opts)
			if err != nil && !tc.expectedError {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.expectedError {
				t.Errorf("Expected error, got nil")
			}
		})
	}
}

type testUnmarshaler struct {
	Value string
}

func (t *testUnmarshaler) UnmarshalText(text []byte) error {
	t.Value = string(text)
	if t.Value == "" {
		return errors.New("empty value")
	}
	return nil
}

func TestParseTextUnmarshalers(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.Value
		data     []string
		expected []string
		err      bool
	}{
		{
			name:     "Valid unmarshaler",
			field:    reflect.ValueOf(&[]*testUnmarshaler{}).Elem(),
			data:     []string{"value1", "value2"},
			expected: []string{"value1", "value2"},
			err:      false,
		},
		{
			name:  "Invalid unmarshaler type",
			field: reflect.ValueOf(&[]*int{}).Elem(),
			data:  []string{"value1", "value2"},
			err:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := parseTextUnmarshalers(tc.field, tc.data)
			if (err != nil) != tc.err {
				t.Errorf("Expected error: %v, got: %v", tc.err, err)
			}
			if !tc.err {
				for i, v := range tc.field.Interface().([]*testUnmarshaler) {
					if v.Value != tc.expected[i] {
						t.Errorf("Expected value: %v, got: %v", tc.expected[i], v.Value)
					}
				}
			}
		})
	}
}

func TestParseElement(t *testing.T) {
	tests := []struct {
		name     string
		target   reflect.Value
		elemType reflect.Type
		value    string
		expected string
		err      bool
	}{
		{
			name:     "Valid unmarshaler",
			target:   reflect.ValueOf(&testUnmarshaler{}).Elem(),
			elemType: reflect.TypeOf(&testUnmarshaler{}).Elem(),
			value:    "test_value",
			expected: "test_value",
			err:      false,
		},
		{
			name:     "Invalid unmarshaler type",
			target:   reflect.ValueOf(&struct{}{}).Elem(),
			elemType: reflect.TypeOf(struct{}{}),
			value:    "test_value",
			err:      true,
		},
		{
			name:     "UnmarshalText error",
			target:   reflect.ValueOf(&testUnmarshaler{}).Elem(),
			elemType: reflect.TypeOf(&testUnmarshaler{}).Elem(),
			value:    "",
			expected: "",
			err:      true,
		},
		{
			name:     "Pointer target",
			target:   reflect.ValueOf(&testUnmarshaler{}).Elem(),
			elemType: reflect.TypeOf(&testUnmarshaler{}).Elem(),
			value:    "pointer_value",
			expected: "pointer_value",
			err:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := parseElement(tc.target, tc.elemType, tc.value)
			if (err != nil) != tc.err {
				t.Errorf("Expected error: %v, got: %v", tc.err, err)
			}
			if !tc.err {
				if unmarshaler, ok := tc.target.Interface().(encoding.TextUnmarshaler); ok {
					if unmarshaler.(*testUnmarshaler).Value != tc.expected {
						t.Errorf("Expected value: %v, got: %v", tc.expected, unmarshaler.(*testUnmarshaler).Value)
					}
				}
			}
		})
	}
}

func TestHandleSlice(t *testing.T) {
	tests := []struct {
		name          string
		v             reflect.Value
		val           string
		sf            reflect.StructField
		expected      interface{}
		expectedError bool
	}{
		{
			name: "Slice of integers",
			v:    reflect.ValueOf(&[]int{}).Elem(),
			val:  "1,2,3",
			sf: reflect.TypeOf(struct {
				Field []int `env:"FIELD"`
			}{}).Field(0),
			expected: []int{1, 2, 3},
		},
		{
			name: "Slice of strings",
			v:    reflect.ValueOf(&[]string{}).Elem(),
			val:  "foo,bar,baz",
			sf: reflect.TypeOf(struct {
				Field []string `env:"FIELD"`
			}{}).Field(0),
			expected: []string{"foo", "bar", "baz"},
		},
		{
			name: "Slice of bools",
			v:    reflect.ValueOf(&[]bool{}).Elem(),
			val:  "true,false,true",
			sf: reflect.TypeOf(struct {
				Field []bool `env:"FIELD"`
			}{}).Field(0),
			expected: []bool{true, false, true},
		},
		{
			name: "Invalid integer slice",
			v:    reflect.ValueOf(&[]int{}).Elem(),
			val:  "1,2,invalid",
			sf: reflect.TypeOf(struct {
				Field []int `env:"FIELD"`
			}{}).Field(0),
			expectedError: true,
		},
		{
			name: "Unsupported slice type",
			v:    reflect.ValueOf(&[]struct{}{}).Elem(),
			val:  "value",
			sf: reflect.TypeOf(struct {
				Field []struct{} `env:"FIELD"`
			}{}).Field(0),
			expectedError: true,
		},
		{
			name: "Slice of unmarshalers",
			v:    reflect.ValueOf(&[]*testUnmarshaler{}).Elem(),
			val:  "value1,value2",
			sf: reflect.TypeOf(struct {
				Field []*testUnmarshaler `env:"FIELD"`
			}{}).Field(0),
			expectedError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := handleSlice(tc.v, tc.val, tc.sf)
			if (err != nil) != tc.expectedError {
				t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
			}
			if !tc.expectedError && tc.expected != nil && !reflect.DeepEqual(tc.v.Interface(), tc.expected) {
				t.Errorf("Expected value: %v, got: %v", tc.expected, tc.v.Interface())
			}
		})
	}
}

func TestGetSeparator(t *testing.T) {
	tests := []struct {
		name     string
		tag      reflect.StructTag
		expected string
	}{
		{
			name:     "Default separator",
			tag:      `env:"FIELD"`,
			expected: ",",
		},
		{
			name:     "Custom separator",
			tag:      `env:"FIELD" envSeparator:";"`,
			expected: ";",
		},
		{
			name:     "Empty separator",
			tag:      `env:"FIELD" envSeparator:""`,
			expected: ",",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sf := reflect.StructField{Tag: tc.tag}
			separator := getSeparator(sf)
			if separator != tc.expected {
				t.Errorf("Expected separator: %v, got: %v", tc.expected, separator)
			}
		})
	}
}

func TestGetParserFunc(t *testing.T) {
	tests := []struct {
		name        string
		elemType    reflect.Type
		expectedErr error
	}{
		{
			name:     "Bool type",
			elemType: reflect.TypeOf(true),
		},
		{
			name:     "Int type",
			elemType: reflect.TypeOf(int(0)),
		},
		{
			name:     "Float64 type",
			elemType: reflect.TypeOf(float64(0)),
		},
		{
			name:     "String type",
			elemType: reflect.TypeOf(""),
		},
		{ // Uses custom unmarshaler from typeParsers
			name:     "Duration type",
			elemType: reflect.TypeOf(time.Nanosecond),
		},
		{
			name:        "Unsupported type",
			elemType:    reflect.TypeOf(struct{}{}),
			expectedErr: errors.New("unsupported type"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := getParserFunc(tc.elemType)
			if (err != nil) != (tc.expectedErr != nil) {
				t.Errorf("Expected error: %v, got: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestParseSliceElements(t *testing.T) {
	tests := []struct {
		name          string
		parts         []string
		elemType      reflect.Type
		parserFunc    func(string) (interface{}, error)
		elemKind      reflect.Type
		expected      interface{}
		expectedError bool
	}{
		{
			name:     "Valid integer slice",
			parts:    []string{"1", "2", "3"},
			elemType: reflect.TypeOf(int(0)),
			parserFunc: func(v string) (interface{}, error) {
				return strconv.Atoi(v)
			},
			elemKind: reflect.TypeOf(int(0)),
			expected: []int{1, 2, 3},
		},
		{
			name:     "Valid string slice",
			parts:    []string{"foo", "bar", "baz"},
			elemType: reflect.TypeOf(""),
			parserFunc: func(v string) (interface{}, error) {
				return v, nil
			},
			elemKind: reflect.TypeOf(""),
			expected: []string{"foo", "bar", "baz"},
		},
		{
			name:     "Invalid integer slice",
			parts:    []string{"1", "invalid", "3"},
			elemType: reflect.TypeOf(int(0)),
			parserFunc: func(v string) (interface{}, error) {
				return strconv.Atoi(v)
			},
			elemKind:      reflect.TypeOf(int(0)),
			expectedError: true,
		},
		{
			name:  "Pointer to int",
			parts: []string{"42"},
			parserFunc: func(v string) (interface{}, error) {
				i, err := strconv.Atoi(v)
				if err != nil {
					return nil, err
				}
				return i, nil
			},
			elemType: reflect.TypeOf(0),
			elemKind: reflect.TypeOf((*int)(nil)),
			expected: []*int{func() *int { i := 42; return &i }()},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseSliceElements(tc.parts, tc.elemType, tc.parserFunc, tc.elemKind)
			if (err != nil) != tc.expectedError {
				t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
			}
			if !tc.expectedError && !reflect.DeepEqual(result.Interface(), tc.expected) {
				t.Errorf("Expected value: %v, got: %v", tc.expected, result.Interface())
			}
		})
	}
}

func TestHandleMap(t *testing.T) {
	tests := []struct {
		name          string
		v             reflect.Value
		val           string
		sf            reflect.StructField
		expected      interface{}
		expectedError bool
	}{
		{
			name: "Valid map of strings",
			v:    reflect.ValueOf(&map[string]string{}).Elem(),
			val:  "key1:value1,key2:value2",
			sf: reflect.TypeOf(struct {
				Field map[string]string `env:"FIELD"`
			}{}).Field(0),
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name: "Valid map of integers",
			v:    reflect.ValueOf(&map[string]int{}).Elem(),
			val:  "key1:1,key2:2",
			sf: reflect.TypeOf(struct {
				Field map[string]int `env:"FIELD"`
			}{}).Field(0),
			expected: map[string]int{"key1": 1, "key2": 2},
		},
		{
			name: "Invalid map format",
			v:    reflect.ValueOf(&map[string]string{}).Elem(),
			val:  "key1:value1,key2",
			sf: reflect.TypeOf(struct {
				Field map[string]string `env:"FIELD"`
			}{}).Field(0),
			expectedError: true,
		},
		{
			name: "Unsupported type",
			v:    reflect.ValueOf(&map[string]struct{}{}).Elem(),
			val:  "key1:value1,key2:value2",
			sf: reflect.TypeOf(struct {
				Field map[string]struct{} `env:"FIELD"`
			}{}).Field(0),
			expectedError: true,
		},
		{
			name: "Unsupported key type",
			v:    reflect.ValueOf(&map[bool]string{}).Elem(),
			val:  "val1:value1,val2:value2",
			sf: reflect.TypeOf(struct {
				Field map[bool]string `env:"FIELD"`
			}{}).Field(0),
			expectedError: true,
		},
		{
			name: "Unsupported val type",
			v:    reflect.ValueOf(&map[string]bool{}).Elem(),
			val:  "key1:value1,key2:value2",
			sf: reflect.TypeOf(struct {
				Field map[string]bool `env:"FIELD"`
			}{}).Field(0),
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := handleMap(tc.v, tc.val, tc.sf)
			if (err != nil) != tc.expectedError {
				t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
			}
			if !tc.expectedError && !reflect.DeepEqual(tc.v.Interface(), tc.expected) {
				t.Errorf("Expected value: %v, got: %v", tc.expected, tc.v.Interface())
			}
		})
	}
}

func TestGetKeyAndElemParsers(t *testing.T) {
	tests := []struct {
		name          string
		mapType       reflect.Type
		expectedKey   interface{}
		expectedElem  interface{}
		expectedError bool
	}{
		{
			name:         "Valid map with string keys and int values",
			mapType:      reflect.TypeOf(map[string]int{}),
			expectedKey:  "key",
			expectedElem: 1,
		},
		{
			name:         "Valid map with int keys and string values",
			mapType:      reflect.TypeOf(map[int]string{}),
			expectedKey:  1,
			expectedElem: "value",
		},
		{
			name:         "Valid map with bool keys and float64 values",
			mapType:      reflect.TypeOf(map[bool]float64{}),
			expectedKey:  true,
			expectedElem: 1.0,
		},
		{
			name:          "Unsupported key type",
			mapType:       reflect.TypeOf(map[struct{}]string{}),
			expectedError: true,
		},
		{
			name:          "Unsupported element type",
			mapType:       reflect.TypeOf(map[string]struct{}{}),
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyParser, elemParser, err := getKeyAndElemParsers(tc.mapType)
			if (err != nil) != tc.expectedError {
				t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
			}
			if !tc.expectedError {
				key, keyErr := keyParser(fmt.Sprintf("%v", tc.expectedKey))
				if keyErr != nil || key != tc.expectedKey {
					t.Errorf("Expected key: %v, got: %v, error: %v", tc.expectedKey, key, keyErr)
				}
				elem, elemErr := elemParser(fmt.Sprintf("%v", tc.expectedElem))
				if elemErr != nil || elem != tc.expectedElem {
					t.Errorf("Expected element: %v, got: %v, error: %v", tc.expectedElem, elem, elemErr)
				}
			}
		})
	}
}

func BenchmarkParseSliceOfStructs(b *testing.B) {
	type TestStruct struct {
		Foo string `env:"FOO"`
	}
	data := map[string]string{
		"PREFIX_0_FOO": "foo_value",
		"PREFIX_1_FOO": "bar_value",
	}
	opts := Options{Env: data, Prefix: "PREFIX"}

	for i := 0; i < b.N; i++ {
		var ts []TestStruct
		ref := reflect.ValueOf(&ts).Elem()
		_ = parseSliceOfStructs(ref, opts)
	}
}
