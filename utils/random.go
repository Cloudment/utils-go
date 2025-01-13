package utils

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
)

// GenerateRandomNumber generates a secure random integer
// between min (inclusive) and max (exclusive) using the default rand.Reader.
//
// Parameters:
//   - min: The minimum value (inclusive).
//   - max: The maximum value (exclusive).
//
// Returns: The generated random number or an error if the generation fails.
//
// Usage:
//
//	num, err := GenerateRandomNumber(1, 10)
//
// Example:
//
//	num, err := GenerateRandomNumber(1, 10)
//	fmt.Println(num) // Output: 5
func GenerateRandomNumber(min int, max int) (int, error) {
	return generateRandomNumber(min, max, rand.Reader)
}

// GenerateRandomString generates a secure random string
// of the given length using the default rand.Reader.
//
// Parameters:
//   - length: The length of the generated string.
//
// Returns: The generated random string or an error if the generation fails.
//
// Usage:
//
//	str, err := GenerateRandomString(10)
//
// Example:
//
//	str, err := GenerateRandomString(10)
//	fmt.Println(str) // Output: "kTvOz81Qdt"
func GenerateRandomString(length int) (string, error) {
	return generateRandomString(length, rand.Reader)
}

// GenerateRandomBytes generates secure random bytes using the default rand.Reader.
//
// Parameters:
//   - n: The number of bytes to generate.
//
// Returns: The generated random bytes or an error if the generation fails.
//
// Usage:
//
//	bytes, err := GenerateRandomBytes(10)
//
// Example:
//
//	bytes, err := GenerateRandomBytes(10)
//	fmt.Println(bytes)
func GenerateRandomBytes(n int) ([]byte, error) {
	return generateRandomBytes(n, rand.Reader)
}

// generateRandomNumber generates a secure random integer
// between min (inclusive) and max (exclusive) using the provided reader.
//
// The minimum value should be less than the maximum value but also greater than or equal to 0.
//
// This function is not intended for negative values; for negative values, negate the value returned.
//
// Parameters:
//   - min: The minimum value (inclusive).
//   - max: The maximum value (exclusive).
//   - reader: The io.Reader to use for generating random numbers.
//
// Returns: The generated random number or an error if the generation fails.
//
// Example:
//
//	num, err := generateRandomNumber(1, 10, rand.Reader)
//	fmt.Println(num) // Output: 5
func generateRandomNumber(min int, max int, reader io.Reader) (int, error) {
	if min >= max {
		return 0, newParseValueError("min should be less than max")
	} else if min < 0 {
		// rand.Int panics if the max value is less than 0, ensuring max is 1 or greater prevents this
		return 0, newParseValueError("min should be greater than or equal to 0")
	}

	// rand.int takes in a max value, generation occurs between 0 and the range size
	// if intended max is 50, min is 10, the range size is 40
	// If the random number generated is 0, when added back within the min value, it will be 10
	rangeSize := int64(max - min)
	n, err := rand.Int(reader, big.NewInt(rangeSize))
	if err != nil {
		return 0, fmt.Errorf("could not generate random number: %w", err)
	}

	return int(n.Int64()) + min, nil
}

// generateRandomString generates a secure random string
// of the given length using the provided reader.
//
// Parameters:
//   - length: The length of the generated string.
//   - reader: The io.Reader to use for generating random numbers.
//
// Returns: The generated random string or an error if the generation fails.
//
// Usage:
//
//	str, err := generateRandomString(10, rand.Reader)
//
// Example:
//
//	str, err := generateRandomString(10, rand.Reader)
//	fmt.Println(str) // Output: "kTvOz81Qdt"
func generateRandomString(length int, reader io.Reader) (string, error) {
	if length <= 0 {
		return "", newParseValueError("length should be greater than 0")
	}

	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	letterCount := len(letters)

	for i := range b {
		num, err := rand.Int(reader, big.NewInt(int64(letterCount)))
		if err != nil {
			return "", fmt.Errorf("could not generate random string: %w", err)
		}
		b[i] = letters[num.Int64()]
	}

	return string(b), nil
}

// generateRandomBytes generates secure random bytes using the provided reader.
//
// Parameters:
//   - n: The number of bytes to generate.
//   - reader: The io.Reader to use for generating random bytes.
//
// Returns: The generated random bytes or an error if the generation fails.
//
// Usage:
//
//	bytes, err := generateRandomBytes(10, rand.Reader)
//
// Example:
//
//	bytes, err := generateRandomBytes(10, rand.Reader)
//	fmt.Println(bytes)
func generateRandomBytes(n int, reader io.Reader) ([]byte, error) {
	if n <= 0 {
		return nil, newParseValueError("n should be greater than 0")
	}
	b := make([]byte, n)
	_, err := reader.Read(b)
	if err != nil {
		return nil, fmt.Errorf("could not generate random bytes: %w", err)
	}
	return b, nil
}