package utils

import "fmt"

type ParseValueError struct {
	Desc string
}

// newParseValueError creates a new ParseValueError with the given description.
// It's a helper function for better error handling within the library and to serve as an example.
func newParseValueError(desc string) error {
	return &ParseValueError{Desc: desc}
}

func (e ParseValueError) Error() string {
	return fmt.Sprintf("input error: %s", e.Desc)
}
