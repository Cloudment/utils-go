package env

import (
	"reflect"
	"testing"
)

func TestToMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string]string
	}{
		{
			name:     "Simple key-value pair",
			input:    []string{"KEY=value"},
			expected: map[string]string{"KEY": "value"},
		},
		{
			name:     "Multiple key-value pairs",
			input:    []string{"KEY1=value1", "KEY2=value2", "KEY3=value3"},
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2", "KEY3": "value3"},
		},
		{
			name:     "Empty input",
			input:    []string{},
			expected: map[string]string{},
		},
		{
			name:     "No equals sign",
			input:    []string{"INVALID"},
			expected: map[string]string{},
		},
		{
			name:     "Mixed valid and invalid entries",
			input:    []string{"KEY=value", "INVALID", "ANOTHER_KEY=another_value"},
			expected: map[string]string{"KEY": "value", "ANOTHER_KEY": "another_value"},
		},
		{
			name:     "Key with empty value",
			input:    []string{"KEY="},
			expected: map[string]string{"KEY": ""},
		},
		{
			name:     "Empty key with value",
			input:    []string{"=value"},
			expected: map[string]string{"": "value"},
		},
		{
			name:     "Key with multiple equals signs",
			input:    []string{"KEY=val=ue"},
			expected: map[string]string{"KEY": "val=ue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toMap(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("toMap(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func BenchmarkToMap(b *testing.B) {
	envVars := []string{"KEY1=value1", "KEY2=value2", "KEY3=value3"}
	for i := 0; i < b.N; i++ {
		toMap(envVars)
	}
}
