package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

// BindRequest binds query parameters, form data, and JSON body to a struct.
//
// Parameters:
//   - r: The HTTP request to bind data from.
//   - dest: A pointer to the struct to bind data to.
//
// Returns: An error if the binding fails.
//
// Usage:
//
//		When binding query parameters and form data, you can use struct tags to specify the field names.
//		The `query` tag specifies the query parameter name, the `form` tag specifies the form field name
//		and the `required` tag specifies if the field is required.
//
//		When binding JSON body, the struct tags are not required. The JSON body is automatically decoded into the struct.
//	 Although specify for consistency.
//
// Example:
//
//	type Request struct {
//	 Field1     string  `query:"field1" form:"field1" json:"field1" required:"true"`
//	 Field2     string  `query:"field2" form:"field2" json:"field2"`
//	}
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	 var req Request
//	 if err := BindRequest(r, &req); err != nil {
//	  http.Error(w, err.Error(), http.StatusBadRequest)
//	  return
//	 }
//	}
//
// Note: This function only supports binding to string, int, uint, float, and bool fields.
// It does not support nested structs or slices. It also does not support binding to unexported fields.
//
// JSON body is only decoded if the Content-Type header is "application/json",
// it will still allow query parameters to be collected.
//
// If JSON data is intended for collection, query parameters may overwrite JSON values.
func BindRequest[T any](r *http.Request, dest *T) error {
	if r.Header.Get("Content-Type") == "application/json" {
		err := decodeJSON(r, dest)
		if err != nil {
			return err
		}

		// Query params may still be present in the URL, so parse them
	}

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	destVal := reflect.ValueOf(dest).Elem()
	destType := destVal.Type()

	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)
		fieldVal := destVal.Field(i)

		queryTag := field.Tag.Get("query")
		formTag := field.Tag.Get("form")
		required := field.Tag.Get("required") == "true"

		if err := bindField(r, fieldVal, queryTag, formTag); err != nil {
			return err
		}

		if required && fieldVal.IsZero() {
			return fmt.Errorf("required field %s is missing", field.Name)
		}
	}

	return nil
}

// decodeJSON is a helper function for BindRequest that decodes JSON data into a struct.
//
// Returns: An error if the JSON decoding fails.
//
// Note: This function is not intended to be used directly, use BindRequest instead.
func decodeJSON[T any](r *http.Request, dest *T) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("failed to decode json: %w", err)
	}
	return nil
}

// bindField tries to set a field from query or form data.
//
// Returns: An error if the field cannot be set.
//
// Note: This function is not intended to be used directly, use BindRequest instead.
func bindField(r *http.Request, fieldVal reflect.Value, queryTag string, formTag string) error {
	if queryTag != "" {
		if val := r.URL.Query().Get(queryTag); val != "" {
			return setFieldValue(fieldVal, val)
		}
	}

	if formTag != "" {
		if val := r.FormValue(formTag); val != "" {
			return setFieldValue(fieldVal, val)
		}
	}

	return nil
}

// setFieldValue sets a field value with reflection, converting string values to the appropriate field type.
//
// Returns: An error if the field value cannot be set, or if the string value cannot be converted to the field type.
//
// Note: This function is not intended to be used directly, use BindRequest instead.
func setFieldValue(field reflect.Value, value string) error {
	if !field.CanSet() {
		return fmt.Errorf("field is not settable")
	}

	var err error
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var intVal int64
		intVal, err = strconv.ParseInt(value, 10, 64)
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var uintVal uint64
		uintVal, err = strconv.ParseUint(value, 10, 64)
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		var floatVal float64
		floatVal, err = strconv.ParseFloat(value, 64)
		field.SetFloat(floatVal)
	case reflect.Bool:
		var boolVal bool
		boolVal, err = strconv.ParseBool(value)
		field.SetBool(boolVal)
	}

	if err != nil {
		return fmt.Errorf("failed to set field value: %w", err)
	}

	return nil
}
