package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type Request struct {
	Field1     string  `query:"field1" form:"field1" json:"field1" required:"true"`
	Field2     string  `query:"field2" form:"field2" json:"field2"`
	Int        int     `query:"int" form:"int" json:"int"`
	Float      float64 `query:"float" form:"float" json:"float"`
	Bool       bool    `query:"bool" form:"bool" json:"bool"`
	unexported string  `query:"unexported" form:"unexported"`
}

// TODO: Add examples for BindRequest

func TestBindRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     *http.Request
		expected    Request
		expectError bool
	}{
		{
			name:    "Valid query parameters",
			request: httptest.NewRequest(http.MethodGet, "/test?field1=value1&field2=value2&int=42&float=42.5&bool=true", nil),
			expected: Request{
				Field1: "value1",
				Field2: "value2",
				Int:    42,
				Float:  42.5,
				Bool:   true,
			},
			expectError: false,
		},
		{
			name: "Valid form data",
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("field1=value1&field2=value2&int=42&float=42.5&bool=true"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			}(),
			expected: Request{
				Field1: "value1",
				Field2: "value2",
				Int:    42,
				Float:  42.5,
				Bool:   true,
			},
			expectError: false,
		},
		{
			name: "Valid JSON body",
			request: func() *http.Request {
				data := map[string]any{
					"field1": "value1",
					"field2": "value2",
					"int":    42,
					"float":  42.5,
					"bool":   true,
				}

				jsonData, _ := json.Marshal(data)

				req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),
			expected: Request{
				Field1: "value1",
				Field2: "value2",
				Int:    42,
				Float:  42.5,
				Bool:   true,
			},
			expectError: false,
		},
		{
			name:        "Missing required field (Field1)",
			request:     httptest.NewRequest(http.MethodGet, "/test?field2=value2", nil),
			expectError: true,
		},
		{
			name:    "Only required field present",
			request: httptest.NewRequest(http.MethodGet, "/test?field1=value1", nil),
			expected: Request{
				Field1: "value1",
				Field2: "",
			},
			expectError: false,
		},
		{
			name: "Empty JSON body",
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("{"))
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),
			expectError: true,
		},
		{
			name: "Empty POST form data",
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/test", nil)
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Body = nil
				return req
			}(),
			expectError: true,
		},
		{
			name:        "Unexported field",
			request:     httptest.NewRequest(http.MethodGet, "/test?field1=value1&unexported=value", nil),
			expected:    Request{},
			expectError: true,
		},
		{
			name: "Invalid POST form data",
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("field1=value1&field2=value2&int=42&float=42.5&bool=thisisnotabool"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqStruct Request
			err := BindRequest(tt.request, &reqStruct)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil && !tt.expectError {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if reqStruct != tt.expected {
				t.Errorf("expected %+v, got %+v", tt.expected, reqStruct)
				return
			}
		})
	}
}

func TestSetFieldValue(t *testing.T) {
	testCases := []struct {
		name          string
		fieldKind     reflect.Kind
		input         string
		expectedValue interface{}
		expectedError bool
	}{
		{"Set string field", reflect.String, "test", "test", false},
		{"Set int field", reflect.Int, "42", int64(42), false},
		{"Set uint field", reflect.Uint, "42", uint64(42), false},
		{"Set float field", reflect.Float64, "42.5", 42.5, false},
		{"Set bool field", reflect.Bool, "true", true, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			field := reflect.New(reflect.TypeOf(tc.expectedValue)).Elem()
			err := setFieldValue(field, tc.input)

			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			actualValue := field.Interface()
			if actualValue != tc.expectedValue {
				t.Errorf("Expected %v, got %v", tc.expectedValue, actualValue)
			}
		})
	}
}
