package env

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ParserFunc defines the signature of a function that can be used within
// `Options`' `FuncMap`.
type ParserFunc func(v string) (interface{}, error)

var (
	// parsers is a map of `reflect.Kind` to `ParserFunc` that can be used to
	// parse a string value into a specific type.
	parsers = map[reflect.Kind]ParserFunc{
		reflect.Bool: func(v string) (interface{}, error) {
			return strconv.ParseBool(v)
		},
		reflect.Int: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 32)
			return int(i), err
		},
		reflect.Int8: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 8)
			return int8(i), err
		},
		reflect.Int16: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 16)
			return int16(i), err
		},
		reflect.Int32: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 32)
			return int32(i), err
		},
		reflect.Int64: func(v string) (interface{}, error) {
			return strconv.ParseInt(v, 10, 64)
		},
		reflect.Uint: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 32)
			return uint(i), err
		},
		reflect.Uint8: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 8)
			return uint8(i), err
		},
		reflect.Uint16: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 16)
			return uint16(i), err
		},
		reflect.Uint32: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 32)
			return uint32(i), err
		},
		reflect.Uint64: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 64)
			return i, err
		},
		reflect.Float32: func(v string) (interface{}, error) {
			f, err := strconv.ParseFloat(v, 32)
			return float32(f), err
		},
		reflect.Float64: func(v string) (interface{}, error) {
			return strconv.ParseFloat(v, 64)
		},
		reflect.String: func(v string) (interface{}, error) {
			return v, nil
		},
	}
	// typeParsers is a map of `reflect.Type` to `ParserFunc` that can be used to
	// parse a string value into a custom type.
	// Commonly for Duration and Location or other custom types.
	// Must return a non-pointer type.
	typeParsers = map[reflect.Type]ParserFunc{
		reflect.TypeOf(time.Nanosecond): func(v string) (interface{}, error) {
			d, err := time.ParseDuration(v)
			// Days are not always 24 hours long
			// See: https://github.com/golang/go/issues/11473
			// See: https://bigthink.com/starts-with-a-bang/day-isnt-24-hours/
			if err != nil && strings.Contains(err.Error(), "unknown unit \"d\"") {
				err = fmt.Errorf("use '24h' instead of '1d' for 24 hours: %w", err)
			}
			return d, err
		},
		reflect.TypeOf(time.Location{}): func(v string) (interface{}, error) {
			loc, err := time.LoadLocation(v)
			if err != nil {
				return nil, fmt.Errorf("unable to parse Location: %w", err)
			}
			return *loc, nil
		},
	}
)

// handleSpecialTypes handles special types like slices and maps.
//
// Parameters:
//   - v: The reflect.Value of the field.
//   - val: The value of the field.
//   - sf: The reflect.StructField of the field.
//
// Returns: An error if there is an issue handling the special type.
func handleSpecialTypes(v reflect.Value, val string, sf reflect.StructField) error {
	switch v.Kind() {
	case reflect.Slice:
		return handleSlice(v, val, sf)
	case reflect.Map:
		return handleMap(v, val, sf)
	default:
		return fmt.Errorf("unsupported type: %v for %v, %s", v.Kind(), sf.Type, sf.Name)
	}
}

// parseSliceOfStructs parses a slice of structs.
//
// Parameters:
//   - v: The reflect.Value of the field.
//   - opts: The Options to use when parsing the struct.
//
// Returns: An error if there is an issue parsing the slice of structs.
func parseSliceOfStructs(v reflect.Value, opts Options) error {
	opts.Prefix = ensureTrailingUnderscore(opts.Prefix)

	prefixedEnvMap := opts.filterPrefixedEnvVars()
	if len(prefixedEnvMap) == 0 {
		return nil
	}

	result, capacity := initialiseSlice(v, findMaxIndex(prefixedEnvMap))

	if err := populateSlice(result, v, prefixedEnvMap, capacity, opts); err != nil {
		return err
	}

	updateReference(v, result)

	return nil
}

