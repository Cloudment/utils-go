package utils

import (
	"reflect"
	"testing"
)

// OptionalQueryParams defines optional query parameters for a database query.
type OptionalQueryParams struct {
	ID string `query:"id = ?"`
	// Array is searching for a value in an array column, rather than searching for multiple values.
	Array string `query:"? = ANY(array)"`
}

func TestGeneratesQueryWithSingleCondition(t *testing.T) {
	params := OptionalQueryParams{ID: "123"}
	expectedQuery := "(id = ?)"
	expectedArgs := []interface{}{"123"}

	query, args := GormSearchQuery(params)

	if query != expectedQuery {
		t.Errorf("expected query to be '%s', got '%s'", expectedQuery, query)
	}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("expected args to be '%v', got '%v'", expectedArgs, args)
	}
}

func TestGeneratesQueryWithMultipleConditions(t *testing.T) {
	params := OptionalQueryParams{ID: "123", Array: "type1"}
	expectedQuery := "(id = ? AND ? = ANY(array))"
	expectedArgs := []interface{}{"123", "type1"}

	query, args := GormSearchQuery(params)

	if query != expectedQuery {
		t.Errorf("expected query to be '%s', got '%s'", expectedQuery, query)
	}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("expected args to be '%v', got '%v'", expectedArgs, args)
	}
}

func TestReturnsEmptyQueryWhenNoConditions(t *testing.T) {
	params := OptionalQueryParams{}
	expectedQuery := ""
	expectedArgs := []interface{}(nil)

	query, args := GormSearchQuery(params)

	if query != expectedQuery {
		t.Errorf("expected query to be '%s', got '%s'", expectedQuery, query)
	}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("expected args to be '%v', got '%v'", expectedArgs, args)
	}
}

func TestIgnoresEmptyFieldValues(t *testing.T) {
	params := OptionalQueryParams{ID: "", Array: "type1"}
	expectedQuery := "(? = ANY(array))"
	expectedArgs := []interface{}{"type1"}

	query, args := GormSearchQuery(params)

	if query != expectedQuery {
		t.Errorf("expected query to be '%s', got '%s'", expectedQuery, query)
	}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("expected args to be '%v', got '%v'", expectedArgs, args)
	}
}

func BenchmarkGormSearchQuery(b *testing.B) {
	for i := 0; i < b.N; i++ {
		params := OptionalQueryParams{ID: "123", Array: "type1"}
		GormSearchQuery(params)
	}
}
