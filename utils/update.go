package utils

import (
	"reflect"
)

// UpdateStruct maps the fields of a new struct to the fields of an existing struct.
//
// If either struct has a field that the other does not, it will be ignored.
// Any field not tagged with `update:"true"` will be ignored.
//
// Parameters:
//   - current: A pointer to the struct that will be updated.
//   - newStruct: A pointer to the struct that will be used to update the current struct.
//
// Usage:
// If you have an update route, and set your database to also provide json tagging, you can use the same struct for both the database and the update route.
// Specifying `update:"true"` in the json tag will allow the UpdateStruct function to update the struct.
//
// A password would not be set with an update tag as it should be hashed before it is stored in the database.
//
// Another example of this is you have a smaller struct that you want to use to update a larger struct.
// And you only want to update the fields that are in the smaller struct.
//
// Example struct:
//
//	type Data struct {
//	 ID	   int
//	 Name  string `update:"true"
//	}
//
// Returns: None, see the current struct for the updated values.
//
// Note: This function is generic and can be used with any struct type.
func UpdateStruct[t interface{}, t2 interface{}](current *t, newStruct *t2) {
	currentValue := reflect.ValueOf(current).Elem()
	updatesValue := reflect.ValueOf(newStruct).Elem()

	currentType := currentValue.Type()

	for i := 0; i < currentValue.NumField(); i++ {
		currentField := currentValue.Field(i)
		currentFieldInfo := currentType.Field(i)
		currentFieldName := currentFieldInfo.Name

		// Check if the field has the update tag
		if currentFieldInfo.Tag.Get("update") != "true" {
			continue
		}

		// Find the corresponding field in newStruct
		updatesField := updatesValue.FieldByName(currentFieldName)

		if !updatesField.IsValid() || currentField.Type() != updatesField.Type() || updatesField.IsZero() {
			continue
		}

		currentField.Set(updatesField)
	}
}
