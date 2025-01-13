package env

import (
	"reflect"
	"testing"
)

func TestGetRawEnv_ReturnsRawEnvVar(t *testing.T) {
	opts := Options{
		rawEnvVars: map[string]string{"VAR": "raw_value"},
		Env:        map[string]string{"VAR": "env_value"},
	}
	result := opts.getRawEnv("VAR")
	if result != "raw_value" {
		t.Errorf("Expected raw_value, got %s", result)
	}
}

func TestGetRawEnv_ReturnsEnvVarWhenRawEnvVarNotSet(t *testing.T) {
	opts := Options{
		rawEnvVars: map[string]string{},
		Env:        map[string]string{"VAR": "env_value"},
	}
	result := opts.getRawEnv("VAR")
	if result != "env_value" {
		t.Errorf("Expected env_value, got %s", result)
	}
}

func TestGetRawEnv_ExpandsEnvVar(t *testing.T) {
	opts := Options{
		rawEnvVars: map[string]string{"VAR": "raw_${NESTED}"},
		Env:        map[string]string{"NESTED": "value"},
	}
	result := opts.getRawEnv("VAR")
	if result != "raw_value" {
		t.Errorf("Expected raw_value, got %s", result)
	}
}

func TestGetRawEnv_ReturnsEmptyStringWhenVarNotSet(t *testing.T) {
	opts := Options{
		rawEnvVars: map[string]string{},
		Env:        map[string]string{},
	}
	result := opts.getRawEnv("VAR")
	if result != "" {
		t.Errorf("Expected empty string, got %s", result)
	}
}

func TestWithPrefix_AppendsPrefix(t *testing.T) {
	opts := Options{Prefix: "PREFIX_"}
	sf := reflect.StructField{Tag: `envPrefix:"NEW_"`}
	newOpts := opts.withPrefix(sf)
	if newOpts.Prefix != "PREFIX_NEW_" {
		t.Errorf("Expected PREFIX_NEW_, got %s", newOpts.Prefix)
	}
}

func TestWithSliceEnvPrefix_AppendsIndexToPrefix(t *testing.T) {
	opts := Options{Prefix: "PREFIX_"}
	newOpts := opts.withSliceEnvPrefix(1)
	if newOpts.Prefix != "PREFIX_1_" {
		t.Errorf("Expected PREFIX_1_, got %s", newOpts.Prefix)
	}
}

func TestFilterPrefixedEnvVars(t *testing.T) {
	opts := Options{
		Prefix: "PREFIX_",
		Env: map[string]string{
			"PREFIX_0_FOO":  "foo",
			"PREFIX_1_BAR":  "bar",
			"Baz":           "baz",
			"FPREFIX_0_FOO": "fake",
		},
	}

	result := opts.filterPrefixedEnvVars()
	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}

	badOpts := Options{
		Prefix: "PREFIX_0",
		Env: map[string]string{
			"PREFIX_0_": "foo",
		},
	}

	badResult := badOpts.filterPrefixedEnvVars()
	if len(badResult) != 0 {
		t.Errorf("Expected 0 result, got %d", len(badResult))
	}
}

func TestDefaultOptions_ReturnsInitialOptions(t *testing.T) {
	opts := defaultOptions()
	if opts.Prefix != "" {
		t.Errorf("Expected empty prefix, got %s", opts.Prefix)
	}
	if len(opts.Env) == 0 {
		t.Errorf("Expected non-empty Env map")
	}
	if len(opts.rawEnvVars) != 0 {
		t.Errorf("Expected empty rawEnvVars map")
	}
}

func BenchmarkGetRawEnv(b *testing.B) {
	opts := Options{
		rawEnvVars: map[string]string{"VAR": "raw_value"},
		Env:        map[string]string{"VAR": "env_value"},
	}
	for i := 0; i < b.N; i++ {
		opts.getRawEnv("VAR")
	}
}

func BenchmarkFilterPrefixedEnvVars(b *testing.B) {
	opts := Options{
		Prefix: "PREFIX_",
		Env: map[string]string{
			"PREFIX_0_FOO":  "foo",
			"PREFIX_1_BAR":  "bar",
			"Baz":           "baz",
			"FPREFIX_0_FOO": "fake",
		},
	}
	for i := 0; i < b.N; i++ {
		opts.filterPrefixedEnvVars()
	}
}
