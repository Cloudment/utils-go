package utils

import (
	"reflect"
	"strings"
)

// GormSearchQuery generates a search query for GORM based on the provided parameters.
// The parameters should be a struct with fields that have a `query` tag.
//
// Parameters:
//
//   - params: A struct with fields that have a `query` tag.
//
// Returns: A string representing the query and a slice of arguments.
//
// Usage:
//
// The `query` tag should be in the format of `condition = ?`, where `condition` is the condition to be checked
// and `?` is the placeholder for the argument. It's identical to what would happen as part of a GORM query.
//
// Example:
//
//	type OptionalQueryParams struct {
//	 ID string `query:"id = ?"`
//	 Array string `query:"? = ANY(array)"`
//	}
//
//	params := OptionalQueryParams{ID: "123", Array: "type1"}
//	query, args := GormSearchQuery(params)
//
//	// query = "(id = ? AND ? = ANY(array))"
//	// args = ["123", "type1"]
//
// db = db.Where(query, args...).Find(&results)
func GormSearchQuery[p interface{}](params p) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	// While it looks like this code could be improved with caching, the advantage would be ~80 ns/op,
	// which compared to the rest of the function, the code would be more complex and harder to read/maintain.
	v := reflect.ValueOf(params)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i)

		// The use of the query tag allows any struct, even the GORM model struct, to be used with this function.
		queryTag := fieldType.Tag.Get("query")

		// Skip if no tag is provided or the field value is empty
		if queryTag == "" || fieldValue.IsZero() {
			continue
		}

		conditions = append(conditions, queryTag)
		args = append(args, fieldValue.Interface())
	}
	if len(conditions) > 0 {
		queryStr := "(" + strings.Join(conditions, " AND ") + ")"

		return queryStr, args
	}

	return "", nil
}
