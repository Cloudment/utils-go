package env

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// FieldTags contains the tags that can be used to customise the behavior of the parser.
//
// Example usages of tags are shown within the struct.
type FieldTags struct {
	// OwnKey is the key that describes the current field within its own context/struct.
	//
	// If set to "-" or omitted, the field is Ignored.
	OwnKey string `env:"key"`
	// Key is the key to use when looking up the environment variable.
	//
	// This is a concatenation of the Prefix and the OwnKey.
	// There is no ability to set this manually.
	Key string
	// Prefix is the prefix to use when looking up the environment variable.
	Prefix string `envPrefix:"prefix"`
	// Default is the default value to use if the environment variable is not set.
	//
	// Cannot be used with Required.
	Default string `envDefault:"default"`
	// Required is set to true if the field is required, use `required:"true"` to set it as required.
	//
	// Cannot be used with Default.
	Required bool `env:",required"`
	// Ignored is set to true if the field should be ignored.
	//
	// Cannot be used with OwnKey.
	Ignored bool `env:"-"`
	// Init is the value to use when the struct is initialised.
	//
	// Uses when you have a pointer to a struct, with inner fields that are variables.
	//
	// Use case:
	//
	//	type Config struct {
	//		Inner *InnerConfig `envPrefix:"INNER_" env:",init"`
	//	}
	//
	//	type InnerConfig struct {
	//		Host string `env:"HOST"`
	//	}
	//
	// In this case, InnerConfig will be initialised with the value of Host,
	// only if INNER_HOST is set as an environment variable.
	Init bool `env:"init"`
	// Expand, uses Default as a template to expand environment variables.
	//
	// Use case:
	//
	//	type Config struct {
	//      Host string `env:"HOST" envDefault:"127.0.0.1"`
	//		Port int    `env:"PORT" envDefault:"8080"`
	//		URL  string `env:"URL" envDefault:"http://${HOST}:${PORT}"`
	//	}
	//
	// In this case, URL will be expanded to "http://127.0.0.1:8080".
	Expand bool `env:",expand"`
	// Unset is provided to unset the environment variable after it has been set.
	//
	// This is useful when you want to set a value, but not keep it in the environment like a password.
	Unset bool `env:",unset"`
}

// Parse parses a struct containing `env` tags and loads its values from environment variables.
//
// This function uses the default options.
//
// Parameters:
//
//   - v: A pointer to a struct containing `env` tags.
//
// Returns: An error if the parsing failed. If successful, it will return nil.
//
// Note: This function is a wrapper around ParseWithOpts. When successful, the struct referenced by v will be updated.
func Parse(v interface{}) error {
	opts := defaultOptions()

	return ParseWithOpts(v, opts)
}

// ParseWithOpts parses a struct containing `env` tags and loads its values from
// environment variables.
//
// This function allows you to pass in options to customise the behavior of the parser.
//
// Parameters:
//
//   - v: A pointer to a struct containing `env` tags.
//   - opts: The options to use when parsing the struct.
//
// Returns: An error if the parsing failed. If successful, it will return nil.
//
// Note: When successful, the struct referenced by v will be updated.
func ParseWithOpts(v interface{}, opts Options) error {
	if v == nil || reflect.ValueOf(v).Kind() != reflect.Ptr {
		return errors.New("expected a pointer to a valid struct")
	}

	// Currently, there is no prefix as it's the root struct.
	// After the first loop, any structs within this struct will have a prefix.
	err := parseInterface(v, opts)

	if err != nil {
		return err
	}

	return nil
}

// parseInterface parses an interface and sets the values of the struct.
//
// A normal process tree would look like this:
//
// parseInterface -> parseStruct -> parseField -> parseStruct -> parseField etc.
//
// It may also call this function, parseInterface, if it finds a pointer to a struct.
//
// Parameters:
//
//   - i: The interface to parse, must be a pointer to a struct.
//   - opts: The options to use when parsing the struct.
//
// Returns: An error if the parsing failed. If successful, it will return nil.
func parseInterface(i interface{}, opts Options) error {
	v := reflect.ValueOf(i)

	if v.IsNil() || v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("expected a pointer to a valid struct")
	}

	return parseStruct(v.Elem(), opts)
}

// parseStruct parses a struct and sets the values of the fields.
//
// This will be called recursively for any fields that are structs.
// It starts at the root struct, that was provided within Parse or ParseWithOpts.
//
// Works it way through the fields of the struct and sets the values, as well as adjusts prefix if needed.
//
// Parameters:
//
//   - ref: The reflect.Value of the struct to parse.
//   - opts: The options to use when parsing the struct.
//
// Returns: An error if the parsing failed. If successful, it will return nil.
func parseStruct(ref reflect.Value, opts Options) error {
	if ref.Kind() == reflect.Ptr {
		ref = ref.Elem()
	}

	if ref.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct, but got %v", ref.Kind())
	}

	refType := ref.Type()

	// Loop through the fields of the struct.
	for i := 0; i < refType.NumField(); i++ {
		f := ref.Field(i)
		sf := refType.Field(i)

		// While aggregating errors could be possible here,
		// if there is an issue, it should be fixed before continuing,
		// minimising wasted processing if there is an issue.
		if err := parseField(f, sf, opts); err != nil {
			return err
		}
	}

	return nil
}

