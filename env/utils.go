package env

import (
	"encoding"
	"reflect"
	"strings"
	"unicode"
)

// isSliceOfStructs checks if the field is a slice of structs.
//
// Parameters:
//   - sf: The reflect.StructField of the field.
//
// Returns:
//   - True if the field is a slice of structs, false otherwise.
func isSliceOfStructs(sf reflect.StructField) bool {
	t := sf.Type

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Struct
}

// asTextUnmarshaler gets the encoding.TextUnmarshaler from the reflect.Value.
//
// Parameters:
//   - v: The reflect.Value to get the encoding.TextUnmarshaler from.
//
// Returns:
//   - The encoding.TextUnmarshaler or nil if it doesn't exist.
func asTextUnmarshaler(v reflect.Value) encoding.TextUnmarshaler {
	if !v.IsValid() {
		return nil
	}

	if v.Kind() != reflect.Ptr && v.CanAddr() {
		v = v.Addr()
	} else if v.Kind() == reflect.Ptr && v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}

	tm, ok := v.Interface().(encoding.TextUnmarshaler)
	if !ok {
		return nil
	}
	return tm
}

// initialisePointer initialises the pointer if it's nil.
//
// Parameters:
//   - v: The reflect.Value to initialise.
func initialisePointer(v reflect.Value) {
	if v.Kind() != reflect.Ptr || !v.IsNil() {
		return
	}

	v.Set(reflect.New(v.Type().Elem()))
	v = v.Elem()
}

// updateReference updates the reference with the result.
//
// Parameters:
//   - ref: The reflect.Value to update.
//   - result: The reflect.Value to update with.
func updateReference(ref, result reflect.Value) {
	if ref.Kind() == reflect.Ptr {
		ref = ref.Elem()
	}

	if ref.CanSet() {
		ref.Set(result)
	}
}

// resolvePointer resolves the pointer to the value and type.
//
// Parameters:
//   - v: The reflect.Value to resolve.
//   - sfType: The reflect.Type of the struct field.
func resolvePointer(v reflect.Value, sfType reflect.Type) (reflect.Value, reflect.Type) {
	if v.Kind() == reflect.Ptr {
		return v.Elem(), sfType.Elem()
	}
	return v, sfType
}

// ensureTrailingUnderscore ensures that the prefix has a trailing underscore.
//
// Parameters:
//   - prefix: The prefix to ensure has a trailing underscore.
//
// Returns:
//   - The prefix with a trailing underscore.
func ensureTrailingUnderscore(prefix string) string {
	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		return prefix + "_"
	}
	return prefix
}

// findMaxIndex finds the maximum index from the prefixed environment map.
//
// Parameters:
//   - prefixedEnvMap: The map of prefixed environment variables.
//
// Returns:
//   - The maximum index.
func findMaxIndex(prefixedEnvMap map[int]bool) int {
	maxIndex := 0
	for idx := range prefixedEnvMap {
		if idx > maxIndex {
			maxIndex = idx
		}
	}
	return maxIndex
}

// getSeparators gets the separators from the struct field.
//
// Parameters:
//   - sf: The reflect.StructField of the field.
//
// Returns:
//   - The separator.
//   - The key value separator.
func getSeparators(sf reflect.StructField) (separator, keyValSeparator string) {
	separator = sf.Tag.Get(SeparatorEnv)
	if separator == "" {
		separator = ","
	}

	keyValSeparator = sf.Tag.Get(KeyValSeparatorEnv)
	if keyValSeparator == "" {
		keyValSeparator = ":"
	}

	return separator, keyValSeparator
}

// hasQuotePrefix checks if the source has a quote prefix.
// Such as a double quote (") or a single quote(').
//
// It takes in a byte slice to allow for additional checks.
//
// Parameters:
//   - src: The source to check for a quote prefix.
//
// Returns: The quote prefix and a boolean indicating if a quote prefix was found.
func hasQuotePrefix(src []byte) (byte, bool) {
	if len(src) == 0 {
		return 0, false
	}

	prefix := src[0]

	if prefix == CharDoubleQuote || prefix == CharSingleQuote {
		return prefix, true
	}

	return 0, false
}

// indexOfChar returns the position of the first occurrence of a character in a byte slice.
//
// This was found to be faster than bytes.IndexFunc.
//
// Parameters:
//   - src: The source to search for the character.
//   - c: The character to search for.
//
// Returns: The position of the first occurrence of the character.
func indexOfChar(src []byte, c rune) int {
	for i, v := range src {
		if v == byte(c) {
			return i
		}
	}
	return -1
}

// indexOfChars returns the position of the first occurrence of a one of the provided characters in a byte slice.
//
// This was found to be faster than bytes.IndexFunc.
//
// Parameters:
//   - src: The source to search for the character.
//   - c: The characters to search for.
//
// Returns: The position of the first occurrence of a character.
func indexOfChars(src []byte, c ...rune) int {
	for i, v := range src {
		for _, char := range c {
			if v == byte(char) {
				return i
			}
		}
	}
	return -1
}

// indexOfNonSpaceChar returns the position of the first non-whitespace character in a byte slice.
//
// This function may also return indexes of \n as the first character.
//
// Parameters:
//   - src: The source to search for the first non-whitespace character.
//
// Returns: The position of the first non-whitespace character.
func indexOfNonSpaceChar(src []byte) int {
	for i, c := range src {
		// Uses unicode.IsSpace to check for whitespace, some valid examples:
		// '\t', '\n', '\v', '\f', '\r', ' ', U+0085 (NEL), U+00A0 (NBSP).
		if !unicode.IsSpace(rune(c)) {
			return i
		}
	}
	return -1
}

// isSpace checks if a rune is a whitespace character, excludes '\n'.
//
// Used for getKeyValue to trim spaces before and after the key and value.
//
// Parameters:
//   - r: The rune to check if it is a whitespace character.
//
// Returns: True if the rune is a whitespace character, false otherwise.
func isSpace(r rune) bool {
	switch r {
	case '\t', '\v', '\f', '\r', ' ', 0x85, 0xA0:
		return true
	}
	return false
}
