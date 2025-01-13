package utils

import (
	"errors"
	"testing"
)

type errorReader struct{}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mocked error reading")
}

func TestGenerateRandomNumber(t *testing.T) {
	num, err := GenerateRandomNumber(1, 10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if num < 1 || num >= 10 {
		t.Errorf("Expected number between 1 and 10, got %d", num)
	}

	_, err = GenerateRandomNumber(10, 1)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	_, err = generateRandomNumber(1, 10, &errorReader{})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	num, err = GenerateRandomNumber(30, 32)
	print(num)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if num < 30 || num >= 50 {
		t.Errorf("Expected number between 30 and 50, got %d", num)
	}

	_, err = GenerateRandomNumber(-10, 10)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGenerateRandomString(t *testing.T) {
	str, err := GenerateRandomString(10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(str) != 10 {
		t.Errorf("Expected string of length 10, got %d", len(str))
	}

	_, err = GenerateRandomString(0)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	_, err = generateRandomString(10, &errorReader{})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	bytes, err := GenerateRandomBytes(10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(bytes) != 10 {
		t.Errorf("Expected 10 bytes, got %d", len(bytes))
	}

	_, err = GenerateRandomBytes(0)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	_, err = generateRandomBytes(10, &errorReader{})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func BenchmarkGenerateRandomNumber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateRandomNumber(1, 100)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkGenerateRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateRandomString(100)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkGenerateRandomBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateRandomBytes(100)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}