// parseField parses a field and sets the value of the field.
//
// This will be called recursively for any fields that are structs.
// Through parseField -> parseStruct -> parseField -> parseStruct etc.
//
// Parameters:
//
//   - v: The reflect.Value of the field to parse.
//   - sf: The reflect.StructField of the field to parse.
//   - opts: The options to use when parsing the field.
//
// Returns: An error if the parsing failed. If successful, it will return nil.
func parseField(v reflect.Value, sf reflect.StructField, opts Options) error {
	if !v.CanSet() {
		return nil
	}

	var err error

	// Interface checking is done first, as it will require going back to parseInterface
	if err = handlePointerStruct(v, sf, opts); err != nil {
		return err
	}

	// Tags are parsed to determine the behavior of the field.
	// Such as `env:"key"` or `env:"key,required"` for required fields.
	tags := parseFieldTags(sf, opts)

	// If the field does not have a key, it's ignored.
	// It may also specify to be ignored with `env:"-"`
	if tags.Ignored {
		return nil
	}

	// set's a value to the field, if it's not empty.
	if err = setField(v, sf, tags, opts); err != nil {
		return err
	}

	initialisePointer(v)

	// If the field is a slice of structs, it will be handled differently.
	// It may also be another struct, which will be handled differently.
	if err = handleStructOrSlice(v, sf, opts, tags); err != nil {
		return err
	}

	return nil
}

// handlePointerStruct handles a pointer to a struct.
//
// If the pointer is not nil, it will call parseInterface to parse the struct.
// If the field is a struct and can be addressed, it will call parseInterface to parse the struct.
//
// Parameters:
//
//   - v: The reflect.Value of the field to parse.
//   - sf: The reflect.StructField of the field to parse.
//   - opts: The options to use when parsing the field.
//
// Returns: An error if the parsing failed. If successful or not applicable, it will return nil.
func handlePointerStruct(v reflect.Value, sf reflect.StructField, opts Options) error {
	if v.Kind() == reflect.Invalid {
		return errors.New("expected a valid reflect.Value")
	}

	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		return parseInterface(v.Interface(), opts.withPrefix(sf))
	}

	if v.Kind() == reflect.Struct && v.CanAddr() && v.Type().Name() == "" {
		return parseInterface(v.Addr().Interface(), opts.withPrefix(sf))
	}

	return nil
}

// handleStructOrSlice handles a struct or a slice of structs.
//
// If the field is a struct, it will call parseStruct to parse the struct.
// If the field is a slice of structs, it will call parseSliceOfStructs to parse the slice.
//
// Parameters:
//
//   - v: The reflect.Value of the field to parse.
//   - sf: The reflect.StructField of the field to parse.
//   - opts: The options to use when parsing the field.
//   - tags: The FieldTags of the field to parse.
//
// Returns: An error if the parsing failed. If successful or not applicable, it will return nil.
func handleStructOrSlice(v reflect.Value, sf reflect.StructField, opts Options, tags FieldTags) error {
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		return parseInterface(v.Interface(), opts.withPrefix(sf))
	}

	if v.Kind() == reflect.Struct {
		if v.CanAddr() {
			return parseStruct(v.Addr(), opts.withPrefix(sf))
		}
		return fmt.Errorf("cannot address struct field: %s", sf.Name)
	}

	if isSliceOfStructs(sf) {
		return parseSliceOfStructs(v, opts.withPrefix(sf))
	}

	// If the field is nil, it will be initialised.
	// An example of this might be a map, where the map is nil.
	invalidPtr := v.Kind() == reflect.Ptr && v.IsNil()
	if tags.Init && invalidPtr {
		v.Set(reflect.New(v.Type().Elem()))
		v = v.Elem()
	}

	return nil
}

// setField sets the value of the field.
//
// If the field is a TextUnmarshaler, it will call UnmarshalText to set the value.
// If the field is a pointer, it will resolve the pointer and the type.
// If the field is a custom type like a Location/Timezone, it will call the special type handler.
//
// Parameters:
//
//   - v: The reflect.Value of the field to parse.
//   - sf: The reflect.StructField of the field to parse.
//   - tags: The FieldTags of the field to parse.
//   - opts: The options to use when parsing the field.
//
// Returns: An error if the parsing failed. If successful, it will return nil.
func setField(v reflect.Value, sf reflect.StructField, tags FieldTags, opts Options) error {
	val, err := resolveValue(tags, opts)
	if err != nil {
		return err
	}

	if val == "" {
		return nil
	}

	handleUnset(tags)

	if tm := asTextUnmarshaler(v); tm != nil {
		return tm.UnmarshalText([]byte(val))
	}

	vp, sfType := resolvePointer(v, sf.Type)

	var ok bool
	if ok, err = applyParser(vp, sfType, val); ok {
		// If it's successful, return nil otherwise it would run handleSpecialTypes
		// which would return an error if it could not be found.
		return nil
	} else if err != nil {
		return err
	}

	// If it's a Slice or Map, it will be handled differently.
	return handleSpecialTypes(v, val, sf)
}

