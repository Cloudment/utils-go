package utils

import (
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"math/big"
	"time"
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
// Example:
//
//	bytes, err := GenerateRandomBytes(10)
//	fmt.Println(bytes)
func GenerateRandomBytes(n int) ([]byte, error) {
	return generateRandomBytes(n, rand.Reader)
}

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// GenerateRandomBytes generates secure random bytes using the default rand.Reader.
//
// Parameters:
//   - n: The number of bytes to generate.
//
// Returns: The generated random bytes or an error if the generation fails.
//
// Example:
//
//	bytes, err := GenerateRandomBytesGeneric[uint8](10)
//	fmt.Println(bytes)
func GenerateRandomBytesGeneric[T Integer](n T) ([]byte, error) {
	return generateRandomBytesGeneric(n, rand.Reader)
}

// GenerateRandomDuration generates a random duration with the given minimum (exclusive) and maximum (inclusive) values.
//
// This is similar to GenerateRandomNumber but generates a random duration instead.
//
// Parameters:
//   - min: The minimum value (inclusive).
//   - max: The maximum value (exclusive).
//   - unit: The unit of the duration.
//
// Returns: The generated random duration or an error if the generation fails.
//
// Example:
//
//	duration, err := GenerateRandomDuration(1, 10, time.Second)
//	fmt.Println(duration) // Output: 5s
func GenerateRandomDuration(min int, max int, unit time.Duration) (time.Duration, error) {
	return generateRandomDuration(min, max, unit, rand.Reader)
}

// generateRandomDuration generates a random duration with the given minimum (exclusive) and maximum (inclusive)
// values using the provided reader. It uses generateRandomNumber to generate the random number.
//
// Parameters:
//   - min: The minimum value (inclusive).
//   - max: The maximum value (exclusive).
//   - unit: The unit of the duration.
//
// Returns: The generated random duration or an error if the generation fails.
func generateRandomDuration(min int, max int, unit time.Duration, reader io.Reader) (time.Duration, error) {
	n, err := generateRandomNumber(min, max, reader)
	if err != nil {
		return 0, err
	}

	return time.Duration(n) * unit, nil
}

// GenerateOTP generates a secure random one-time password (OTP) of the given length.
//
// Parameters:
//   - length: The length of the generated OTP.
//
// Returns: The generated OTP or an error if the generation fails.
//
// Example:
//
//	otp, err := GenerateOTP(6)
//	fmt.Println(otp) // Output: "123
func GenerateOTP(length int) (otp int, err error) {
	return generateOTP(length, rand.Reader)
}

// generateOTP generates a secure random one-time password (OTP) of the given length using the provided reader.
//
// Parameters:
//   - length: The length of the generated OTP. The length should be greater than 0.
//   - reader: The io.Reader to use for generating random numbers.
//
// Returns: The generated OTP or an error if the generation fails.
func generateOTP(length int, reader io.Reader) (otp int, err error) {
	if length <= 0 {
		return 0, newParseValueError("length should be greater than 0")
	}

	// Avoids having to loop:
	//  for i := 1; i < length; i++ {
	//	   num *= 10
	//  }
	// As math.Pow10 stores precomputed values it's m
	//
	//  BenchmarkMathPow10                   	1000000000	         0.3159 ns/op
	//  BenchmarkLoop                        	379423045	         3.148  ns/op
	maxVal := int64(math.Pow10(length))
	minVal := int64(math.Pow10(length - 1))
	// maxVal is 100 (exclusive) if length is 2, which would do 10^2
	// minVal is 10 (inclusive) if length is 2 (2-1), which would do 10^1
	// Producing a range number between 10 and 99, which would allow for 89 random numbers for a length of 2.

	n, err := rand.Int(
		reader,
		big.NewInt(maxVal),
	)

	if err != nil {
		return 0, fmt.Errorf("could not generate OTP: %w", err)
	}

	otp = int(n.Int64())

	if otp < int(minVal) {
		otp += int(minVal)
	}

	return otp, nil
}

// generateRandomNumber generates a secure random integer
// between min (inclusive) and max (exclusive) using the provided reader.
//
// The minimum value should be less than the maximum value but also greater than or equal to 0.
//
// This function is not intended for negative values; for negative values, negate the value returned.
// Boolean values can be generated by using min 0, max 1. For example, 0 is false, 1 is true.
//
// Parameters:
//   - min: The minimum value (inclusive).
//   - max: The maximum value (exclusive).
//   - reader: The io.Reader to use for generating random numbers.
//
// Returns: The generated random number or an error if the generation fails.
func generateRandomNumber(min int, max int, reader io.Reader) (int, error) {
	if min >= max {
		return 0, newParseValueError("min should be less than max")
	} else if min < 0 {
		// rand.Int panics if the max value is less than 0, ensuring max is 1 or greater prevents this
		return 0, newParseValueError("min should be greater than or equal to 0")
	}

	// rand.int takes in a max value, generation occurs between 0 and the range size
	// if intended max is 50, min is 10, the range size is 40
	// If the random number generated is 0, when added back with the min value, it will be 10
	// Another example, if it was 39, it will be 49 when added back with the min value
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

// generateRandomBytes generates secure random bytes using the provided reader.
//
// Parameters:
//   - n: The number of bytes to generate.
//   - reader: The io.Reader to use for generating random bytes.
//
// Returns: The generated random bytes or an error if the generation fails.
func generateRandomBytesGeneric[T Integer](n T, reader io.Reader) ([]byte, error) {
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
