package env

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

type FileOpener func(string) (*os.File, error)

// ParseFromFilesIntoStruct loads environment variables from a file into a struct.
//
// Parameters:
//   - filenames: The filenames to load the environment variables from.
//
// Example:
//
//	err := env.ParseFromFilesIntoStruct(&config, ".env")
//
// Returns: An error if the parsing fails.
//
// Note: If no filenames are provided, it will default to ".env".
// When successful, the struct referenced by v will be updated.
//
// All processing occurs in ParseWithOpts.
func ParseFromFilesIntoStruct(v interface{}, filenames ...string) error {
	if len(filenames) == 0 {
		filenames = []string{".env"}
	}

	var err error
	envMap := make(map[string]string)

	for _, filename := range filenames {
		var tEnvMap map[string]string
		if tEnvMap, err = parseFile(filename, os.Open); err != nil {
			return err
		}

		for key, val := range tEnvMap {
			envMap[key] = val
		}
	}

	// While this could be used with ParseFromFileIntoStruct, it would error every time a required key is missing.
	// For example, a .database.env file could be used to load database creds,
	// but the .env file would determine the database of choice.
	return ParseWithOpts(v, Options{
		Env: envMap,
	})
}

// ParseFromFileIntoStruct loads environment variables from a file into a struct.
//
// This function may be slightly faster than ParseFromFilesIntoStruct as it lacks the overhead of iterating over the filenames.
//
// Parameters:
//   - filenames: The filenames to load the environment variables from.
//
// Example:
//
//	err := env.ParseFromFileIntoStruct(&config, ".env")
//
// Returns: An error if the parsing fails.
//
// Note: If no filenames are provided, it will default to ".env".
// When successful, the struct referenced by v will be updated.
//
// All processing occurs in ParseWithOpts.
func ParseFromFileIntoStruct(v interface{}, filename string) error {
	envMap, err := parseFile(filename, os.Open)

	if err != nil {
		return err
	}

	return ParseWithOpts(v, Options{
		Env: envMap,
	})
}

// ParseFromFiles loads environment variables from multiple file.
//
// It allows for a callback function to be called for each key-value pair, to allow for os.Setenv or to return back the key-value pair.
//
// Parameters:
//   - callbackFunc: The function to call for each key-value pair.
//   - filenames: The filenames to load the environment variables from.
//
// Example:
//
//	err := env.ParseFromFiles(func(key, value string) error {
//		return os.Setenv(key, value)
//	}, ".env")
//
// Note: does not support expanding variables.
func ParseFromFiles(callbackFunc func(key, value string) error, filenames ...string) error {
	if len(filenames) == 0 {
		filenames = []string{".env"}
	}

	var err error
	for _, filename := range filenames {
		if err = ParseFromFile(callbackFunc, filename); err != nil {
			return err
		}
	}

	return nil
}

// ParseFromFile loads environment variables from a file.
//
// It allows for a callback function to be called for each key-value pair, to allow for os.Setenv or to return back the key-value pair.
//
// This function might be slightly faster than ParseFromFiles as it lacks the overhead of iterating over the filenames.
//
// Parameters:
//   - callbackFunc: The function to call for each key-value pair.
//   - filename: The filename to load the environment variables from.
//
// Example:
//
//	err := env.ParseFromFiles(func(key, value string) error {
//		return os.Setenv(key, value)
//	}, ".env")
//
// Note: does not support expanding variables.
func ParseFromFile(callbackFunc func(key, value string) error, filename string) error {
	var err error
	var envMap map[string]string
	if envMap, err = parseFile(filename, os.Open); err != nil {
		return err
	}

	for key, val := range envMap {
		err = callbackFunc(key, val)
		if err != nil {
			return err
		}
	}

	return nil
}

