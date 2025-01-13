package env

import (
	"strings"
)

// On Windows, environment variables could start with '='?
// See env_windows.go in the Go source(2023): https://github.com/golang/go/blob/master/src/syscall/env_windows.go#L63
// See their original reference (2010): https://devblogs.microsoft.com/oldnewthing/20100506-00/?p=14133
// See current reference (2021): https://learn.microsoft.com/en-us/windows/win32/procthread/environment-variables
// If an issue arises, please create an issue.

// toMap converts a slice of environment variables into a map.
//
// Parameters:
//   - env: A slice of environment variables.
//
// Returns:
//   - A map of environment variables.
func toMap(env []string) map[string]string {
	r := make(map[string]string, len(env))
	for _, e := range env {
		if i := strings.IndexByte(e, '='); i != -1 {
			// Split at the first '=' character, :i gets the key, i+1: gets the value.
			r[e[:i]] = e[i+1:]
		}
	}
	return r
}
