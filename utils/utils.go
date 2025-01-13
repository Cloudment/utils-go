package utils

import (
	"reflect"
	"strings"
)

// ValidatePagination checks if the page and limit are valid, returns the corrected values (page, limit).
//
// Parameters:
//   - page: The current page number.
//   - limit: The number of items per page.
//
// Returns:
//   - The corrected page number.
//   - The corrected limit.
//
// Usage:
//
//	page, limit := ValidatePagination(1, 50)
//
// Example:
//
//	page, limit := ValidatePagination(1, 50)
//	fmt.Println(page, limit)
//
// Note: The limit is set to 10 if it is 0, and capped at 100 if it exceeds 100. Negative page numbers are set to 0.
func ValidatePagination(page int, limit int) (newPage int, newLimit int) {
	if limit == 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	if page < 0 {
		page = 0
	}

	newPage = page
	newLimit = limit

	return
}

// ToAnySlice converts a slice of all types to a slice of interface{}.
func ToAnySlice[T any](collection []T) []any {
	result := make([]any, len(collection))
	for i := range collection {
		result[i] = collection[i]
	}
	return result
}

// GetOperatingSystemFromUserAgent returns the operating system from the user agent string.
//
// Parameters:
//   - userAgent: The user agent string.
//
// Returns: The operating system.
//
// Usage:
//
//	GetOperatingSystemFromUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/117.")
//	-> "Windows"
func GetOperatingSystemFromUserAgent(userAgent string) string {
	// TODO: Make more comprehensive, add more OSes
	if strings.Contains(userAgent, "iPhone") {
		return "iOS"
	} else if strings.Contains(userAgent, "Android") {
		return "Android"
	} else if strings.Contains(userAgent, "Windows") {
		return "Windows"
	} else if strings.Contains(userAgent, "Mac") {
		return "Mac"
	} else if strings.Contains(userAgent, "Linux") {
		return "Linux"
	} else {
		return "Unknown"
	}
}

// IsEqual compares two interfaces and returns true if they are equal.
//
// Mainly used for testing.
//
// Parameters:
//   - a: The first interface.
//   - b: The second interface.
//
// Returns: True if the interfaces are equal, false otherwise.
//
// Usage:
//
//	IsEqual(1, 1) // -> true
//	IsEqual(1, 2) // -> false
func IsEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if reflect.DeepEqual(a, b) {
		return true
	}
	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)
	return aValue == bValue
}