// parseFile loads environment variables from a file into a map.
//
// Opener is required, as it allows for testing.
//
// Parameters:
//   - filename: The filename to load the environment variables from.
//   - opener: The function to open the file.
func parseFile(filename string, opener FileOpener) (map[string]string, error) {
	file, err := opener(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var envMap map[string]string
	envMap, err = readWithIO(file)

	if err != nil {
		return nil, err
	}

	return envMap, nil
}

// readWithIO reads the environment variables from an io.Reader, calling parseEnvFileBytes.
//
// Parameters:
//   - r: The io.Reader to read the environment variables from.
//
// Returns: The map of environment variables and an error if the reading fails.
func readWithIO(r io.Reader) (map[string]string, error) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		return nil, err
	}

	var envMap map[string]string
	envMap, err = parseEnvFileBytes(bytes.Replace(buf.Bytes(), []byte("\r\n"), []byte("\n"), -1))
	if err != nil {
		return nil, err
	}

	return envMap, err
}

// parseEnvFileBytes parses the environment variables from a byte slice.
//
// Parameters:
//   - src: The byte slice to parse the environment variables from.
//
// Returns: The map of environment variables and an error if the parsing fails.
func parseEnvFileBytes(src []byte) (map[string]string, error) {
	envMap := make(map[string]string)

	if len(src) == 0 {
		return envMap, errors.New("empty file")
	}

	for {
		src = getStart(src)
		if src == nil {
			return envMap, nil
		}

		var key string
		var value string
		var err error

		key, value, src, err = getKeyValue(src)

		if err != nil {
			return nil, err
		}

		envMap[key] = value
	}
}

// getStart returns position of the first non-whitespace character
//
// Parameters:
//   - src: The source to search for the first non-whitespace character.
//
// Returns: The position of the first non-whitespace character.
func getStart(src []byte) []byte {
	// Get the position of the first non-whitespace character
	// This function may also get \n as the first character
	pos := indexOfNonSpaceChar(src)

	if pos == -1 {
		return nil
	}

	src = src[pos:]

	// Check if the first character is a comment
	// For example a line with "    # comment" is a comment line
	// If the first character is not a comment, it could be a key-value pair
	if src[0] != CharComment {
		return src
	}

	// Since it was a comment, get the position of the next line
	pos = indexOfChar(src, '\n')
	if pos == -1 {
		// If there's no newline, it's the end of the file
		return nil
	}

	// Recurse to see if the next line satisfies the conditions
	return getStart(src[pos:])
}

// getKeyValue returns the key, value, and remaining bytes after the key-value pair.
//
// Parameters:
//   - src: The source to search for the key-value pair.
//
// Returns:
//   - The key.
//   - The value.
//   - The remaining bytes after the key-value pair.
//   - An error if the key-value pair is invalid.
func getKeyValue(src []byte) (string, string, []byte, error) {
	var key string
	var value string
	var err error
	key, src, err = getKey(src)

	if src == nil {
		return key, value, src, err
	} else if err != nil {
		return "", "", src, err
	}

	value, src, err = getValue(src)

	if err != nil {
		return "", "", nil, err
	}

	return key, value, src, nil
}

// getValue returns the value and remaining bytes after the value for getKeyValue.
//
// Parameters:
//   - src: The source to search for the value.
//
// Returns:
//   - The value.
//   - The remaining bytes after the value.
//   - An error if the value is invalid.
func getValue(src []byte) (string, []byte, error) {
	quote, hasQuote := hasQuotePrefix(src)

	if hasQuote {
		return getValueWithinQuotes(src, quote)
	}

	return getValueWithoutQuotes(src)
}

// getValueWithinQuotes returns the value and remaining bytes after the value for getKeyValue.
//
// Parameters:
//   - src: The source to search for the value.
//   - quote: The quote prefix it can either be a double quote (") or a single quote(').
//
// Returns:
//   - The value.
//   - The remaining bytes after the value.
//   - An error if the value is invalid.
func getValueWithinQuotes(src []byte, quote byte) (string, []byte, error) {
	for i := 1; i < len(src); i++ {
		if src[i] != quote {
			continue
		}

		// If previous char is \, it's an escaped quote
		// This is for sure as this current loop iteration is not a confirmed matching quote
		if src[i-1] == '\\' {
			fmt.Println("escaped quote")
			continue
		}

		isQuote := func(r rune) bool {
			return r == rune(quote)
		}

		value := string(bytes.TrimLeftFunc(bytes.TrimRightFunc(src[0:i], isQuote), isQuote))

		if quote == CharDoubleQuote {
			value = unescapeQuotes(value)
		}

		return value, src[i+1:], nil
	}

	return "", nil, errors.New("unterminated closing quote")
}

