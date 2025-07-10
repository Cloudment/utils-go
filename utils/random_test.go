package utils

import (
	"crypto/rand"
	"errors"
	"testing"
	"time"
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

	num, err = GenerateRandomNumber(0, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	} else if num != 0 && num != 1 {
		t.Errorf("Expected number to be 0 or 1, got %d", num)
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

func TestGenerateRandomBytesGeneric(t *testing.T) {
	bytes, err := GenerateRandomBytesGeneric[uint8](10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(bytes) != 10 {
		t.Errorf("Expected 10 bytes, got %d", len(bytes))
	}

	bytes, err = GenerateRandomBytesGeneric[int8](10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(bytes) != 10 {
		t.Errorf("Expected 10 bytes, got %d", len(bytes))
	}

	bytes, err = GenerateRandomBytesGeneric[uint64](10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(bytes) != 10 {
		t.Errorf("Expected 10 bytes, got %d", len(bytes))
	}

	bytes, err = GenerateRandomBytesGeneric[int64](10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(bytes) != 10 {
		t.Errorf("Expected 10 bytes, got %d", len(bytes))
	}

	_, err = GenerateRandomBytesGeneric[uint8](0)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	_, err = GenerateRandomBytesGeneric[int8](0)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	_, err = generateRandomBytesGeneric[uint8](10, &errorReader{})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGenerateRandomDuration(t *testing.T) {
	duration, err := GenerateRandomDuration(1, 10, time.Second)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if duration < 1*time.Second || duration >= 10*time.Second {
		t.Errorf("Expected duration between 1 and 10 seconds, got %v", duration)
	}

	_, err = GenerateRandomDuration(10, 1, time.Second)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	duration, err = GenerateRandomDuration(1, 10, time.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	} else if duration < 1*time.Millisecond || duration >= 10*time.Millisecond {
		t.Errorf("Expected duration between 1 and 10 milliseconds, got %v", duration)
	}

	_, err = generateRandomDuration(1, 10, time.Second, &errorReader{})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func BenchmarkGenerateRandomDuration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateRandomDuration(1, 10, time.Second)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestGenerateOTP(t *testing.T) {
	// Large number of iterations to ensure tests try to fail it
	for i := 0; i < 100; i++ {
		otp, err := generateOTP(2, rand.Reader)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if otp < 10 || otp >= 100 { // Upper bound
			t.Errorf("Expected OTP between 10 and 100, got %d", otp)
		}
	}

	_, err := generateOTP(0, rand.Reader)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	_, err = generateOTP(6, &errorReader{})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	_, err = GenerateOTP(6)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func BenchmarkGenerateOTP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := generateOTP(6, rand.Reader)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
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

func BenchmarkGenerateRandomBytesWithGenericsUint16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateRandomBytesGeneric[uint16](100)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkGenerateRandomBytesWithGenericsUint8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateRandomBytesGeneric[uint8](100)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkGenerateRandomBytesWithGenericsInt32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateRandomBytesGeneric[int32](100)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}
