package env

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestParseFieldTags(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.StructField
		opts     Options
		expected FieldTags
	}{
		{
			name: "Ignored field",
			field: reflect.StructField{
				Name: "IgnoredField",
				Tag:  `env:"-"`,
			},
			opts: Options{},
			expected: FieldTags{
				OwnKey:  "-",
				Ignored: true,
			},
		},
		{
			name: "Required field",
			field: reflect.StructField{
				Name: "RequiredField",
				Tag:  `env:"REQUIRED_FIELD,required"`,
			},
			opts: Options{},
			expected: FieldTags{
				OwnKey:   "REQUIRED_FIELD",
				Key:      "REQUIRED_FIELD",
				Required: true,
			},
		},
		{
			name: "Field with default value",
			field: reflect.StructField{
				Name: "DefaultField",
				Tag:  `env:"DEFAULT_FIELD" envDefault:"default_value"`,
			},
			opts: Options{},
			expected: FieldTags{
				OwnKey:  "DEFAULT_FIELD",
				Key:     "DEFAULT_FIELD",
				Default: "default_value",
			},
		},
		{
			name: "Field with multiple tags",
			field: reflect.StructField{
				Name: "ComplexField",
				Tag:  `env:"COMPLEX_FIELD,required,expand,init,unset"`,
			},
			opts: Options{},
			expected: FieldTags{
				OwnKey:   "COMPLEX_FIELD",
				Key:      "COMPLEX_FIELD",
				Required: true,
				Expand:   true,
				Init:     true,
				Unset:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := parseFieldTags(tt.field, tt.opts)
			if !reflect.DeepEqual(tags, tt.expected) {
				t.Errorf("parseFieldTags() = %v; want %v", tags, tt.expected)
			}
		})
	}
}

func TestApplyParser(t *testing.T) {
	tests := []struct {
		name     string
		v        reflect.Value
		sfType   reflect.Type
		val      string
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "Valid string parser",
			v:        reflect.New(reflect.TypeOf("")),
			sfType:   reflect.TypeOf(""),
			val:      "test",
			expected: "test",
			wantErr:  false,
		},
		{
			name:     "Valid int parser",
			v:        reflect.New(reflect.TypeOf(0)),
			sfType:   reflect.TypeOf(0),
			val:      "123",
			expected: 123,
			wantErr:  false,
		},
		{
			name:     "Invalid int parser",
			v:        reflect.New(reflect.TypeOf(0)),
			sfType:   reflect.TypeOf(0),
			val:      "invalid",
			expected: nil,
			wantErr:  true,
		},
		{ // An unsupported type should return false but should never error
			name:     "Unsupported type",
			v:        reflect.New(reflect.TypeOf([]string{})),
			sfType:   reflect.TypeOf([]string{}),
			val:      "unsupported",
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "Valid time.Duration parser",
			v:        reflect.New(reflect.TypeOf(time.Nanosecond)),
			sfType:   reflect.TypeOf(time.Nanosecond),
			val:      "1s",
			expected: time.Second,
			wantErr:  false,
		},
		{
			name:     "Invalid time.Duration parser",
			v:        reflect.New(reflect.TypeOf(time.Nanosecond)),
			sfType:   reflect.TypeOf(time.Nanosecond),
			val:      "invalid",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyParser(tt.v.Elem(), tt.sfType, tt.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyParser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != (tt.expected != nil) {
				t.Errorf("applyParser() got = %v, expected %v", got, tt.expected != nil)
			}
			if tt.expected != nil && !reflect.DeepEqual(tt.v.Elem().Interface(), tt.expected) {
				t.Errorf("applyParser() = %v, expected %v", tt.v.Elem().Interface(), tt.expected)
			}
		})
	}
}