// unescapeQuotes unescapes quotes in a string, such as \n and \r.
//
// This could be done with regex, but it was seen with a 161% performance improvement.
//
// Parameters:
//   - s: The string to unescape quotes from.
//
// Returns: The string with unescaped quotes.
func unescapeQuotes(s string) string {
	var builder strings.Builder

	// Pre-allocate the builder with the length of the input string
	builder.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				builder.WriteByte('\n')
				i++
			case 'r':
				builder.WriteByte('\r')
				i++
			default:
				builder.WriteByte(s[i+1])
				i++
			}
		} else {
			builder.WriteByte(s[i])
		}
	}

	return builder.String()
}

// getValueWithoutQuotes returns the value and remaining bytes after the value for getValue.
//
// Parameters:
//   - src: The source to search for the value.
//
// Returns:
//   - The value.
//   - The remaining bytes after the value.
//   - An error if the value is invalid.
func getValueWithoutQuotes(src []byte) (string, []byte, error) {
	endOfLine := findEndOfLine(src)
	if endOfLine == 0 {
		// Empty line or end of file
		return "", nil, nil
	}

	line := src[:endOfLine]
	src = src[endOfLine:]

	value := extractValueFromLine(line)
	return value, src, nil
}

// findEndOfLine returns the position of the end of the line.
//
// Parameters:
//   - src: The source to search for the end of the line.
//
// Returns: The position of the end of the line.
func findEndOfLine(src []byte) int {
	endOfLine := indexOfChars(src, '\n', '\r')
	if endOfLine == -1 {
		return len(src) // End of file
	}
	return endOfLine
}

// extractValueFromLine extracts the value from a line.
//
// Parameters:
//   - line: The line to extract the value from.
//
// Returns: The value.
func extractValueFromLine(line []byte) string {
	endOfVar := len(line)
	for i := 1; i < endOfVar; i++ {
		if line[i] == CharComment && isSpace(rune(line[i-1])) {
			endOfVar = i
			break
		}
	}
	return string(bytes.TrimFunc(line[:endOfVar], isSpace))
}

// getKey returns the key and remaining bytes after the key for getKeyValue.
//
// Parameters:
//   - src: The source to search for the key.
//
// Returns:
//   - The key.
//   - The remaining bytes after the key.
//   - An error if the key is invalid.
func getKey(src []byte) (string, []byte, error) {
	src = bytes.TrimLeftFunc(src, isSpace) // Trim leading spaces
	key, remaining, err := extractKey(src)
	if err != nil {
		return "", remaining, err
	}
	err = validateKey(key)
	if err != nil {
		return "", remaining, err
	}
	return key, remaining, nil
}

// extractKey extracts the key and remaining bytes after the separator.
//
// Parameters:
//   - src: The source to search for the key.
//
// Returns:
//   - The key.
//   - The remaining bytes after the separator.
//   - An error if the key is invalid.
func extractKey(src []byte) (string, []byte, error) {
	for i := 0; i < len(src); i++ {
		char := rune(src[i])
		if isSpace(char) {
			continue
		}
		if char == '=' || char == ':' {
			// Extract the key and remaining bytes after separator
			key := string(bytes.TrimRightFunc(src[:i], isSpace))
			return key, src[i+1:], nil
		}
	}
	return "", src, errors.New("key-value separator not found")
}

// validateKey validates the key against whether it starts with a capital letter.
//
// Parameters:
//   - key: The key to validate.
//
// Returns: An error if the key is invalid.
func validateKey(key string) error {
	if key == "" || !unicode.IsLetter(rune(key[0])) || !unicode.IsUpper(rune(key[0])) {
		return errors.New("invalid key: must start with a capital letter")
	}
	return nil
}
