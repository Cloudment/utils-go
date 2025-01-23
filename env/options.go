package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Tags used for the struct tags, some are options within the Env tag.
// Allows for some customisation of the struct tags.
const (
	// Env is the tag to use for the environment variables, this could be renamed to json to use json tags.
	// However, env is recommended so that DATABASE_URL is used as a tag instead of database_url.
	Env = "env"
	// DefaultEnv is the default tag to use when looking up the tag, this is used when no env is found.
	DefaultEnv = "envDefault"
	// RequiredEnv is the option for specifying that the field is required.
	RequiredEnv = "required"
	// ExpandEnv is the option for specifying that the field should be expanded, this is used when expanding variables.
	ExpandEnv = "expand"
	// InitEnv is the option for specifying that the field should be initialised.
	InitEnv = "init"
	// PrefixEnv is the option for specifying the prefix to use when looking up the tag.
	PrefixEnv = "envPrefix"
	// UnsetEnv is the option for specifying that the field should be unset/deleted from os.Environ().
	UnsetEnv = "unset"
	// SeparatorEnv is the option for specifying the separator like , for slices.
	SeparatorEnv = "envSeparator"
	// KeyValSeparatorEnv is the option for specifying the key value separator like = for slices.
	KeyValSeparatorEnv = "envKeyValSeparator"

	// File specific

	// CharComment is the definition of the char for comments like # hi this is a comment
	CharComment = '#'
	// CharSingleQuote is the definition of the char for single quotes like 'hello'
	CharSingleQuote = '\''
	// CharDoubleQuote is the definition of the char for double quotes like "hello"
	CharDoubleQuote = '"'
)

// Options contains the options to pass through the parser.
//
// Example uses of Options might be to add in additional Env keys and values which could be taken from other sources.
// Such as loading from a secure store, then unsetting all the environment variables after load.
type Options struct {
	// Env keys and values. This is fetched from os.Environ()
	Env map[string]string

	// Prefix is the prefix to apply before the key. Usually taken from the struct tag.
	//
	// Such as "PREFIX_"
	Prefix string

	// rawEnvVars is the raw environment variables, this is used when expanding variables.
	//
	// Appended everytime a new key is found. Otherwise, this could be used for additional configuration.
	rawEnvVars map[string]string
}

// getRawEnv is a helper function to get the raw environment variable in expanded form.
//
// Parameters:
//   - s: The string to get the raw environment variable for.
//
// Returns:
//   - The raw environment variable in expanded form.
//
// See: https://pkg.go.dev/os#Expand
func (opts Options) getRawEnv(s string) string {
	// All fields that are scanned are put into the rawEnvVars map.
	// This added with opts.rawEnvVars[tags.OwnKey] within the cmd.go file.
	val := opts.rawEnvVars[s]
	if val == "" {
		val = opts.Env[s]
	}
	return os.Expand(val, opts.getRawEnv)
}

// withPrefix returns a new Options struct with the prefix set.
//
// Parameters:
//   - sf: The struct field to get the prefix from a tag.
//
// Returns:
//   - A new Options struct with the prefix set.
//
// See: https://pkg.go.dev/reflect#StructField
func (opts Options) withPrefix(sf reflect.StructField) Options {
	opts.Prefix = opts.Prefix + sf.Tag.Get(PrefixEnv)
	return opts
}

// withSliceEnvPrefix returns a new Options struct with the prefix set.
//
// Parameters:
//   - index: The index to use for the prefix.
//
// Usage:
//
// Typically used when parsing a slice of structs, this will append the index to the prefix.
//   - prefix is "PREFIX_" and index is 0, the new prefix will be "PREFIX_0_"
//   - prefix is "PREFIX_" and index is 1, the new prefix will be "PREFIX_1_"
//
// Returns:
//   - A new Options struct with the prefix set.
func (opts Options) withSliceEnvPrefix(index int) Options {
	opts.Prefix = fmt.Sprintf("%s%d_", opts.Prefix, index)
	return opts
}

// filterPrefixedEnvVars filters the environment variables that have the current prefix.
//
// If it's currently in the struct of "PREFIX_", it will filter the environment variables that have "PREFIX_0_FOO".
//
// Since it's used in a slice environment, it would require the prefix to be of the form "PREFIX_" and not "PREFIX_0_".
// Then each Environment variable would be "PREFIX_0_FOO", "PREFIX_1_FOO", "PREFIX_2_FOO", etc.
//
// Returns: A map of the index of the environment variable.
//
// Note: mainly used for parseSliceOfStructs.
func (opts Options) filterPrefixedEnvVars() map[int]bool {
	prefixedEnvMap := make(map[int]bool)

	// prefixLen is the length of the prefix, it's as a variable to ensure it's only calculated once.
	prefixLen := len(opts.Prefix)

	for env := range opts.Env {
		if !strings.HasPrefix(env, opts.Prefix) {
			continue
		}

		// SplitN expects 2 underscores, if there's 3 it will ignore the last part.
		// For example PREFIX_2_a_b -> [2 a_b]
		parts := strings.SplitN(env[prefixLen:], "_", 2)
		// If there's not 2 parts or both are empty, it's not a valid environment variable.
		// For example: PREFIX_0_FOO -> [0 FOO]
		// For example: PREFIX_0_FOO_BAR -> [0 FOO_BAR]
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			continue
		}

		if idx, err := strconv.Atoi(parts[0]); err == nil {
			prefixedEnvMap[idx] = true
		}
	}
	return prefixedEnvMap
}

// defaultOptions is the initial options to use when parsing the struct.
//
// This is used to clean up the parameters during parsing.
//
// Returns:
//   - An env map with the environment variables from os.Environ().
//   - An empty prefix, as this is the root struct.
//   - An empty rawEnvVars map, as this is the root struct.
//
// Note:  This cannot be a pointer value, as it's modified within the parseStruct function for additional prefixes
func defaultOptions() Options {
	return Options{
		Env:        toMap(os.Environ()),
		Prefix:     "",
		rawEnvVars: make(map[string]string),
	}
}