func TestResolveValue(t *testing.T) {
	tests := []struct {
		name     string
		tags     FieldTags
		opts     Options
		expected string
		wantErr  bool
	}{
		{
			name: "Value from environment variable",
			tags: FieldTags{
				Key: "TEST_ENV_VAR",
			},
			opts: Options{
				Env: map[string]string{
					"TEST_ENV_VAR": "value",
				},
			},
			expected: "value",
			wantErr:  false,
		},
		{
			name: "Value from default",
			tags: FieldTags{
				Key:     "TEST_ENV_VAR",
				Default: "default_value",
			},
			opts: Options{
				Env: map[string]string{},
			},
			expected: "default_value",
			wantErr:  false,
		},
		{
			name: "Required value not set",
			tags: FieldTags{
				Key:      "TEST_ENV_VAR",
				Required: true,
			},
			opts: Options{
				Env: map[string]string{},
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "Expand environment variable",
			tags: FieldTags{
				Key:     "TEST_ENV_VAR",
				Default: "default_${EXPAND_VAR}",
				Expand:  true,
			},
			opts: Options{
				Env: map[string]string{
					"EXPAND_VAR": "expanded",
				},
			},
			expected: "default_expanded",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := resolveValue(tt.tags, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if val != tt.expected {
				t.Errorf("resolveValue() = %v, expected %v", val, tt.expected)
			}
		})
	}
}

func TestSetField(t *testing.T) {
	tests := []struct {
		name     string
		v        reflect.Value
		sf       reflect.StructField
		tags     FieldTags
		opts     Options
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "Set string field",
			v:        reflect.New(reflect.TypeOf("")).Elem(),
			sf:       reflect.StructField{Name: "StringField", Type: reflect.TypeOf(""), Tag: `env:"STRING_FIELD"`},
			tags:     FieldTags{Key: "STRING_FIELD"},
			opts:     Options{Env: map[string]string{"STRING_FIELD": "value"}},
			expected: "value",
			wantErr:  false,
		},
		{
			name:     "Set int field",
			v:        reflect.New(reflect.TypeOf(0)).Elem(),
			sf:       reflect.StructField{Name: "IntField", Type: reflect.TypeOf(0), Tag: `env:"INT_FIELD"`},
			tags:     FieldTags{Key: "INT_FIELD"},
			opts:     Options{Env: map[string]string{"INT_FIELD": "123"}},
			expected: 123,
			wantErr:  false,
		},
		{
			name:     "Unset environment variable",
			v:        reflect.New(reflect.TypeOf("")).Elem(),
			sf:       reflect.StructField{Name: "UnsetField", Type: reflect.TypeOf(""), Tag: `env:"UNSET_FIELD" env:",unset"`},
			tags:     FieldTags{Key: "UNSET_FIELD", Unset: true},
			opts:     Options{Env: map[string]string{"UNSET_FIELD": "value"}},
			expected: "value",
			wantErr:  false,
		},
		{
			name:     "Invalid int field",
			v:        reflect.New(reflect.TypeOf(0)).Elem(),
			sf:       reflect.StructField{Name: "InvalidIntField", Type: reflect.TypeOf(0), Tag: `env:"INVALID_INT_FIELD"`},
			tags:     FieldTags{Key: "INVALID_INT_FIELD"},
			opts:     Options{Env: map[string]string{"INVALID_INT_FIELD": "invalid"}},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Handle special types",
			v:        reflect.New(reflect.TypeOf(time.Nanosecond)).Elem(),
			sf:       reflect.StructField{Name: "DurationField", Type: reflect.TypeOf(time.Nanosecond), Tag: `env:"DURATION_FIELD"`},
			tags:     FieldTags{Key: "DURATION_FIELD"},
			opts:     Options{Env: map[string]string{"DURATION_FIELD": "1s"}},
			expected: time.Second,
			wantErr:  false,
		},
		{
			name:     "Unsupported type",
			v:        reflect.New(reflect.TypeOf([]string{})).Elem(),
			sf:       reflect.StructField{Name: "UnsupportedField", Type: reflect.TypeOf([]string{}), Tag: `env:"UNSUPPORTED_FIELD"`},
			tags:     FieldTags{Key: "UNSUPPORTED_FIELD"},
			opts:     Options{Env: map[string]string{"UNSUPPORTED_FIELD": "unsupported"}},
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "Invalid time.Duration field",
			v:        reflect.New(reflect.TypeOf(time.Nanosecond)).Elem(),
			sf:       reflect.StructField{Name: "InvalidDurationField", Type: reflect.TypeOf(time.Nanosecond), Tag: `env:"INVALID_DURATION_FIELD"`},
			tags:     FieldTags{Key: "INVALID_DURATION_FIELD"},
			opts:     Options{Env: map[string]string{"INVALID_DURATION_FIELD": "invalid"}},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Handle slice",
			v:        reflect.New(reflect.TypeOf([]string{})).Elem(),
			sf:       reflect.StructField{Name: "SliceField", Type: reflect.TypeOf([]string{}), Tag: `env:"SLICE_FIELD"`},
			tags:     FieldTags{Key: "SLICE_FIELD"},
			opts:     Options{Env: map[string]string{"SLICE_FIELD": "item1,item2,item3"}},
			expected: []string{"item1", "item2", "item3"},
			wantErr:  false,
		},
		{
			name:     "Handle map",
			v:        reflect.New(reflect.TypeOf(map[string]string{})).Elem(),
			sf:       reflect.StructField{Name: "MapField", Type: reflect.TypeOf(map[string]string{}), Tag: `env:"MAP_FIELD"`},
			tags:     FieldTags{Key: "MAP_FIELD"},
			opts:     Options{Env: map[string]string{"MAP_FIELD": "key1:value1,key2:value2"}},
			expected: map[string]string{"key1": "value1", "key2": "value2"},
			wantErr:  false,
		},
		{
			name:     "Use default value",
			v:        reflect.New(reflect.TypeOf("")).Elem(),
			sf:       reflect.StructField{Name: "DefaultField", Type: reflect.TypeOf(""), Tag: `env:"DEFAULT_FIELD" envDefault:"default_value"`},
			tags:     FieldTags{Key: "DEFAULT_FIELD", Default: "default_value"},
			opts:     Options{Env: map[string]string{}},
			expected: "default_value",
			wantErr:  false,
		},
		{
			name:     "Required value not set",
			v:        reflect.New(reflect.TypeOf("")).Elem(),
			sf:       reflect.StructField{Name: "RequiredField", Type: reflect.TypeOf(""), Tag: `env:"REQUIRED_FIELD,required"`},
			tags:     FieldTags{Key: "REQUIRED_FIELD", Required: true},
			opts:     Options{Env: map[string]string{}},
			expected: "",
			wantErr:  true,
		},
		{
			"empty value",
			reflect.ValueOf(""),
			reflect.StructField{Name: "EmptyField", Type: reflect.TypeOf(""), Tag: `env:"EMPTY_FIELD"`},
			FieldTags{Key: "EMPTY_FIELD"},
			Options{Env: map[string]string{"EMPTY_FIELD": ""}},
			"",
			false,
		},
		{
			name: "Set TextUnmarshaler field",
			v:    reflect.New(reflect.TypeOf(&time.Time{})).Elem(),
			sf: reflect.StructField{
				Name: "TimeField",
				Type: reflect.TypeOf(&time.Time{}),
				Tag:  `env:"TIME_FIELD"`,
			},
			tags: FieldTags{Key: "TIME_FIELD"},
			opts: Options{Env: map[string]string{"TIME_FIELD": "2021-01-01T00:00:00Z"}},
			expected: func() *time.Time {
				x := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
				return &x
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setField(tt.v, tt.sf, tt.tags, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("setField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.expected != nil && !reflect.DeepEqual(tt.v.Interface(), tt.expected) {
				fmt.Println("Type of expected:", reflect.TypeOf(tt.expected))
				fmt.Println("Type of v.Interface():", reflect.TypeOf(tt.v.Interface()))
				t.Errorf("setField() = %v, expected %v", tt.v.Interface(), tt.expected)
			}
		})
	}
}

func TestHandleStructOrSlice(t *testing.T) {
	tests := []struct {
		name    string
		v       reflect.Value
		sf      reflect.StructField
		opts    Options
		tags    FieldTags
		wantErr bool
	}{
		{
			name:    "Pointer to struct",
			v:       reflect.New(reflect.TypeOf(struct{ Field string }{})),
			sf:      reflect.StructField{Name: "StructField", Type: reflect.TypeOf(struct{ Field string }{})},
			opts:    Options{},
			tags:    FieldTags{},
			wantErr: false,
		},
		{ // The difference between this and the others is that .Elem() is called on the value
			name:    "Addressable struct",
			v:       reflect.ValueOf(&struct{ Field string }{}).Elem(),
			sf:      reflect.StructField{Name: "StructField", Type: reflect.TypeOf(struct{ Field string }{})},
			opts:    Options{},
			tags:    FieldTags{},
			wantErr: false,
		},
		{
			name:    "Non-addressable struct",
			v:       reflect.ValueOf(struct{ Field string }{}),
			sf:      reflect.StructField{Name: "StructField", Type: reflect.TypeOf(struct{ Field string }{})},
			opts:    Options{},
			tags:    FieldTags{},
			wantErr: true,
		},
		{
			name:    "Slice of structs",
			v:       reflect.ValueOf([]struct{ Field string }{}),
			sf:      reflect.StructField{Name: "SliceField", Type: reflect.TypeOf([]struct{ Field string }{})},
			opts:    Options{},
			tags:    FieldTags{},
			wantErr: false,
		},
		{
			name:    "Invalid pointer to map with init tag",
			v:       reflect.New(reflect.TypeOf((*map[string]string)(nil))).Elem(),
			sf:      reflect.StructField{Name: "PointerField", Type: reflect.TypeOf((*map[string]string)(nil)).Elem()},
			opts:    Options{},
			tags:    FieldTags{Init: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handleStructOrSlice(tt.v, tt.sf, tt.opts, tt.tags); (err != nil) != tt.wantErr {
				t.Errorf("handleStructOrSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandlePointerStruct(t *testing.T) {
	tests := []struct {
		name    string
		v       reflect.Value
		sf      reflect.StructField
		opts    Options
		wantErr bool
	}{
		{
			name:    "Pointer to struct",
			v:       reflect.New(reflect.TypeOf(struct{ Field string }{})),
			sf:      reflect.StructField{Name: "StructField", Type: reflect.TypeOf(struct{ Field string }{})},
			opts:    Options{},
			wantErr: false,
		},
		{
			name:    "Addressable struct",
			v:       reflect.ValueOf(&struct{ Field string }{}).Elem(),
			sf:      reflect.StructField{Name: "StructField", Type: reflect.TypeOf(struct{ Field string }{})},
			opts:    Options{},
			wantErr: false,
		},
		{
			name:    "Non-addressable struct",
			v:       reflect.ValueOf(struct{ Field string }{}),
			sf:      reflect.StructField{Name: "StructField", Type: reflect.TypeOf(struct{ Field string }{})},
			opts:    Options{},
			wantErr: false,
		},
		{
			name:    "Nil pointer to struct",
			v:       reflect.New(reflect.TypeOf((*struct{ Field string })(nil))).Elem(),
			sf:      reflect.StructField{Name: "StructField", Type: reflect.TypeOf(struct{ Field string }{})},
			opts:    Options{},
			wantErr: false,
		},
		{
			name:    "Nil pointer to struct",
			v:       reflect.ValueOf((*struct{ Field string })(nil)).Elem(),
			sf:      reflect.StructField{Name: "StructField", Type: reflect.TypeOf(struct{ Field string }{})},
			opts:    Options{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handlePointerStruct(tt.v, tt.sf, tt.opts); (err != nil) != tt.wantErr {
				t.Errorf("handlePointerStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseField(t *testing.T) {
	tests := []struct {
		name    string
		v       reflect.Value
		sf      reflect.StructField
		opts    Options
		wantErr bool
	}{
		{
			name:    "Handle pointer struct",
			v:       reflect.New(reflect.TypeOf(&struct{ Field string }{})).Elem(),
			sf:      reflect.StructField{Name: "PointerStructField", Type: reflect.TypeOf(&struct{ Field string }{}), Tag: `env:"POINTER_STRUCT_FIELD"`},
			opts:    Options{},
			wantErr: false,
		},
		{
			name:    "Set field value",
			v:       reflect.New(reflect.TypeOf("")).Elem(),
			sf:      reflect.StructField{Name: "SetField", Type: reflect.TypeOf(""), Tag: `env:"SET_FIELD"`},
			opts:    Options{Env: map[string]string{"SET_FIELD": "value"}},
			wantErr: false,
		},
		{
			name:    "Required field missing value",
			v:       reflect.New(reflect.TypeOf("")).Elem(),
			sf:      reflect.StructField{Name: "RequiredField", Type: reflect.TypeOf(""), Tag: `env:"REQUIRED_FIELD,required"`},
			opts:    Options{Env: map[string]string{}},
			wantErr: true,
		},
		{
			name:    "Handle pointer to struct",
			v:       reflect.New(reflect.TypeOf(&struct{ Field string }{})),
			sf:      reflect.StructField{Name: "PointerToStructField", Type: reflect.TypeOf(&struct{ Field string }{}), Tag: `env:"POINTER_TO_STRUCT_FIELD"`},
			opts:    Options{},
			wantErr: false,
		},
		{
			name:    "Slice of structs",
			v:       reflect.New(reflect.TypeOf([]struct{ Field string }{})).Elem(),
			sf:      reflect.StructField{Name: "SliceOfStructs", Type: reflect.TypeOf([]struct{ Field string }{}), Tag: `env:"SLICE_FIELD"`},
			opts:    Options{},
			wantErr: false,
		},
		{
			name:    "Initialise pointer",
			v:       reflect.New(reflect.TypeOf((*string)(nil)).Elem()).Elem(),
			sf:      reflect.StructField{Name: "PointerField", Type: reflect.TypeOf((*string)(nil)).Elem(), Tag: `env:"POINTER_FIELD"`},
			opts:    Options{},
			wantErr: false,
		},
		{
			name:    "Handle struct or slice",
			v:       reflect.New(reflect.TypeOf(struct{ Field string }{})).Elem(),
			sf:      reflect.StructField{Name: "StructField", Type: reflect.TypeOf(struct{ Field string }{}), Tag: `env:"STRUCT_FIELD"`},
			opts:    Options{},
			wantErr: false,
		},
		{
			name:    "Handle pointer struct with nil value",
			v:       reflect.New(reflect.TypeOf((*struct{ Field string })(nil))).Elem(),
			sf:      reflect.StructField{Name: "PointerStructField", Type: reflect.TypeOf(&struct{ Field string }{}), Tag: `env:"POINTER_STRUCT_FIELD"`},
			opts:    Options{},
			wantErr: false,
		},
		{
			name:    "Handle pointer struct with nil value",
			v:       reflect.New(reflect.TypeOf((*struct{ Field string })(nil)).Elem()).Elem(),
			sf:      reflect.StructField{Name: "PointerStructField", Type: reflect.TypeOf(&struct{ Field string }{}), Tag: `env:"POINTER_STRUCT_FIELD"`},
			opts:    Options{},
			wantErr: false,
		},
		{
			name: "Parse inner struct fails",
			v: reflect.ValueOf(&struct {
				Foo struct {
					Number int `env:"NUMBER"`
				}
			}{}).Elem(),
			sf: reflect.StructField{Name: "Foo", Type: reflect.TypeOf(struct {
				Number int `env:"NUMBER"`
			}{})},
			opts:    Options{Env: map[string]string{"NUMBER": "not-a-number"}},
			wantErr: true,
		},
		{
			name: "Parse nested field required not set",
			v: reflect.ValueOf(&struct {
				Foo *[]struct {
					Str string `env:"STR,required"`
					Num int    `env:"NUM"`
				} `envPrefix:"FOO"`
			}{}).Elem(),
			sf: reflect.StructField{Name: "Foo", Type: reflect.TypeOf(&[]struct {
				Str string `env:"STR,required"`
				Num int    `env:"NUM"`
			}{})},
			opts:    Options{Env: map[string]string{"FOO_0_NUM": "101"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseField(tt.v, tt.sf, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseField() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseStruct(t *testing.T) {
	tests := []struct {
		name    string
		ref     interface{}
		opts    Options
		wantErr bool
	}{
		{
			name: "Valid struct",
			ref: struct {
				Field string `env:"FIELD"`
			}{},
			opts:    Options{},
			wantErr: false,
		},
		{
			name: "Pointer to struct",
			ref: &struct {
				Field string `env:"FIELD"`
			}{},
			opts:    Options{},
			wantErr: false,
		},
		{
			name:    "Invalid type",
			ref:     "invalid",
			opts:    Options{},
			wantErr: true,
		},
		{
			name: "Struct with nested struct",
			ref: struct {
				Nested struct {
					Field string `env:"NESTED_FIELD"`
				}
			}{},
			opts:    Options{},
			wantErr: false,
		},
		{
			name: "Struct with nested struct pointer",
			ref: &struct {
				Field string `env:"FIELD,required"`
			}{},
			opts:    Options{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refValue := reflect.ValueOf(tt.ref)
			if err := parseStruct(refValue, tt.opts); (err != nil) != tt.wantErr {
				t.Errorf("parseStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseInterface(t *testing.T) {
	tests := []struct {
		name    string
		ref     interface{}
		opts    Options
		wantErr bool
	}{
		{
			name:    "ValidPointerToStruct",
			ref:     &struct{ Field string }{},
			opts:    Options{},
			wantErr: false,
		},
		{
			name:    "NonStructPointer",
			ref:     new(int),
			opts:    Options{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseInterface(tt.ref, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInterface() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseWithOpts(t *testing.T) {
	tests := []struct {
		name    string
		ref     interface{}
		opts    Options
		wantErr bool
	}{
		{
			name:    "ValidPointerToStruct",
			ref:     &struct{ Field string }{},
			opts:    Options{},
			wantErr: false,
		},
		{
			name:    "NilPointer",
			ref:     nil,
			opts:    Options{},
			wantErr: true,
		},
		{
			name:    "NonPointer",
			ref:     struct{ Field string }{},
			opts:    Options{},
			wantErr: true,
		},
		{
			name:    "NonStructPointer",
			ref:     new(int),
			opts:    Options{},
			wantErr: true,
		},
		{
			name:    "ValidStructPointer",
			ref:     &struct{ Field string }{},
			opts:    Options{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseWithOpts(tt.ref, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWithOpts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		ref     interface{}
		wantErr bool
	}{
		{
			name:    "ValidPointerToStruct",
			ref:     &struct{ Field string }{},
			wantErr: false,
		},
		{
			name:    "NilPointer",
			ref:     nil,
			wantErr: true,
		},
		{
			name:    "NonPointer",
			ref:     struct{ Field string }{},
			wantErr: true,
		},
		{
			name:    "NonStructPointer",
			ref:     new(int),
			wantErr: true,
		},
		{
			name:    "ValidStructPointer",
			ref:     &struct{ Field string }{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Parse(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
	type Nested struct {
		Foo string `env:"FOO"`
	}
	type TestStruct struct {
		Foo    string            `env:"FOO"`
		Bar    *string           `env:"BAR"`
		Baz    map[string]string `env:"BAZ"`
		Qux    []string          `env:"QUX"`
		Nested `envPrefix:"NESTED_"`
	}

	b.Setenv("FOO", "foo_value")
	b.Setenv("BAR", "bar_value")
	b.Setenv("BAZ", "key1:value1,key2:value2")
	b.Setenv("QUX", "item1,item2,item3")
	b.Setenv("NESTED_FOO", "nested_foo_value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ts TestStruct
		err := Parse(&ts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseWithOpts(b *testing.B) {
	type Nested struct {
		Foo string `env:"FOO"`
	}
	type TestStruct struct {
		Foo    string            `env:"FOO"`
		Bar    *string           `env:"BAR"`
		Baz    map[string]string `env:"BAZ"`
		Qux    []string          `env:"QUX"`
		Nested `envPrefix:"NESTED_"`
	}

	data := map[string]string{
		"FOO":        "foo_value",
		"BAR":        "bar_value",
		"BAZ":        "key1:value1,key2:value2",
		"QUX":        "item1,item2,item3",
		"NESTED_FOO": "nested_foo_value",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ts TestStruct
		err := ParseWithOpts(&ts, Options{Env: data})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseInterface(b *testing.B) {
	type TestStruct struct {
		Foo string `env:"FOO"`
		Bar string `env:"BAR"`
		Baz string `env:"BAZ"`
	}

	data := map[string]string{
		"FOO": "foo_value",
		"BAR": "bar_value",
		"BAZ": "baz_value",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ts TestStruct
		err := parseInterface(&ts, Options{
			Env: data,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseStruct(b *testing.B) {
	type TestStruct struct {
		Foo string `env:"FOO"`
		Bar string `env:"BAR"`
		Baz string `env:"BAZ"`
	}

	data := map[string]string{
		"FOO": "foo_value",
		"BAR": "bar_value",
		"BAZ": "baz_value",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ts TestStruct
		err := parseStruct(reflect.ValueOf(&ts).Elem(), Options{
			Env: data,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseField(b *testing.B) {
	type NestedStruct struct {
		NestedField1 string
		NestedField2 int
	}
	type TestStruct struct {
		Field1 string
		Field2 int
		Field3 *NestedStruct
	}

	v := reflect.ValueOf(&TestStruct{
		Field1: "test",
		Field2: 42,
		Field3: &NestedStruct{
			NestedField1: "nested",
			NestedField2: 99,
		},
	}).Elem()

	sf, _ := v.Type().FieldByName("Field3")
	opts := Options{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := parseField(v.FieldByName("Field3"), sf, opts); err != nil {
			b.Fatalf("parseField failed: %v", err)
		}
	}
}

func BenchmarkHandlePointerStruct(b *testing.B) {
	type TestStruct struct {
		Foo string `env:"FOO"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := handlePointerStruct(reflect.New(reflect.TypeOf(&TestStruct{})).Elem(), reflect.StructField{}, Options{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHandleStructOrSlice(b *testing.B) {
	v := reflect.ValueOf([]struct{ Field string }{})
	sf := reflect.StructField{Name: "SliceField", Type: reflect.TypeOf([]struct{ Field string }{})}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := handleStructOrSlice(v, sf, Options{}, FieldTags{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSetField(b *testing.B) {
	v := reflect.New(reflect.TypeOf("")).Elem()
	sf := reflect.StructField{Name: "Field", Type: reflect.TypeOf(""), Tag: `env:"FIELD"`}
	tags := FieldTags{Key: "FIELD"}
	opts := Options{Env: map[string]string{"FIELD": "value"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := setField(v, sf, tags, opts); err != nil {
			b.Fatalf("setField failed: %v", err)
		}
	}
}

func BenchmarkResolveValue(b *testing.B) {
	tags := FieldTags{Key: "FIELD"}
	opts := Options{Env: map[string]string{"FIELD": "value"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := resolveValue(tags, opts); err != nil {
			b.Fatalf("resolveValue failed: %v", err)
		}
	}
}

func BenchmarkApplyParser(b *testing.B) {
	v := reflect.New(reflect.TypeOf("")).Elem()
	sfType := reflect.TypeOf("")
	val := "test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := applyParser(v, sfType, val); err != nil {
			b.Fatalf("applyParser failed: %v", err)
		}
	}
}

func BenchmarkParseFieldTags(b *testing.B) {
	field := reflect.StructField{
		Name: "Foo",
		Tag:  `env:"FOO" envDefault:"default" envPrefix:"PREFIX_"`,
	}
	opts := Options{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseFieldTags(field, opts)
	}
}