// initialiseSlice initialises the slice with the correct length.
//
// Parameters:
//   - ref: The reflect.Value of the field.
//   - maxIndex: The maximum index of the slice.
//
// Returns:
//   - The reflect.Value of the slice.
//   - The capacity of the slice.
func initialiseSlice(v reflect.Value, maxIndex int) (reflect.Value, int) {
	sliceType := v.Type()
	isPointer := v.Kind() == reflect.Ptr
	if isPointer {
		sliceType = sliceType.Elem()
	}

	initialised := 0
	if !isPointer {
		initialised = v.Len()
	}

	newLength := maxIndex + 1
	capacity := newLength
	if initialised > newLength {
		capacity = initialised
	}

	result := reflect.MakeSlice(sliceType, capacity, capacity)
	return result, capacity
}

// populateSlice populates the slice with the correct values.
//
// Parameters:
//   - result: The reflect.Value of the slice.
//   - ref: The reflect.Value of the field.
//   - prefixedEnvMap: The map of the index of the environment variable.
//   - capacity: The capacity of the slice.
//   - opts: The Options to use when populating the slice.
//
// Returns: An error if there is an issue populating the slice.
func populateSlice(result, v reflect.Value, prefixedEnvMap map[int]bool, capacity int, opts Options) error {
	initialised := 0
	if v.Kind() != reflect.Ptr {
		initialised = v.Len()
	}

	for i := 0; i < capacity; i++ {
		item := result.Index(i)

		if i < initialised {
			item.Set(v.Index(i))
		}

		if !prefixedEnvMap[i] {
			continue
		}

		if err := parseStruct(item, opts.withSliceEnvPrefix(i)); err != nil {
			return err
		}
	}
	return nil
}

// parseTextUnmarshalers parses the text unmarshalers through parseElement.
//
// Parameters:
//   - field: The reflect.Value of the field.
//   - data: The slice of strings to parse.
//
// Returns: An error if there is an issue parsing the text unmarshalers.
func parseTextUnmarshalers(field reflect.Value, data []string) error {
	elemType := field.Type().Elem()
	length := len(data)
	slice := reflect.MakeSlice(reflect.SliceOf(elemType), length, length)

	for i, v := range data {
		if err := parseElement(slice.Index(i), elemType, v); err != nil {
			return err
		}
	}

	field.Set(slice)
	return nil
}

// parseElement parses the element using encoding.TextUnmarshaler.
//
// It calls the provided UnmarshalText method on the element.
// A use case of this is when a custom type implements encoding.TextUnmarshaler.
//
// Parameters:
//   - target: The reflect.Value of the target.
//   - elemType: The reflect.Type of the element.
//   - value: The value to parse.
//
// Returns: An error if there is an issue parsing the element.
func parseElement(target reflect.Value, elemType reflect.Type, value string) error {
	var item reflect.Value
	if target.Kind() == reflect.Ptr {
		item = reflect.New(elemType.Elem())
	} else {
		item = target.Addr()
	}

	tm, ok := item.Interface().(encoding.TextUnmarshaler)
	if !ok {
		return fmt.Errorf("type %v does not implement encoding.TextUnmarshaler", elemType)
	}

	if err := tm.UnmarshalText([]byte(value)); err != nil {
		return err
	}

	if target.Kind() == reflect.Ptr {
		target.Set(item)
	}
	return nil
}

// handleSlice handles the slice type.
//
// Parameters:
//   - v: The reflect.Value of the field.
//   - val: The value of the field.
//   - sf: The reflect.StructField of the field.
//
// Returns: An error if there is an issue handling the slice type.
func handleSlice(v reflect.Value, val string, sf reflect.StructField) error {
	separator := getSeparator(sf)
	parts := strings.Split(val, separator)

	elemType := sf.Type.Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	if _, ok := reflect.New(elemType).Interface().(encoding.TextUnmarshaler); ok {
		return parseTextUnmarshalers(v, parts)
	}

	parserFunc, err := getParserFunc(elemType)
	if err != nil {
		return err
	}

	result, err := parseSliceElements(parts, elemType, parserFunc, sf.Type.Elem())
	if err != nil {
		return err
	}

	v.Set(result)
	return nil
}

