package utils

import "testing"

func TestParseValueError_Error(t *testing.T) {
	err := newParseValueError("test")
	if err.Error() != "input error: test" {
		t.Errorf("Expected error to be 'input error: test', got %s", err.Error())
	}
}