// resolveValue resolves the value of the field.
// This uses the opts.Env map to get the value of the field.
//
// If expanding is set, it will expand the value.
//
// Parameters:
//
//   - tags: The FieldTags of the field to parse.
//   - opts: The options to use when parsing the field.
//
// Returns: The value of the field, or an error if the value could not be resolved.
func resolveValue(tags FieldTags, opts Options) (string, error) {
	val, exists := opts.Env[tags.Key]
	if (tags.Key == "" || !exists || val == "") && tags.Default != "" {
		val = tags.Default
	}

	if opts.rawEnvVars == nil {
		opts.rawEnvVars = make(map[string]string)
	}

	if tags.Expand {
		val = os.Expand(val, opts.getRawEnv)
	}

	opts.rawEnvVars[tags.OwnKey] = val

	if tags.Required && (tags.OwnKey == "" || val == "") {
		return "", fmt.Errorf("required environment variable not set: %s", tags.Key)
	}

	return val, nil
}

// handleUnset unsets the environment variable if the Unset tag is set.
//
// Parameters:
//
//   - tags: The FieldTags of the field to parse.
//
// Returns: Nothing.
//
// Note: This function is called after the value has been set.
func handleUnset(tags FieldTags) {
	if !tags.Unset || tags.Key == "" {
		return
	}

	defer func(key string) {
		// Even though it might fail, it's not critical.
		// Logging this error might give a hint this system is vulnerable
		// to environment variable attacks as it explicitly states it was not unset.
		_ = os.Unsetenv(key)
	}(tags.Key)
}

// applyParser applies the parser to the value of the field.
//
// If the field is a special type (Duration/Location), it will use typeParsers for the type.
// If the field is a general type (int/bool), it will use parsers for the kind.
//
// Parameters:
//
//   - v: The reflect.Value of the field to parse.
//   - sfType: The reflect.Type of the field to parse.
//   - val: The value to parse.
//
// Returns: A boolean indicating if the parser was applied, or an error if the parser could not be applied.
func applyParser(v reflect.Value, sfType reflect.Type, val string) (bool, error) {
	if parseFunc, ok := typeParsers[sfType]; ok {
		parsedVal, err := parseFunc(val)
		if err != nil {
			return false, fmt.Errorf("failed to parse value: %v", err)
		}
		v.Set(reflect.ValueOf(parsedVal))
		return true, nil
	}

	if parseFunc, ok := parsers[sfType.Kind()]; ok {
		parsedVal, err := parseFunc(val)
		if err != nil {
			return false, fmt.Errorf("failed to parse value: %v", err)
		}
		v.Set(reflect.ValueOf(parsedVal).Convert(sfType))
		return true, nil
	}

	return false, nil
}

// parseFieldTags parses the tags of a field and returns a FieldTags struct.
//
// Tags are defined within options.go.
//
// Parameters:
//
//   - sf: The reflect.StructField of the field to parse.
//   - opts: The options to use when parsing the field.
//
// Returns: The FieldTags of the field.
//
// Note: This function is called before the value of the field is set.
func parseFieldTags(sf reflect.StructField, opts Options) FieldTags {
	// While slightly slower, having all tag lookups grouped looks slightly cleaner
	// To speed up the code, defaultValue can be moved after the ignore checking.
	// It would only save ~5 ns/op
	_, hasPrefix := sf.Tag.Lookup(PrefixEnv)
	env, hasEnv := sf.Tag.Lookup(Env)
	defaultValue := sf.Tag.Get(DefaultEnv)

	o := strings.Split(env, ",")
	ownKey, tags := o[0], o[1:]

	if (ownKey == "-" || !hasEnv) && !hasPrefix {
		return FieldTags{
			OwnKey:  ownKey,
			Ignored: true,
		}
	}

	res := FieldTags{
		OwnKey:   ownKey,
		Key:      opts.Prefix + ownKey,
		Default:  defaultValue,
		Required: false,
	}

	for _, tag := range tags {
		switch tag {
		case RequiredEnv:
			res.Required = true
		case ExpandEnv:
			res.Expand = true
		case InitEnv:
			res.Init = true
		case UnsetEnv:
			res.Unset = true
		}
	}

	return res
}