// getSeparator gets the separator from the struct field.
//
// Parameters:
//   - sf: The reflect.StructField of the field.
//
// Returns: The separator.
func getSeparator(sf reflect.StructField) string {
	separator := sf.Tag.Get(SeparatorEnv)
	if separator == "" {
		separator = ","
	}
	return separator
}

// getParserFunc gets the parser function for the element type.
//
// Parameters:
//   - elemType: The reflect.Type of the element.
//
// Returns:
//   - The parser function.
//   - An error if there is an issue getting the parser function.
func getParserFunc(elemType reflect.Type) (func(string) (interface{}, error), error) {
	if parserFunc, ok := typeParsers[elemType]; ok {
		return parserFunc, nil
	}

	if parserFunc, ok := parsers[elemType.Kind()]; ok {
		return parserFunc, nil
	}

	return nil, errors.New("unsupported type")
}

// parseSliceElements parses the slice elements.
//
// Parameters:
//   - parts: The slice of strings to parse.
//   - elemType: The reflect.Type of the element.
//   - parserFunc: The parser function.
//   - elemKind: The reflect.Type of the element kind.
//
// Returns:
//   - The reflect.Value of the slice.
//   - An error if there is an issue parsing the slice elements.
func parseSliceElements(parts []string, elemType reflect.Type, parserFunc func(string) (interface{}, error), elemKind reflect.Type) (reflect.Value, error) {
	result := reflect.MakeSlice(reflect.SliceOf(elemKind), 0, len(parts))
	for _, part := range parts {
		res, err := parserFunc(part)
		if err != nil {
			return reflect.Value{}, err
		}

		vp := reflect.ValueOf(res).Convert(elemType)
		if elemKind.Kind() == reflect.Ptr {
			ptr := reflect.New(elemType)
			ptr.Elem().Set(vp)
			vp = ptr
		}

		result = reflect.Append(result, vp)
	}
	return result, nil
}

// handleMap handles the map type by parsing the key and value.
//
// Parameters:
//   - field: The reflect.Value of the field.
//   - value: The value of the field.
//   - sf: The reflect.StructField of the field.
//
// Returns: An error if there is an issue handling the map type.
//
// Note: Can be used to parse a map of any supported type.
func handleMap(field reflect.Value, value string, sf reflect.StructField) error {
	keyParserFunc, elemParserFunc, err := getKeyAndElemParsers(sf.Type)
	if err != nil {
		return err
	}

	separator, keyValSeparator := getSeparators(sf)

	result := reflect.MakeMap(sf.Type)

	for _, part := range strings.Split(value, separator) {
		pairs := strings.SplitN(part, keyValSeparator, 2)
		if len(pairs) != 2 {
			return fmt.Errorf(`%q should be in "key%svalue" format`, part, keyValSeparator)
		}

		var key interface{}
		var elem interface{}

		key, err = keyParserFunc(pairs[0])
		if err != nil {
			return fmt.Errorf(`failed to parse key %q: %v`, pairs[0], err)
		}

		elem, err = elemParserFunc(pairs[1])
		if err != nil {
			return fmt.Errorf(`failed to parse value %q: %v`, pairs[1], err)
		}

		result.SetMapIndex(reflect.ValueOf(key).Convert(sf.Type.Key()), reflect.ValueOf(elem).Convert(sf.Type.Elem()))
	}

	field.Set(result)
	return nil
}

// getKeyAndElemParsers gets the key and element parsers for the map type.
//
// The key and element parsers may be different depending on map types.
//
// Parameters:
//   - mapType: The reflect.Type of the map.
//
// Returns:
//   - The key parser function.
//   - The element parser function.
//   - An error if there is an issue getting the key and element parsers.
func getKeyAndElemParsers(mapType reflect.Type) (keyParser, elemParser func(string) (interface{}, error), err error) {
	keyParserFunc, ok := parsers[mapType.Key().Kind()]
	if !ok {
		return nil, nil, errors.New("unsupported key type")
	}

	elemParserFunc, ok := parsers[mapType.Elem().Kind()]
	if !ok {
		return nil, nil, errors.New("unsupported element type")
	}

	return keyParserFunc, elemParserFunc, nil
}
